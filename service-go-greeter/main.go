// Copyright 2025 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0

// A small HTTP greeter service.
//
// Besides the greeting itself it exposes liveness and readiness endpoints for
// health probes, writes JSON logs carrying the trace and span id of the request
// that produced them, and exports OpenTelemetry traces. A caller that
// propagates trace context therefore shares one trace with this service. It
// reports its version, so builds of different branches can be told apart.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// version identifies this build. The release-lines-next branch, which backs the
// OpenChoreo release lines sample, sets it to v1.1.
const version = "v1.1"

var logger *slog.Logger

func main() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel()}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdownTracing := initTracing(ctx)
	defer shutdownTracing()

	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/greeter/greet", greet)
	serverMux.HandleFunc("/greeter/version", reportVersion)
	// Liveness: is the process up. It must not depend on anything else, or a
	// dependency's outage would restart healthy replicas.
	serverMux.HandleFunc("/healthz", health("ok"))
	// Readiness: is this instance able to serve. This service has no
	// dependencies, so it is ready as soon as it is listening.
	serverMux.HandleFunc("/readyz", health("ready"))

	serverPort := envInt("PORT", 9090)
	server := http.Server{
		Addr:              fmt.Sprintf(":%d", serverPort),
		Handler:           otelhttp.NewHandler(serverMux, "go-greeter"),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("starting HTTP greeter", "port", serverPort, "version", version)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP ListenAndServe", "err", err)
			os.Exit(1)
		}
		logger.Info("HTTP server stopped serving new requests")
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("shutting down the server")
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP shutdown", "err", err)
		os.Exit(1)
	}
	logger.Info("shutdown complete")
}

func greet(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	name := r.URL.Query().Get("name")
	defaulted := name == ""
	if defaulted {
		name = "Stranger"
	}

	span := trace.SpanFromContext(r.Context())
	span.SetAttributes(
		attribute.String("greet.name", name),
		attribute.Bool("greet.name_defaulted", defaulted),
	)

	fmt.Fprintf(w, "Hello, %s!\n", name)

	logFor(r.Context()).Info("greeted",
		"name", name,
		"name_defaulted", defaulted,
		"elapsed_ms", time.Since(started).Milliseconds(),
		"remote", r.RemoteAddr,
		"version", version,
	)
}

func reportVersion(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, version)

	logFor(r.Context()).Info("reported version",
		"remote", r.RemoteAddr,
		"version", version,
	)
}

func health(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, body)
	}
}

// logFor returns a logger carrying the trace and span id, so a log line can be
// matched to the trace it belongs to.
func logFor(ctx context.Context) *slog.Logger {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return logger
	}
	return logger.With("trace_id", sc.TraceID().String(), "span_id", sc.SpanID().String())
}

// initTracing wires the OTLP exporter. Tracing is best-effort: with no endpoint
// configured the service runs normally and exports nothing.
func initTracing(ctx context.Context) func() {
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" && os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") == "" {
		logger.Info("no OTLP endpoint set, traces disabled")
		return func() {}
	}

	exp, err := otlptracehttp.New(ctx)
	if err != nil {
		logger.Warn("otlp exporter unavailable, continuing without traces", "err", err)
		return func() {}
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes("",
			attribute.String("service.name", env("OTEL_SERVICE_NAME", "go-greeter")),
			attribute.String("service.version", version),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))
	logger.Info("traces enabled", "endpoint", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))

	return func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tp.Shutdown(c)
	}
}

func logLevel() slog.Level {
	switch os.Getenv("LOG_LEVEL") {
	case "debug", "DEBUG":
		return slog.LevelDebug
	case "warn", "WARN":
		return slog.LevelWarn
	case "error", "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	var n int
	if v := os.Getenv(k); v != "" {
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n > 0 {
			return n
		}
	}
	return def
}
