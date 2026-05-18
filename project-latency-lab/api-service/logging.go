package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// logger returns a slog.Logger preloaded with the active trace_id / span_id
// so log lines can be correlated to the trace that produced them. Falls
// back to the default logger when there's no span on the context.
func logger(ctx context.Context) *slog.Logger {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return slog.Default()
	}
	return slog.With(
		"trace_id", sc.TraceID().String(),
		"span_id", sc.SpanID().String(),
	)
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)
		logger(r.Context()).Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"status", rw.statusCode,
			"duration_ms", float64(time.Since(start).Microseconds())/1000,
		)
	})
}
