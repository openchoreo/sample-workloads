package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

var httpClient = newProxyClient()

func proxyHandler(targetBase string, pathPrefix string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, span := tracer.Start(ctx, "proxy."+pathPrefix)
		defer span.End()

		targetURL := targetBase + r.URL.Path
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}

		proxyReq, err := http.NewRequestWithContext(ctx, r.Method, targetURL, r.Body)
		if err != nil {
			slog.Error("failed to create proxy request", "target", targetURL, "error", err)
			http.Error(w, "proxy error", http.StatusBadGateway)
			return
		}

		// Copy headers
		for k, vv := range r.Header {
			for _, v := range vv {
				proxyReq.Header.Add(k, v)
			}
		}

		resp, err := httpClient.Do(proxyReq)
		if err != nil {
			slog.Error("failed to proxy request", "target", targetURL, "error", err)
			http.Error(w, "service unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode >= 500 {
			slog.Error("upstream error",
				"target", targetURL,
				"status", resp.StatusCode,
				"body", strings.TrimSpace(string(body)),
			)
		}

		// Copy response headers
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
	}
}

func staticHandler() http.Handler {
	dir := os.Getenv("STATIC_DIR")
	if dir == "" {
		dir = "static"
	}
	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve index.html for SPA routes (non-file paths)
		path := r.URL.Path
		if path != "/" && !strings.Contains(path, ".") {
			r.URL.Path = "/"
		}
		fs.ServeHTTP(w, r)
	})
}
