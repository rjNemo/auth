BIN_DIR := bin
BIN_NAME := auth-server
FMT_PATHS := $(shell go list -f '{{.Dir}}' ./...)
MIGRATIONS_DIR := internal/driver/db/migrations
SQLC_CONFIG := internal/driver/db/sqlc.yaml
DB_URL ?= $(AUTH_DATABASE_URL)
DB_URL := $(strip $(DB_URL))
ifeq ($(DB_URL),)
DB_URL := postgres://localhost/auth_dev?sslmode=disable
endif

.PHONY: run dev build test fmt lint tidy clean migrate-status migrate-up migrate-down migrate-reset migrate-new sqlc-generate

run:
	go run ./cmd/server

dev:
	air

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BIN_NAME) ./cmd/server

test:
	go test ./... -cover -count=1

fmt:
	gofmt -w $(FMT_PATHS)

lint:
	golangci-lint run

tidy:
	go mod tidy

clean:
	rm -rf $(BIN_DIR)

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" status

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down

migrate-reset:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" reset

migrate-new:
	@if [ -z "$(name)" ]; then \
		echo "usage: make migrate-new name=add_feature"; \
		exit 1; \
	fi
	goose -dir $(MIGRATIONS_DIR) create $(name) sql

sqlc-generate:
	sqlc generate -f $(SQLC_CONFIG)
