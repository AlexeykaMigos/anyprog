DOCKER_COMPOSE = docker-compose
APP_SERVICE = app
DB_SERVICE = db
MIGRATE_CMD = migrate -path /app/migrations -database postgres://postgres:secret@db:5432/pgdb?sslmode=disable


.PHONY: build up down .env


build:
	$(DOCKER_COMPOSE) build

up:
	$(DOCKER_COMPOSE) up -d

down:
	$(DOCKER_COMPOSE) down

.env:
	@echo "Generating .env file with default values..."
	@echo "DB_HOST=localhost" > .env
	@echo "DB_PORT=5432" >> .env
	@echo "DB_USER=postgres" >> .env
	@echo "DB_PASSWORD=postgres" >> .env
	@echo "DB_NAME=pgdb" >> .env
	@echo ".env file created successfully!"

# Цель для удаления .env файла
clean-env:
	@if [ -f .env ]; then \
		rm .env; \
		echo ".env file removed."; \
	else \
		echo ".env file does not exist."; \
	fi
