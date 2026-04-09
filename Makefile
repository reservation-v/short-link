POSTGRES_DSN ?= postgres://postgres:postgres@localhost:5433/shortener?sslmode=disable
LOAD_VUS ?= 200
LOAD_DURATION ?= 15s
LOAD_SHARED_URL ?= https://example.com/load-test/shared
LOADTEST_PROJECT ?= short-link-loadtest

.PHONY: help run-memory run-postgres docker-up docker-down reset test test-integration load-test

help:
	@echo "run-memory    - run the service locally with in-memory storage"
	@echo "run-postgres  - run the service locally with PostgreSQL"
	@echo "docker-up     - run the full stack in Docker"
	@echo "docker-down   - stop the Docker stack"
	@echo "reset         - stop the Docker stack and remove volumes"
	@echo "test-integration - run postgres integration tests (requires Docker)"
	@echo "load-test     - run isolated k6 load test using docker-compose.yml + docker-compose.k6.yml"
	@echo "load-test-running - run k6 load test against an already running service"

run-memory:
	go run ./cmd/short-link --storage=memory

run-postgres:
	go run ./cmd/short-link --storage=postgres --postgres-dsn='$(POSTGRES_DSN)'

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

reset:
	docker compose down -v

test:
	go test ./...

test-integration:
	go test -tags=integration ./internal/storage/postgres -count=1

load-test:
	@set -e; \
	COMPOSE="docker compose -f docker-compose.yml -f docker-compose.k6.yml -p $(LOADTEST_PROJECT)"; \
	trap "$$COMPOSE down -v --remove-orphans >/dev/null 2>&1 || true" EXIT; \
	$$COMPOSE up --build -d postgres migrate app; \
	$$COMPOSE run --rm \
		-e BASE_URL='http://app:8081' \
		-e VUS='$(LOAD_VUS)' \
		-e DURATION='$(LOAD_DURATION)' \
		-e SHARED_URL='$(LOAD_SHARED_URL)' \
		k6
