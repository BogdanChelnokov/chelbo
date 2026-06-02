.PHONY: help build run test dev docker-up docker-down migrate-up migrate-down

help:
	@echo "Available commands:"
	@echo "  make dev          - Run in development mode (air)"
	@echo "  make build        - Build binary"
	@echo "  make test         - Run tests"
	@echo "  make docker-up    - Start Docker containers"
	@echo "  make docker-down  - Stop Docker containers"
	@echo "  make migrate-up   - Apply database migrations"
	@echo "  make migrate-down - Rollback migrations"

dev:
	cd backend && air

build:
	cd backend && go build -o bin/chelbo cmd/server/main.go

test:
	cd backend && go test -v ./...

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate-up:
	cd backend && go run cmd/migrate/main.go up

migrate-down:
	cd backend && go run cmd/migrate/main.go down