DEV_ENV_FILE=deploy/.env.dev
DEV_COMPOSE_FILE=deploy/docker-compose.dev.yml

.PHONY: dev-up dev-down dev-build

dev-up:
	docker compose -f $(DEV_COMPOSE_FILE) --env-file $(DEV_ENV_FILE) up

dev-down:
	docker compose -f $(DEV_COMPOSE_FILE) --env-file $(DEV_ENV_FILE) down

dev-build:
	docker compose -f $(DEV_COMPOSE_FILE) --env-file $(DEV_ENV_FILE) build

.PHONY: backend-test
backend-test:
	docker compose -f $(DEV_COMPOSE_FILE) --env-file $(DEV_ENV_FILE) run --rm \
		-w /app backend go test -v ./...

.PHONY: backend-test-coverage
backend-test-coverage:
	docker compose -f $(DEV_COMPOSE_FILE) --env-file $(DEV_ENV_FILE) run --rm \
		-w /app backend go test -v -coverprofile=coverage.out ./...
	docker compose -f $(DEV_COMPOSE_FILE) --env-file $(DEV_ENV_FILE) run --rm \
		-w /app backend go tool cover -html=coverage.out -o coverage.html

.PHONY: backend-test-handler
backend-test-handler:
	docker compose -f $(DEV_COMPOSE_FILE) --env-file $(DEV_ENV_FILE) run --rm \
		-w /app backend go test -v ./internal/interface/handler/...

.PHONY: backend-test-usecase
backend-test-usecase:
	docker compose -f $(DEV_COMPOSE_FILE) --env-file $(DEV_ENV_FILE) run --rm \
		-w /app backend go test -v ./internal/usecase/...

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
	@echo "[Generate] Running pnpm run gen:client"
	cd frontend && pnpm run gen:client

.PHONY: gen-all
gen-all: gen-swagger-v2 gen-openapi-v3 gen-client
	@echo "[Done] All generation tasks completed."

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  dev-up                  - Start development environment"
	@echo "  dev-down                - Stop development environment"
	@echo "  dev-build               - Build development containers"
	@echo "  backend-test            - Run all backend tests"
	@echo "  backend-test-coverage   - Run tests with coverage report"
	@echo "  backend-test-handler    - Run handler tests only"
	@echo "  backend-test-usecase    - Run usecase tests only"
	@echo "  gen-swagger-v2          - Generate Swagger v2 docs"
	@echo "  gen-openapi-v3          - Generate OpenAPI v3 docs"
	@echo "  gen-client              - Generate API client"
	@echo "  gen-all                 - Run all generation tasks"