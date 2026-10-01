# Multi-stage Dockerfile for all three Golang services.
# Build with: docker build --target <api|worker|mock-campus-api> -t presensi-<service> .

# ─── Stage 1: Builder ─────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy go mod files first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build all three binaries
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /out/worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /out/mock-campus-api ./cmd/mock-campus-api

# ─── Stage 2: API Service ──────────────────────────────────────────────────────
FROM alpine:3.19 AS api

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /out/api .
COPY migrations/ ./migrations/

EXPOSE 8080
ENTRYPOINT ["./api"]

# ─── Stage 3: Sync Worker ─────────────────────────────────────────────────────
FROM alpine:3.19 AS worker

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /out/worker .
COPY migrations/ ./migrations/

ENTRYPOINT ["./worker"]

# ─── Stage 4: Mock Campus API ─────────────────────────────────────────────────
FROM alpine:3.19 AS mock-campus-api

RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=builder /out/mock-campus-api .

EXPOSE 9000
ENTRYPOINT ["./mock-campus-api"]
