FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/document ./cmd

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /out/document /app/document
EXPOSE 8080
CMD ["/app/document"]
