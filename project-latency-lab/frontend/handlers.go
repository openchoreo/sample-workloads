package main

import (
	"io"
	"net/http"
	"os"
	"strings"
)

var httpClient = tracedHTTPClient()

// proxy forwards requests (preserving method, body, headers, AND query
// string) to the chosen upstream. Query string is what carries delay_ms /
// delay_stage / fail_rate, so it must be passed through.
func proxy(targetBase string, label string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "proxy."+label)
		defer span.End()
		r = r.WithContext(ctx)
		targetURL := targetBase + r.URL.Path
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}

		preq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
		if err != nil {
			logger(ctx).Error("build proxy request failed", "label", label, "target", targetURL, "error", err)
			http.Error(w, "proxy error", http.StatusBadGateway)
			return
		}
		for k, vv := range r.Header {
			for _, v := range vv {
				preq.Header.Add(k, v)
			}
		}

		resp, err := httpClient.Do(preq)
		if err != nil {
			logger(ctx).Error("proxy upstream unreachable", "label", label, "target", targetURL, "error", err)
			http.Error(w, "service unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode >= 500 {
			logger(ctx).Error("upstream 5xx", "label", label, "target", targetURL, "status", resp.StatusCode, "body", strings.TrimSpace(string(body)))
		} else if resp.StatusCode >= 400 {
			logger(ctx).Info("upstream 4xx", "label", label, "target", targetURL, "status", resp.StatusCode)
		}
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
		path := r.URL.Path
		if path != "/" && !strings.Contains(path, ".") {
			r.URL.Path = "/"
		}
		fs.ServeHTTP(w, r)
	})
}
