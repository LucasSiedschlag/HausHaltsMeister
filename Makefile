DB_NAME ?= haushaltsmeister
DB_USER ?= postgres
DB_PASS ?= postgres
DB_HOST ?= localhost
DB_PORT ?= 5432

DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

.PHONY: all build run migrate migrate-status sqlc test clean

all: build

build:
	go build -o bin/api cmd/api/main.go

run:
	go run cmd/api/main.go

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate:
	tern migrate -c internal/db/tern.conf -m migrations

migrate-status:
	tern status -c internal/db/tern.conf -m migrations

sqlc:
	sqlc generate

test:
	go test ./...

clean:
	rm -rf bin
