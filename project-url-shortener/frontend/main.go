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

	mux := http.NewServeMux()

	// Proxy routes - order matters: more specific first
	mux.HandleFunc("/api/shorten", proxyHandler(apiURL, "shorten"))
	mux.HandleFunc("/api/urls/", proxyHandler(apiURL, "urls"))
	mux.HandleFunc("/api/urls", proxyHandler(apiURL, "urls"))
	mux.HandleFunc("/api/analytics/", proxyHandler(analyticsURL, "analytics"))
	mux.HandleFunc("/r/", proxyHandler(apiURL, "redirect"))

	// Static files (SPA fallback)
	mux.Handle("/", staticHandler())

	handler := loggingMiddleware(tracingMiddleware(mux))

	log.Printf("Frontend BFF listening on :%s", port)
	log.Printf("  API proxy -> %s", apiURL)
	log.Printf("  Analytics proxy -> %s", analyticsURL)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
