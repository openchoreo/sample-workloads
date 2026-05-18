package main

import (
	"log"
	"net/http"
	"os"
)

// NOTE: This service is intentionally broken to demonstrate build-failure
// handling in OpenChoreo. The reference to `undefinedHandler` below does not
// exist, so `go build` will fail with an "undefined" error.

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", undefinedHandler)

	log.Printf("api-service-broken listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
