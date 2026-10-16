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

## Tracing

Incoming `traceparent`, `tracestate`, `baggage`, `x-request-id` and B3 headers
are copied onto the downstream request, so browser -> web-ui -> greeting-service
joins into a single trace rather than two unrelated ones.

## Run locally

    GREETING_SERVICE_URL=http://localhost:9090 go run .
