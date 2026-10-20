// web-ui: a small Go web application that renders a greeting page and calls
// greeting-service to produce the greeting itself.
//
//	browser -> web-ui -> greeting-service
//
// The address of greeting-service is never hardcoded. It arrives in
// GREETING_SERVICE_URL, which OpenChoreo injects from the dependency declared
// on this component's workload and resolves per environment.
//
// The app emits OpenTelemetry traces and JSON logs that carry the trace and
// span id, so a log line can be tied back to the request that produced it.
package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

//go:embed templates/index.html
var content embed.FS

// A friendly second line to go with the greeting.
var flourishes = []string{
	"Lovely to meet you.",
	"Hope your day is going well.",
	"That has a nice ring to it.",
	"Someone had to say it first.",
	"Consider yourself greeted.",
	"Always good to see a new name.",
	"Here's to a good one.",
	"Glad you stopped by.",
}

type greetResponse struct {
	Greeting  string `json:"greeting"`
	Flourish  string `json:"flourish"`
	Upstream  string `json:"upstream"`
	ElapsedMs int64  `json:"elapsedMs"`
	TraceID   string `json:"traceId,omitempty"`
	Error     string `json:"error,omitempty"`
}

type server struct {
	greeterURL  string
	greeterPath string
	client      *http.Client
	page        *template.Template
	log         *slog.Logger
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownTracing := initTracing(ctx, log)
	defer shutdownTracing()

	tmpl, err := template.ParseFS(content, "templates/index.html")
	if err != nil {
		log.Error("parse page", "err", err)
		os.Exit(1)
	}

	s := &server{
		greeterURL:  strings.TrimRight(os.Getenv("GREETING_SERVICE_URL"), "/"),
		greeterPath: env("GREETING_SERVICE_API_PATH", "/greeter/greet"),
		page:        tmpl,
		log:         log,
		// otelhttp.NewTransport records a span for the outbound call and injects
		// the trace context headers, so this hop joins the caller's trace.
		client: &http.Client{
			Timeout:   5 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}

	if s.greeterURL == "" {
		log.Warn("GREETING_SERVICE_URL is not set; the page will load but greetings will fail")
	} else {
		log.Info("resolved greeting-service", "url", s.greeterURL)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/greet", s.handleGreet)
	mux.HandleFunc("/healthz", ok("ok"))
	// Readiness reports on this app only, never on greeting-service. A readiness
	// probe that checked a dependency would take this pod out of service during
	// someone else's outage.
	mux.HandleFunc("/readyz", ok("ready"))

	port := env("PORT", "8080")
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           otelhttp.NewHandler(mux, "web-ui"),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("web-ui listening", "port", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutCtx)
}

// initTracing wires the OTLP exporter. Tracing is best-effort: if no collector
// is reachable the app runs normally and simply exports nothing.
func initTracing(ctx context.Context, log *slog.Logger) func() {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" && os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") == "" {
		log.Info("no OTLP endpoint set, traces disabled")
		return func() {}
	}

	exp, err := otlptracehttp.New(ctx)
	if err != nil {
		log.Warn("otlp exporter unavailable, continuing without traces", "err", err)
		return func() {}
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(newResource()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))
	log.Info("traces enabled", "endpoint", endpoint)

	return func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tp.Shutdown(c)
	}
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.page.Execute(w, nil); err != nil {
		s.logFor(r.Context()).Error("render page", "err", err)
	}
}

func (s *server) handleGreet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := s.logFor(ctx)

	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		name = "there"
	}

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("greet.name", name))

	out := greetResponse{
		Upstream: s.greeterURL,
		Flourish: flourishes[rand.Intn(len(flourishes))],
		TraceID:  traceID(ctx),
	}

	started := time.Now()
	greeting, err := s.callGreeter(ctx, name)
	out.ElapsedMs = time.Since(started).Milliseconds()

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Error("greet failed", "name", name, "elapsed_ms", out.ElapsedMs, "upstream", s.greeterURL, "err", err)
		span.RecordError(err)
		out.Error = err.Error()
		w.WriteHeader(http.StatusBadGateway)
	} else {
		log.Info("greet ok", "name", name, "elapsed_ms", out.ElapsedMs)
		out.Greeting = greeting
	}
	_ = json.NewEncoder(w).Encode(out)
}

func (s *server) callGreeter(ctx context.Context, name string) (string, error) {
	if s.greeterURL == "" {
		return "", errors.New("GREETING_SERVICE_URL is not set, so there is no address to call")
	}

	target := s.greeterURL + s.greeterPath + "?name=" + url.QueryEscape(name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "text/plain")

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

// logFor returns a logger that carries the trace and span id, so a log line can
// be matched to the trace it came from.
func (s *server) logFor(ctx context.Context) *slog.Logger {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return s.log
	}
	return s.log.With("trace_id", sc.TraceID().String(), "span_id", sc.SpanID().String())
}

func traceID(ctx context.Context) string {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		return sc.TraceID().String()
	}
	return ""
}

func ok(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, body)
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
