FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o ai-reading-assistant .

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app/ai-reading-assistant /ai-reading-assistant
USER nonroot:nonroot
ENTRYPOINT ["/ai-reading-assistant"]
