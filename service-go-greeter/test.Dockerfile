FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o go-greeter .

# Docker Hub pulls
FROM busybox:1.36 AS test-busybox
RUN echo "busybox" > /registry-proof

FROM alpine:3.20 AS test-alpine
RUN echo "alpine" > /tmp/registry-proof

# GCR pull (~1 MB)
FROM gcr.io/google-containers/pause:3.9 AS test-gcr
RUN echo "gcr" > /tmp/registry-proof

# GHCR pull (~4 MB)
FROM ghcr.io/jqlang/jq:latest AS test-ghcr
RUN echo "ghcr" > /tmp/registry-proof

# Quay pull (~4 MB)
FROM quay.io/prometheus/busybox:latest AS test-quay
RUN echo "quay" > /tmp/registry-proof

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/go-greeter .
COPY --from=test-busybox /registry-proof /tmp/proof-busybox
COPY --from=test-alpine /tmp/registry-proof /tmp/proof-alpine
COPY --from=test-gcr /tmp/registry-proof /tmp/proof-gcr
COPY --from=test-ghcr /tmp/registry-proof /tmp/proof-ghcr
COPY --from=test-quay /tmp/registry-proof /tmp/proof-quay

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid 10014 \
    "choreo"
USER 10014

ENTRYPOINT ["./go-greeter"]
