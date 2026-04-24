FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o reading-notes-mcp .

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app/reading-notes-mcp /reading-notes-mcp
USER nonroot:nonroot
ENTRYPOINT ["/reading-notes-mcp"]