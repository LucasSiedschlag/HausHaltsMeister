DB_NAME ?= ledger
DB_USER ?= postgres
DB_PASS ?= postgres
DB_HOST ?= localhost
DB_PORT ?= 5434

DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
TERN ?= tern
TERN_CONF ?= internal/db/tern.conf

.PHONY: all build run migrate migrate-status migrate-prod migrate-status-prod sqlc sqlc-ledger test clean

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
	$(TERN) migrate -c $(TERN_CONF) -m migrations

migrate-status:
	$(TERN) status -c $(TERN_CONF) -m migrations

migrate-prod:
	@if [ "$(TERN_CONF)" = "internal/db/tern.conf" ]; then \
		echo "Set TERN_CONF to a production config path."; \
		exit 1; \
	fi
	$(TERN) migrate -c $(TERN_CONF) -m migrations

migrate-status-prod:
	@if [ "$(TERN_CONF)" = "internal/db/tern.conf" ]; then \
		echo "Set TERN_CONF to a production config path."; \
		exit 1; \
	fi
	$(TERN) status -c $(TERN_CONF) -m migrations

migrestore:
	@echo "Restoring database..."
	@docker exec -i cashflow-db psql -U user -d cashflow < backups/latest_backup.sql

swagger:
	@echo "Generating Swagger documentation..."
	@$(HOME)/go/bin/swag init -g cmd/api/main.go -o internal/adapters/http/docs

sqlc:
	sqlc generate

sqlc-ledger:
	sqlc generate -f sqlc-ledger.yaml

test:
	go test ./...

clean:
	rm -rf bin
