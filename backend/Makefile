-include .env
export

.PHONY: run migrate-up migrate-down migrate-status build tidy

run:
	go run ./cmd/api

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

migrate-status:
	go run ./cmd/migrate status

build:
	go build -o bin/api ./cmd/api

tidy:
	go mod tidy
