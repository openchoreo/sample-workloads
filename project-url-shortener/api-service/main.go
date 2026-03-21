package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	shutdown := initTracer("api-service")
	defer shutdown()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/urlshortener?sslmode=disable"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	store := NewStore(dsn)

	cache := NewCache(redisAddr)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/shorten", handleShorten(store, cache))
	mux.HandleFunc("GET /r/", handleRedirect(store, cache))
	mux.HandleFunc("GET /api/urls", handleListURLs(store))
	mux.HandleFunc("DELETE /api/urls/", handleDeleteURL(store, cache))
	mux.HandleFunc("GET /health", handleHealth(store))

	handler := loggingMiddleware(corsMiddleware(tracingMiddleware(mux)))

	log.Printf("API service listening on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Tracing-Enabled")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
