package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/attribute"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type NoteStats struct {
	NoteID      string   `json:"note_id"`
	ViewCount   int64    `json:"view_count"`
	RecentViews []string `json:"recent_views"`
}

func handleNoteStats(store *Store, cache *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "handler.NoteStats")
		defer span.End()
		applyDelay(ctx, StageHandler)

		id := strings.TrimPrefix(r.URL.Path, "/api/analytics/")
		if id == "" || id == "top" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "note id required"})
			return
		}
		span.SetAttributes(attribute.String("note.id", id))

		if shouldFail(ctx) {
			span.SetAttributes(attribute.Bool("fault.injected", true))
			logger(ctx).Warn("fault injected", "path", r.URL.Path, "fail_rate", optsFrom(ctx).failRate)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "injected failure"})
			return
		}

		count, recent, hit := cache.GetStats(ctx, id)
		span.SetAttributes(attribute.Bool("cache.hit", hit))
		if !hit {
			c, err := store.GetViewCount(ctx, id)
			if err != nil {
				if err == sql.ErrNoRows {
					writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not found"})
					return
				}
				logger(ctx).Error("get view count failed", "id", id, "error", err)
				writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal error"})
				return
			}
			rs, err := store.RecentViews(ctx, id, 50)
			if err != nil {
				logger(ctx).Warn("recent views failed", "id", id, "error", err)
			}
			count = c
			recent = rs
			cache.SetStats(ctx, id, count, recent)
		}

		if recent == nil {
			recent = []string{}
		}
		writeJSON(w, http.StatusOK, NoteStats{
			NoteID:      id,
			ViewCount:   count,
			RecentViews: recent,
		})
	}
}

func handleTop(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "handler.Top")
		defer span.End()
		applyDelay(ctx, StageHandler)

		if shouldFail(ctx) {
			span.SetAttributes(attribute.Bool("fault.injected", true))
			logger(ctx).Warn("fault injected", "path", r.URL.Path, "fail_rate", optsFrom(ctx).failRate)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "injected failure"})
			return
		}

		rows, err := store.TopNotes(ctx, 10)
		if err != nil {
			logger(ctx).Error("top notes failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal error"})
			return
		}
		if rows == nil {
			rows = []TopNote{}
		}
		writeJSON(w, http.StatusOK, rows)
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
