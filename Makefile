.PHONY: help build run test clean docker-build docker-up docker-down migrate

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	go build -o bin/bookwise-api ./cmd/server

run: ## Run the application
	go run ./cmd/server/main.go

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	rm -rf bin/
	go clean

deps: ## Download dependencies
	go mod download
	go mod tidy

docker-build: ## Build Docker image
	docker-compose build

docker-up: ## Start Docker containers
	docker-compose up -d

docker-down: ## Stop Docker containers
	docker-compose down

docker-logs: ## Show Docker logs
	docker-compose logs -f api

docker-restart: ## Restart Docker containers
	docker-compose restart api

migrate: ## Run database migrations
	go run ./cmd/server/main.go

lint: ## Run linter
	golangci-lint run

.DEFAULT_GOAL := help

