.PHONY: build run clean lint test swagger

APP_NAME := go-admin
BUILD_DIR := ./dist

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

run:
	go run ./cmd/server

clean:
	rm -rf $(BUILD_DIR)

lint:
	go vet ./...

test:
	go test ./...

swagger:
	swag init -g cmd/server/main.go -o docs

deps:
	go mod tidy

help:
	@echo "Available commands:"
	@echo "  make build    - Build the application"
	@echo "  make run      - Run the application"
	@echo "  make clean    - Clean build artifacts"
	@echo "  make lint     - Run go vet"
	@echo "  make test     - Run tests"
	@echo "  make swagger  - Generate swagger docs"
	@echo "  make deps     - Tidy go modules"
