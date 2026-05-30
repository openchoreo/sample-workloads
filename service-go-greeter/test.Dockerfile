FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o go-greeter .

# Docker Hub pulls
FROM busybox:1.36 AS test-busybox
RUN echo "busybox" > /registry-proof

FROM alpine:3.20 AS test-alpine
RUN echo "alpine" > /tmp/registry-proof

FROM nginx:1.27-alpine AS test-nginx
RUN echo "nginx" > /tmp/registry-proof

FROM python:3.12-alpine AS test-python
RUN echo "python" > /tmp/registry-proof

FROM node:22-alpine AS test-node
RUN echo "node" > /tmp/registry-proof

# GCR pull
FROM gcr.io/distroless/static-debian12:latest AS test-gcr

# GHCR pull
FROM ghcr.io/actions/actions-runner:2.321.0 AS test-ghcr
RUN echo "ghcr" > /tmp/registry-proof

# Quay pull
FROM quay.io/podman/stable:latest AS test-quay
RUN echo "quay" > /tmp/registry-proof

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/go-greeter .
COPY --from=test-busybox /registry-proof /tmp/proof-busybox
COPY --from=test-alpine /tmp/registry-proof /tmp/proof-alpine
COPY --from=test-nginx /tmp/registry-proof /tmp/proof-nginx
COPY --from=test-python /tmp/registry-proof /tmp/proof-python
COPY --from=test-node /tmp/registry-proof /tmp/proof-node
COPY --from=test-gcr /etc/os-release /tmp/proof-gcr
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
