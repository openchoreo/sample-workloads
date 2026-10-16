# webapp-go-greeter (`web-ui`)

A small Go web application for OpenChoreo Course 2. It serves a page with a name
field; submitting it calls `greeting-service` and shows which service produced
which half of the answer.

    browser  ->  web-ui (this app)  ->  greeting-service

That second hop is the point. `web-ui` never hardcodes an address for
`greeting-service`; it reads `GREETING_SERVICE_URL`, which OpenChoreo injects
from the dependency declared on this component's workload and resolves per
environment.

## Environment

| Variable | Required | Meaning |
|---|---|---|
| `GREETING_SERVICE_URL` | yes | injected by OpenChoreo from the declared dependency |
| `GREETING_SERVICE_API_PATH` | no | path on greeting-service, default `/greeter/greet` |
| `PORT` | no | listen port, default `8080` |

With `GREETING_SERVICE_URL` unset the page still loads and the greeting fails
with a readable error, which is what you want a learner to see.

## Endpoints

| Path | Purpose |
|---|---|
| `/` | the page |
| `/api/greet?name=` | calls greeting-service, returns JSON |
| `/healthz` | liveness |
| `/readyz` | readiness |

`/readyz` reports on this app only and never checks `greeting-service`. A
readiness probe that checked a dependency would pull this pod out of service
during someone else's outage — the mistake Course 2 lesson 6 warns about.

## Tracing and logs

The app exports OpenTelemetry traces over OTLP and writes JSON logs that carry
the `trace_id` and `span_id` of the request that produced them, so a log line
can be matched to its trace and the other way round.

A span is recorded for the inbound request and for the outbound call to
greeting-service, and trace context is propagated on that call, so
browser -> web-ui -> greeting-service is one trace rather than two unrelated
ones.

Tracing is best-effort. With no OTLP endpoint set it is simply disabled, and if
the collector is unreachable the app logs a warning and serves normally.

| Variable | Meaning |
|---|---|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | the collector, e.g. `http://opentelemetry-collector.openchoreo-observability-plane.svc.cluster.local:4318` |
| `OTEL_SERVICE_NAME` | service name in traces, default `web-ui` |

`OPENCHOREO_COMPONENT`, `OPENCHOREO_ENVIRONMENT` and `OPENCHOREO_PROJECT` are
attached as resource attributes when present.

## Run locally

    GREETING_SERVICE_URL=http://localhost:9090 go run .
