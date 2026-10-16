// web-ui: a small Go web application that renders a greeting page and calls
// greeting-service to produce the greeting itself.
//
// The address of greeting-service is never hardcoded. It arrives in
// GREETING_SERVICE_URL, which OpenChoreo injects from the dependency declared
// on this component's workload, resolved per environment.
package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

//go:embed templates/index.html
var content embed.FS

// Trace context headers are forwarded to the downstream call so the two hops
// join into one trace. Without this the browser -> web-ui -> greeting-service
// chain shows up as two unrelated traces.
var traceHeaders = []string{
	"traceparent", "tracestate", "baggage",
	"x-request-id",
	"x-b3-traceid", "x-b3-spanid", "x-b3-parentspanid", "x-b3-sampled", "x-b3-flags",
}

// web-ui adds its own flourish so the page can show which service produced
// which half of the response.
var flourishes = []string{
	"Hope your deploy is green.",
	"Someone had to say it first.",
	"Delivered across two services.",
	"That greeting travelled further than you'd think.",
	"Freshly resolved, no addresses were hardcoded.",
	"Brought to you by a declared dependency.",
	"Your name went on a short journey.",
	"Two hops, one hello.",
}

type greetResponse struct {
	Greeting  string `json:"greeting"` // from greeting-service
	Flourish  string `json:"flourish"` // from web-ui
	Upstream  string `json:"upstream"` // which address was called
	ElapsedMs int64  `json:"elapsedMs"`
	Error     string `json:"error,omitempty"`
}

type server struct {
	greeterURL  string
	greeterPath string
	client      *http.Client
	page        *template.Template
}

func main() {
	port := env("PORT", "8080")
	s := &server{
		greeterURL:  strings.TrimRight(os.Getenv("GREETING_SERVICE_URL"), "/"),
		greeterPath: env("GREETING_SERVICE_API_PATH", "/greeter/greet"),
		client:      &http.Client{Timeout: 5 * time.Second},
	}

	tmpl, err := template.ParseFS(content, "templates/index.html")
	if err != nil {
		log.Fatalf("parse page: %v", err)
	}
	s.page = tmpl

	if s.greeterURL == "" {
		log.Print("GREETING_SERVICE_URL is not set; the page will load but greetings will fail")
	} else {
		log.Printf("greeting-service resolved to %s", s.greeterURL)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/greet", s.handleGreet)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "ok")
	})
	// Readiness reports on this app only, never on greeting-service. A readiness
	// probe that checked a dependency would take this pod out of service during
	// someone else's outage.
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "ready")
	})

	log.Printf("web-ui listening on :%s", port)
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server: %v", err)
	}
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.page.Execute(w, map[string]string{"Upstream": s.greeterURL}); err != nil {
		log.Printf("render: %v", err)
	}
}

func (s *server) handleGreet(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		name = "there"
	}

	out := greetResponse{Upstream: s.greeterURL, Flourish: flourishes[rand.Intn(len(flourishes))]}
	started := time.Now()

	greeting, err := s.callGreeter(r.Context(), r, name)
	out.ElapsedMs = time.Since(started).Milliseconds()

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Printf("greet %q failed after %dms: %v", name, out.ElapsedMs, err)
		out.Error = err.Error()
		w.WriteHeader(http.StatusBadGateway)
	} else {
		out.Greeting = greeting
	}
	json.NewEncoder(w).Encode(out)
}

func (s *server) callGreeter(ctx context.Context, in *http.Request, name string) (string, error) {
	if s.greeterURL == "" {
		return "", errors.New("GREETING_SERVICE_URL is not set, so there is no address to call")
	}

	target := s.greeterURL + s.greeterPath + "?name=" + url.QueryEscape(name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "text/plain")
	for _, h := range traceHeaders {
		if v := in.Header.Get(h); v != "" {
			req.Header.Set(h, v)
		}
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("greeting-service answered " + resp.Status)
	}
	return strings.TrimSpace(string(body)), nil
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
