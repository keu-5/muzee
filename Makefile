# Makefile

DEV_ENV_FILE=deploy/.env.dev
DEV_COMPOSE_FILE=deploy/docker-compose.dev.yml

.PHONY: dev-up dev-down dev-build

dev-up:
	docker compose -f $(DEV_COMPOSE_FILE) --env-file $(DEV_ENV_FILE) up

dev-down:
	docker compose -f $(DEV_COMPOSE_FILE) --env-file $(DEV_ENV_FILE) down

dev-build:
	docker compose -f $(DEV_COMPOSE_FILE) --env-file $(DEV_ENV_FILE) build
