package main

import (
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
)

// newResource names this service in the traces it exports. OpenChoreo sets the
// component and environment in the pod's environment, so they are attached too
// when present.
func newResource() *resource.Resource {
	attrs := []attribute.KeyValue{
		attribute.String("service.name", env("OTEL_SERVICE_NAME", "web-ui")),
	}
	for _, p := range []struct{ env, key string }{
		{"OPENCHOREO_COMPONENT", "openchoreo.component"},
		{"OPENCHOREO_ENVIRONMENT", "openchoreo.environment"},
		{"OPENCHOREO_PROJECT", "openchoreo.project"},
	} {
		if v := os.Getenv(p.env); v != "" {
			attrs = append(attrs, attribute.String(p.key, v))
		}
	}
	return resource.NewWithAttributes("", attrs...)
}
