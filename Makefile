.PHONY: db-up db-down migrate-up migrate-down generate dev webhook worker api

-include .env
export

DB_URL ?= "postgres://chaos:password@localhost:5432/chaosci?sslmode=disable"

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

migrate-up:
	~/go/bin/migrate -path migrations -database $(DB_URL) up

migrate-down:
	~/go/bin/migrate -path migrations -database $(DB_URL) down

generate:
	sqlc generate

webhook:
	go run cmd/webhook/main.go

worker:
	go run cmd/worker/main.go

api:
	go run cmd/api/main.go

dev:
	@echo "Starting webhook, worker, and api servers in background..."
	@go run cmd/webhook/main.go &
	@go run cmd/worker/main.go &
	@go run cmd/api/main.go &
	@wait
