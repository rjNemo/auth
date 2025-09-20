BIN_DIR := bin
BIN_NAME := auth-server
FMT_PATHS := $(shell go list -f '{{.Dir}}' ./...)

.PHONY: run dev build test fmt lint tidy clean

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
