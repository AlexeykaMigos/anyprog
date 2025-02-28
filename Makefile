DOCKER_COMPOSE = docker-compose
APP_SERVICE = app
DB_SERVICE = db
MIGRATE_CMD = migrate -path /app/migrations -database postgres://postgres:secret@db:5432/pgdb?sslmode=disable


.PHONY: build up down


build:
	$(DOCKER_COMPOSE) build

up:
	$(DOCKER_COMPOSE) up -d

down:
	$(DOCKER_COMPOSE) down


