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

.PHONY: gen-swagger-v2
gen-swagger-v2:
	@echo "[OAS2] Generate swagger.yaml & swagger.json"
	docker run --rm -v $(PWD)/backend:/app -w /app golang:1.25-alpine \
	  sh -c "go install github.com/swaggo/swag/cmd/swag@latest && \
	  swag fmt && \
	  swag init -g cmd/server/main.go --parseDependency"

.PHONY: gen-openapi-v3
gen-openapi-v3:
	@echo "[OAS3] Convert swagger.yaml → openapi.yaml"
	docker run --rm -v $(PWD)/backend/docs:/work openapitools/openapi-generator-cli:latest-release \
	  generate -i /work/swagger.yaml -o /work/v3 -g openapi-yaml --minimal-update

	@echo "[OAS3] Convert swagger.json → openapi.json"
	docker run --rm -v $(PWD)/backend/docs:/work openapitools/openapi-generator-cli:latest-release \
	  generate -s -i /work/swagger.json -o /work/v3/openapi -g openapi --minimal-update

	@echo "[Cleanup]"
	docker run --rm -v $(PWD)/backend/docs/v3:/work golang:1.21-alpine \
	  sh -c "mv /work/openapi/openapi.yaml /work && mv /work/openapi/openapi.json /work && rm -rf /work/openapi"

.PHONY: gen-client
gen-client:
	@echo "[Clean] Remove old generated client"
	rm -rf frontend/src/api/__generated__
	@echo "[Generate] Running npm run gen:client"
	cd frontend && npm run gen:client