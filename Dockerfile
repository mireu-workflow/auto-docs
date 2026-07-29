# Multi-stage Dockerfile for auto-docs CLI
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download || true

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o auto-docs main.go

FROM alpine:latest
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY --from=builder /app/auto-docs /usr/local/bin/auto-docs

ENTRYPOINT ["auto-docs"]
