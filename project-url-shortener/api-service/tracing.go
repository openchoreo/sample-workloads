package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

func initTracer(serviceName string) func() {
	enabled := os.Getenv("OTEL_ENABLED")
	if strings.EqualFold(enabled, "false") {
		tracer = otel.Tracer(serviceName)
		return func() {}
	}

	endpoint := os.Getenv("OTEL_EXPORTER_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:4318"
	}
	// Strip protocol for the OTLP client
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	ctx := context.Background()

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		slog.Warn("failed to create otel trace exporter", "error", err)
		tracer = otel.Tracer(serviceName)
		return func() {}
	}

	res, _ := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceNameKey.String(serviceName)),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer = tp.Tracer(serviceName)
	slog.Info("otel tracing initialized", "service", serviceName, "endpoint", endpoint)

	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shut down otel tracer", "error", err)
		}
	}
}

func tracingMiddleware(next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, "http.request")
}

func tracedHTTPClient() *http.Client {
	return &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}
