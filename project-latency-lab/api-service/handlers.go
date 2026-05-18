package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// authenticate resolves X-Auth-Token via the auth-service. On failure it
// writes the response and returns ok=false. The latency-knob query string
// is forwarded so injection works end-to-end across services.
func authenticate(ctx context.Context, w http.ResponseWriter, r *http.Request, auth *AuthClient, span trace.Span) (string, bool) {
	log := logger(ctx)
	token := r.Header.Get("X-Auth-Token")
	if token == "" {
		log.Info("auth rejected: missing X-Auth-Token", "path", r.URL.Path)
		span.SetAttributes(attribute.String("auth.outcome", "missing_token"))
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "missing X-Auth-Token"})
		return "", false
	}
	username, err := auth.Verify(ctx, token, r.URL.RawQuery)
	if err != nil {
		if errors.Is(err, ErrUnauthorized) {
			log.Info("auth rejected: invalid token", "path", r.URL.Path)
			span.SetAttributes(attribute.String("auth.outcome", "invalid_token"))
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
			return "", false
		}
		log.Error("auth verify failed", "error", err, "path", r.URL.Path)
		span.SetAttributes(attribute.String("auth.outcome", "backend_error"), attribute.String("auth.error", err.Error()))
		writeJSON(w, http.StatusBadGateway, ErrorResponse{Error: "auth unavailable"})
		return "", false
	}
	log.Debug("auth ok", "username", username)
	return username, true
}

type CreateRequest struct {
	Body string `json:"body"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func handleCreate(store *Store, cache *Cache, auth *AuthClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "handler.Create")
		defer span.End()
		applyDelay(ctx, StageHandler)

		author, ok := authenticate(ctx, w, r, auth, span)
		if !ok {
			return
		}

		var req CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
			return
		}
		if req.Body == "" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "body is required"})
			return
		}
		if len(req.Body) > 10_000 {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "body too long (max 10000 chars)"})
			return
		}

		if shouldFail(ctx) {
			span.SetAttributes(attribute.Bool("fault.injected", true))
			logger(ctx).Warn("fault injected", "path", r.URL.Path, "fail_rate", optsFrom(ctx).failRate)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "injected failure"})
			return
		}

		id := generateID()
		span.SetAttributes(attribute.String("note.id", id), attribute.String("author", author))

		n, err := store.InsertNote(ctx, id, author, req.Body)
		if err != nil {
			logger(ctx).Error("insert note failed", "id", id, "author", author, "error", err)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to create note"})
			return
		}
		cache.SetNote(ctx, n)

		writeJSON(w, http.StatusCreated, n)
	}
}

func handleGet(store *Store, cache *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "handler.Get")
		defer span.End()
		applyDelay(ctx, StageHandler)

		id := strings.TrimPrefix(r.URL.Path, "/api/notes/")
		if id == "" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "id required"})
			return
		}
		span.SetAttributes(attribute.String("note.id", id))

		if shouldFail(ctx) {
			span.SetAttributes(attribute.Bool("fault.injected", true))
			logger(ctx).Warn("fault injected", "path", r.URL.Path, "fail_rate", optsFrom(ctx).failRate)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "injected failure"})
			return
		}

		n, err := cache.GetNote(ctx, id)
		if err != nil {
			n, err = store.GetNote(ctx, id)
			if err != nil {
				if err == sql.ErrNoRows {
					writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not found"})
					return
				}
				logger(ctx).Error("get note failed", "id", id, "error", err)
				writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal error"})
				return
			}
			cache.SetNote(ctx, n)
		}

		if err := store.RecordView(ctx, id); err != nil {
			logger(ctx).Warn("record view failed", "id", id, "error", err)
		}
		cache.IncrViewCount(ctx, id)

		writeJSON(w, http.StatusOK, n)
	}
}

func handleList(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "handler.List")
		defer span.End()
		applyDelay(ctx, StageHandler)

		if shouldFail(ctx) {
			span.SetAttributes(attribute.Bool("fault.injected", true))
			logger(ctx).Warn("fault injected", "path", r.URL.Path, "fail_rate", optsFrom(ctx).failRate)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "injected failure"})
			return
		}

		author := r.URL.Query().Get("author")
		span.SetAttributes(attribute.String("author", author))
		notes, err := store.ListNotes(ctx, author)
		if err != nil {
			logger(ctx).Error("list notes failed", "author", author, "error", err)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal error"})
			return
		}
		if notes == nil {
			notes = []Note{}
		}
		writeJSON(w, http.StatusOK, notes)
	}
}

func handleDelete(store *Store, cache *Cache, auth *AuthClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "handler.Delete")
		defer span.End()
		applyDelay(ctx, StageHandler)

		author, ok := authenticate(ctx, w, r, auth, span)
		if !ok {
			return
		}

		id := strings.TrimPrefix(r.URL.Path, "/api/notes/")
		if id == "" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "id required"})
			return
		}
		span.SetAttributes(attribute.String("note.id", id), attribute.String("author", author))

		if shouldFail(ctx) {
			span.SetAttributes(attribute.Bool("fault.injected", true))
			logger(ctx).Warn("fault injected", "path", r.URL.Path, "fail_rate", optsFrom(ctx).failRate)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "injected failure"})
			return
		}

		if err := store.DeleteNote(ctx, id); err != nil {
			if err == sql.ErrNoRows {
				writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not found"})
				return
			}
			logger(ctx).Error("delete note failed", "id", id, "author", author, "error", err)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal error"})
			return
		}
		cache.DeleteNote(ctx, id)
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleHealth(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := store.db.Ping(); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unhealthy", "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}
