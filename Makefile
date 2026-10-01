.PHONY: help dev build test test-crypto migrate-up migrate-down mock-api worker clean

# Default target
help:
	@echo "Smart Campus Geofence Attendance System"
	@echo ""
	@echo "Usage:"
	@echo "  make dev           Start Docker dev environment (PostgreSQL + Redis)"
	@echo "  make run-api       Run API server locally"
	@echo "  make run-worker    Run sync worker locally"
	@echo "  make run-mock      Run Mock Campus API locally"
	@echo "  make run-web       Run Vue.js Web Dashboard locally"
	@echo "  make build         Build all binaries"
	@echo "  make test          Run all tests"
	@echo "  make test-crypto   Run crypto module tests only"
	@echo "  make migrate-up    Apply all DB migrations"
	@echo "  make migrate-down  Rollback last DB migration"
	@echo "  make clean         Remove built binaries"

# ─── Dev Environment (Docker) ─────────────────────────────────────────────────
up:
	docker compose up -d --build
	@echo "✅ All services are up and running!"
	@echo "   🌐 Web Dashboard: http://localhost:5173"
	@echo "   🚀 API Server:    http://localhost:8080"
	@echo "   🏫 Mock SIAKAD:   http://localhost:9001"
	@echo "   🐘 PostgreSQL:    localhost:5434"
	@echo "   ⚡ Redis:         localhost:6380"

down:
	docker compose down
	@echo "🛑 All services stopped."

logs:
	docker compose logs -f

ps:
	docker compose ps

dev:
	docker compose up -d postgres redis
	@echo "✅ PostgreSQL and Redis are running"
	@echo "   Postgres: localhost:5434"
	@echo "   Redis:    localhost:6380"

dev-all: up
dev-down: down

# ─── Run Services Locally ─────────────────────────────────────────────────────
run-api:
	cd backend && go run ./cmd/api

run-worker:
	cd backend && go run ./cmd/worker

run-mock:
	cd backend && APP_PORT=9001 go run ./cmd/mock-campus-api

run-web:
	cd web-dashboard && npm run dev

# ─── Build ────────────────────────────────────────────────────────────────────
build:
	@mkdir -p bin
	cd backend && go build -ldflags="-w -s" -o ../bin/api ./cmd/api
	cd backend && go build -ldflags="-w -s" -o ../bin/worker ./cmd/worker
	cd backend && go build -ldflags="-w -s" -o ../bin/mock-campus-api ./cmd/mock-campus-api
	@echo "✅ Binaries built in ./bin/"

# ─── Testing ──────────────────────────────────────────────────────────────────
test:
	cd backend && go test ./... -v -race -timeout 60s

test-crypto:
	cd backend && go test ./internal/crypto/... -v

test-cover:
	cd backend && go test ./... -coverprofile=coverage.out
	cd backend && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: backend/coverage.html"

# ─── Database Migrations ──────────────────────────────────────────────────────
DB_URL ?= postgres://presensi_user:presensi_secret@localhost:5434/presensi_db?sslmode=disable

MIGRATE := $(shell go env GOPATH)/bin/migrate

migrate-up:
	@test -f $(MIGRATE) || (echo "Installing golang-migrate..." && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	$(MIGRATE) -path backend/migrations -database "$(DB_URL)" up
	@echo "✅ Migrations applied"

migrate-down:
	$(MIGRATE) -path backend/migrations -database "$(DB_URL)" down 1
	@echo "✅ Last migration rolled back"

migrate-status:
	$(MIGRATE) -path backend/migrations -database "$(DB_URL)" version

# ─── Utilities ────────────────────────────────────────────────────────────────
gen-aes-key:
	@cd backend && go run -e 'package main; import ("fmt"; "github.com/ilham/presensi-online/backend/internal/crypto"); func main() { k, _ := crypto.GenerateKeyHex(); fmt.Println(k) }' 2>/dev/null || \
	openssl rand -hex 32
	@echo "👆 Copy this key to AES_KEY_HEX in .env"

tidy:
	cd backend && go mod tidy

vet:
	cd backend && go vet ./...

clean:
	rm -rf bin/
	cd backend && go clean -cache
