FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o ai-chat-service .

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app/ai-chat-service /ai-chat-service
USER nonroot:nonroot
ENTRYPOINT ["/ai-chat-service"]
