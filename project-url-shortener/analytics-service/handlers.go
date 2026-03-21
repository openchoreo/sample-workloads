package main

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

type AnalyticsResponse struct {
	ShortCode    string         `json:"short_code"`
	OriginalURL  string         `json:"original_url"`
	Title        string         `json:"title"`
	FaviconURL   string         `json:"favicon_url"`
	ClickCount   int64          `json:"click_count"`
	CreatedAt    time.Time      `json:"created_at"`
	RecentClicks []ClickRecord  `json:"recent_clicks"`
}

type UserAnalyticsResponse struct {
	Username   string         `json:"username"`
	TotalURLs  int            `json:"total_urls"`
	TotalClicks int64         `json:"total_clicks"`
	URLs       []URLAnalytics `json:"urls"`
}

func handleGetAnalytics(store *Store, cache *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, span := tracer.Start(ctx, "handler.GetAnalytics")
		defer span.End()

		code := strings.TrimPrefix(r.URL.Path, "/api/analytics/")
		if code == "" || strings.Contains(code, "/") {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "short code required"})
			return
		}
		span.SetAttributes(attribute.String("short_code", code))

		// Get URL info from postgres
		urlAnalytics, err := store.GetURLAnalytics(ctx, code)
		if err != nil {
			if err == sql.ErrNoRows {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
				return
			}
			slog.Error("failed to get URL analytics from postgres", "short_code", code, "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}

		// Try cache for click data
		clickCount := urlAnalytics.ClickCount
		var recentClicks []ClickRecord

		cachedCount, err := cache.GetClickCount(ctx, code)
		if err == nil {
			clickCount = cachedCount
		}

		cachedRecent, err := cache.GetRecentClicks(ctx, code)
		if err == nil {
			for _, ts := range cachedRecent {
				t, _ := time.Parse(time.RFC3339, ts)
				recentClicks = append(recentClicks, ClickRecord{ClickedAt: t})
			}
		} else {
			// Fallback to postgres
			recentClicks, err = store.GetRecentClicks(ctx, code, 50)
			if err != nil {
				slog.Warn("failed to get recent clicks from postgres", "short_code", code, "error", err)
				recentClicks = []ClickRecord{}
			}
			// Cache the result
			if len(recentClicks) > 0 {
				timestamps := make([]string, len(recentClicks))
				for i, c := range recentClicks {
					timestamps[i] = c.ClickedAt.UTC().Format(time.RFC3339)
				}
				cache.SetClickData(ctx, code, clickCount, timestamps)
			}
		}

		if recentClicks == nil {
			recentClicks = []ClickRecord{}
		}

		writeJSON(w, http.StatusOK, AnalyticsResponse{
			ShortCode:    urlAnalytics.ShortCode,
			OriginalURL:  urlAnalytics.OriginalURL,
			Title:        urlAnalytics.Title,
			FaviconURL:   urlAnalytics.FaviconURL,
			ClickCount:   clickCount,
			CreatedAt:    urlAnalytics.CreatedAt,
			RecentClicks: recentClicks,
		})
	}
}

func handleGetUserAnalytics(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, span := tracer.Start(ctx, "handler.GetUserAnalytics")
		defer span.End()

		username := strings.TrimPrefix(r.URL.Path, "/api/analytics/user/")
		if username == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username required"})
			return
		}
		span.SetAttributes(attribute.String("username", username))

		urls, err := store.GetUserAnalytics(ctx, username)
		if err != nil {
			slog.Error("failed to get user analytics from postgres", "username", username, "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}

		if urls == nil {
			urls = []URLAnalytics{}
		}

		var totalClicks int64
		for _, u := range urls {
			totalClicks += u.ClickCount
		}

		writeJSON(w, http.StatusOK, UserAnalyticsResponse{
			Username:    username,
			TotalURLs:   len(urls),
			TotalClicks: totalClicks,
			URLs:        urls,
		})
	}
}

func handleGetTopURLs(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, span := tracer.Start(ctx, "handler.GetTopURLs")
		defer span.End()

		urls, err := store.GetTopURLs(ctx, 50)
		if err != nil {
			slog.Error("failed to get top URLs from postgres", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		if urls == nil {
			urls = []URLAnalytics{}
		}

		var totalClicks int64
		for _, u := range urls {
			totalClicks += u.ClickCount
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"total_urls":   len(urls),
			"total_clicks": totalClicks,
			"urls":         urls,
		})
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
