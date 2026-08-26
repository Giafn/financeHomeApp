-include .env
export

.PHONY: run worker migrate-up migrate-down migrate-status build tidy

run:
	go run ./cmd/api

worker:
	go run ./cmd/worker

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

migrate-status:
	go run ./cmd/migrate status

build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker

tidy:
	go mod tidy
