package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	shutdown := initTracer("frontend")
	defer shutdown()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	apiURL := os.Getenv("API_SERVICE_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}
	analyticsURL := os.Getenv("ANALYTICS_SERVICE_URL")
	if analyticsURL == "" {
		analyticsURL = "http://localhost:8081"
	}
	authURL := os.Getenv("AUTH_SERVICE_URL")
	if authURL == "" {
		authURL = "http://localhost:8082"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/notes", proxy(apiURL, "notes"))
	mux.HandleFunc("/api/notes/", proxy(apiURL, "notes"))
	mux.HandleFunc("/api/analytics/", proxy(analyticsURL, "analytics"))
	mux.HandleFunc("/api/auth/", proxy(authURL, "auth"))
	mux.Handle("/", staticHandler())

	handler := loggingMiddleware(tracingMiddleware(mux))

	log.Printf("frontend listening on :%s", port)
	log.Printf("  api  -> %s", apiURL)
	log.Printf("  anl  -> %s", analyticsURL)
	log.Printf("  auth -> %s", authURL)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
