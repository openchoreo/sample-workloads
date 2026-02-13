package document

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Server struct {
	store *Store
}

type CreateDocumentRequest struct {
	DisplayName string `json:"displayName"`
}

type UpdateTitleRequest struct {
	DisplayName string `json:"displayName"`
}

type DocumentResponse struct {
	DocumentID  string `json:"document_id"`
	DisplayName string `json:"displayName"`
	Content     string `json:"content"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type DocumentListItem struct {
	DocumentID  string `json:"document_id"`
	DisplayName string `json:"displayName"`
	UpdatedAt   string `json:"updated_at"`
}

func NewServer(store *Store) *Server {
	return &Server{store: store}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(logRequests)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type"},
	}))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Route("/documents", func(r chi.Router) {
		r.Post("/", s.handleCreateDocument)
		r.Get("/", s.handleListDocuments)
		r.Get("/{document_id}", s.handleGetDocument)
		r.Put("/{document_id}/title", s.handleUpdateTitle)
		r.Delete("/{document_id}", s.handleDeleteDocument)
	})

	return r
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		log.Printf("document %s %s %d %s", r.Method, r.URL.Path, ww.Status(), time.Since(start))
	})
}

func (s *Server) handleCreateDocument(w http.ResponseWriter, r *http.Request) {
	var req CreateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	doc, err := s.store.CreateDocument(r.Context(), req.DisplayName)
	if err != nil {
		log.Printf("create document error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create_failed"})
		return
	}

	writeJSON(w, http.StatusCreated, documentToResponse(doc))
}

func (s *Server) handleGetDocument(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "document_id")
	docID, err := uuid.Parse(idParam)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_document_id"})
		return
	}
	doc, err := s.store.GetDocument(r.Context(), docID)
	if err != nil {
		if IsNotFound(err) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
			return
		}
		log.Printf("get document error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "fetch_failed"})
		return
	}

	writeJSON(w, http.StatusOK, documentToResponse(doc))
}

func (s *Server) handleListDocuments(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	limit := parseInt(r.URL.Query().Get("limit"), 50)
	offset := parseInt(r.URL.Query().Get("offset"), 0)

	docs, err := s.store.ListDocuments(r.Context(), query, limit, offset)
	if err != nil {
		log.Printf("list documents error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list_failed"})
		return
	}

	items := make([]DocumentListItem, 0, len(docs))
	for _, doc := range docs {
		items = append(items, DocumentListItem{
			DocumentID:  doc.DocumentID.String(),
			DisplayName: doc.DisplayName,
			UpdatedAt:   doc.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
	})
}

func (s *Server) handleUpdateTitle(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "document_id")
	docID, err := uuid.Parse(idParam)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_document_id"})
		return
	}

	var req UpdateTitleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	if req.DisplayName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "display_name_required"})
		return
	}

	if err := s.store.UpdateTitle(r.Context(), docID, req.DisplayName); err != nil {
		if IsNotFound(err) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
			return
		}
		log.Printf("update title error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "update_failed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleDeleteDocument(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "document_id")
	docID, err := uuid.Parse(idParam)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_document_id"})
		return
	}

	if err := s.store.DeleteDocument(r.Context(), docID); err != nil {
		if IsNotFound(err) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
			return
		}
		log.Printf("delete document error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "delete_failed"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func documentToResponse(doc Document) DocumentResponse {
	return DocumentResponse{
		DocumentID:  doc.DocumentID.String(),
		DisplayName: doc.DisplayName,
		Content:     base64.StdEncoding.EncodeToString(doc.Content),
		CreatedAt:   doc.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   doc.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json error: %v", err)
	}
}

func parseInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func OpenDatabase(dsn string) (*gorm.DB, error) {
	return gorm.Open(openPostgres(dsn), &gorm.Config{})
}
