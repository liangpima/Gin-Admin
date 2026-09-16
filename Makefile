.PHONY: build run clean lint test swagger migrate migrate-status

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

# 执行未应用的数据库迁移（读取 config/config.yaml 里的 database.* 配置）
migrate:
	go run ./cmd/migrate

# 只查看迁移状态，不执行任何 SQL
migrate-status:
	go run ./cmd/migrate -status

swagger:
	swag init -g cmd/server/main.go -o docs

deps:
	go mod tidy

help:
	@echo "Available commands:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make lint           - Run go vet"
	@echo "  make test           - Run tests"
	@echo "  make migrate        - Apply pending DB migrations"
	@echo "  make migrate-status - Show DB migration status"
	@echo "  make swagger        - Generate swagger docs"
	@echo "  make deps           - Tidy go modules"
