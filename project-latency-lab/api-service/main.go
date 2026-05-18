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
		dsn = "postgres://postgres:postgres@localhost:5432/notes?sslmode=disable"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	authURL := os.Getenv("AUTH_SERVICE_URL")
	if authURL == "" {
		authURL = "http://localhost:8082"
	}

	store := NewStore(dsn)
	cache := NewCache(redisAddr)
	auth := NewAuthClient(authURL)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/notes", handleCreate(store, cache, auth))
	mux.HandleFunc("GET /api/notes", handleList(store))
	mux.HandleFunc("GET /api/notes/", handleGet(store, cache))
	mux.HandleFunc("DELETE /api/notes/", handleDelete(store, cache, auth))
	mux.HandleFunc("GET /health", handleHealth(store))

	handler := loggingMiddleware(corsMiddleware(tracingMiddleware(delayMiddleware(mux))))

	log.Printf("api-service listening on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Auth-Token")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
