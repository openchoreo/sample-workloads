package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

type ShortenRequest struct {
	URL        string `json:"url"`
	Username   string `json:"username"`
	CustomSlug string `json:"custom_slug,omitempty"`
}

type ShortenResponse struct {
	ShortCode string `json:"short_code"`
	URL       *URL   `json:"url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func handleShorten(store *Store, cache *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, span := tracer.Start(ctx, "handler.Shorten")
		defer span.End()

		var req ShortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
			return
		}

		if req.URL == "" || req.Username == "" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "url and username are required"})
			return
		}

		// Auto-prepend https:// if no scheme provided
		if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
			req.URL = "https://" + req.URL
		}

		// Validate URL
		parsed, err := url.ParseRequestURI(req.URL)
		if err != nil || parsed.Host == "" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid URL"})
			return
		}

		// Generate or validate slug
		shortCode := req.CustomSlug
		if shortCode != "" {
			if len(shortCode) < 3 || len(shortCode) > 20 {
				writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "custom slug must be 3-20 characters"})
				return
			}
			exists, err := store.ShortCodeExists(ctx, shortCode)
			if err != nil {
				slog.Error("failed to check slug availability in postgres", "short_code", shortCode, "error", err)
				writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal error"})
				return
			}
			if exists {
				writeJSON(w, http.StatusConflict, ErrorResponse{Error: "slug already taken"})
				return
			}
		} else {
			shortCode = generateShortCode()
		}

		span.SetAttributes(attribute.String("short_code", shortCode))

		u, err := store.InsertURL(ctx, shortCode, req.URL, req.Username)
		if err != nil {
			slog.Error("failed to insert URL into postgres", "short_code", shortCode, "url", req.URL, "error", err)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to create short URL"})
			return
		}

		// Cache the URL
		cache.SetURL(ctx, shortCode, req.URL)

		// Async: fetch metadata (detach from request cancellation, keep trace context)
		go fetchMetadata(context.WithoutCancel(ctx), shortCode, req.URL, store)

		writeJSON(w, http.StatusCreated, ShortenResponse{
			ShortCode: shortCode,
			URL:       u,
		})
	}
}

func handleRedirect(store *Store, cache *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, span := tracer.Start(ctx, "handler.Redirect")
		defer span.End()

		code := strings.TrimPrefix(r.URL.Path, "/r/")
		if code == "" {
			http.NotFound(w, r)
			return
		}
		span.SetAttributes(attribute.String("short_code", code))

		// Try cache first
		originalURL, err := cache.GetURL(ctx, code)
		if err != nil {
			// Cache miss - hit postgres
			u, err := store.GetURL(ctx, code)
			if err != nil {
				if err == sql.ErrNoRows {
					http.NotFound(w, r)
					return
				}
				slog.Error("failed to get URL from postgres for redirect", "short_code", code, "error", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			originalURL = u.OriginalURL
			cache.SetURL(ctx, code, originalURL)
		}

		// Async: record click (detach from request cancellation, keep trace context)
		asyncCtx := context.WithoutCancel(ctx)
		go func() {
			if err := store.RecordClick(asyncCtx, code); err != nil {
				slog.Warn("failed to record click in postgres", "short_code", code, "error", err)
			}
			cache.IncrClickCount(asyncCtx, code)
		}()

		http.Redirect(w, r, originalURL, http.StatusFound)
	}
}

func handleListURLs(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, span := tracer.Start(ctx, "handler.ListURLs")
		defer span.End()

		username := r.URL.Query().Get("username")
		if username == "" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "username query parameter required"})
			return
		}

		urls, err := store.ListURLs(ctx, username)
		if err != nil {
			slog.Error("failed to list URLs from postgres", "username", username, "error", err)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal error"})
			return
		}

		if urls == nil {
			urls = []URL{}
		}
		writeJSON(w, http.StatusOK, urls)
	}
}

func handleDeleteURL(store *Store, cache *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, span := tracer.Start(ctx, "handler.DeleteURL")
		defer span.End()

		code := strings.TrimPrefix(r.URL.Path, "/api/urls/")
		if code == "" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "short code required"})
			return
		}

		if err := store.DeleteURL(ctx, code); err != nil {
			if err == sql.ErrNoRows {
				writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not found"})
				return
			}
			slog.Error("failed to delete URL from postgres", "short_code", code, "error", err)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal error"})
			return
		}

		cache.DeleteURL(ctx, code)
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

func generateShortCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}
