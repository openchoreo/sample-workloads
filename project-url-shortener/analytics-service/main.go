package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	shutdown := initTracer("analytics-service")
	defer shutdown()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
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
	mux.HandleFunc("GET /api/analytics/top", handleGetTopURLs(store))
	mux.HandleFunc("GET /api/analytics/user/", handleGetUserAnalytics(store))
	mux.HandleFunc("GET /api/analytics/", handleGetAnalytics(store, cache))
	mux.HandleFunc("GET /health", handleHealth(store))

	handler := loggingMiddleware(corsMiddleware(tracingMiddleware(mux)))

	log.Printf("Analytics service listening on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
