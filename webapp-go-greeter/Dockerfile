# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o webapp .

# Runtime stage
FROM alpine:latest

RUN adduser -D -u 10001 appuser
WORKDIR /app
COPY --from=builder /build/webapp /app/webapp

EXPOSE 8080
USER appuser
ENTRYPOINT ["/app/webapp"]
