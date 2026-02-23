.PHONY: help dev up down build logs clean migrate

COMPOSE_FILE := deploy/docker-compose.yml
COMPOSE_DEV := deploy/docker-compose.dev.yml
ENV_FILE := deploy/.env

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## Start development stack
	docker compose -f $(COMPOSE_FILE) -f $(COMPOSE_DEV) --env-file $(ENV_FILE) up --build

up: ## Start production stack
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) up -d --build

down: ## Stop all services
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) down

logs: ## Tail logs
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) logs -f

build: ## Build all images
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) build

clean: ## Remove all containers, volumes, and images
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) down -v --rmi local

infra: ## Start only postgres and minio
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) up -d postgres minio

migrate: ## Run database migrations
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) exec backend /app/server migrate

psql: ## Connect to PostgreSQL
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) exec postgres psql -U sharepwd -d sharepwd
