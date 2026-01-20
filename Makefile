.PHONY: help run build test clean docker-up docker-down docker-compose-up docker-compose-down docker-compose-dev docker-compose-logs docker-compose-build

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the API server
	go run cmd/api/main.go

build: ## Build the API server binary
	go build -o bin/api cmd/api/main.go

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -cover ./...

clean: ## Clean build artifacts
	rm -rf bin/

docker-up: ## Start MongoDB with Docker
	docker run -d -p 27017:27017 --name mongodb-datacollector mongo:latest

docker-down: ## Stop MongoDB Docker container
	docker stop mongodb-datacollector && docker rm mongodb-datacollector

deps: ## Download Go dependencies
	go mod download
	go mod tidy

fmt: ## Format Go code
	go fmt ./...

lint: ## Run linter
	golangci-lint run

# Docker Compose commands
docker-compose-build: ## Build all Docker images
	docker-compose build

docker-compose-up: ## Start all services with Docker Compose (production)
	docker-compose up -d

docker-compose-down: ## Stop all services with Docker Compose
	docker-compose down

docker-compose-dev: ## Start all services in development mode
	docker-compose -f docker-compose.dev.yml up

docker-compose-logs: ## View logs from all services
	docker-compose logs -f

docker-compose-restart: ## Restart all services
	docker-compose restart

docker-compose-clean: ## Stop and remove all containers, volumes, and images
	docker-compose down -v --rmi all

.DEFAULT_GOAL := help
