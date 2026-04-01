HTTP_ADDR ?= :8081
POSTGRES_DSN ?= postgres://postgres:postgres@localhost:5433/shortener?sslmode=disable

.PHONY: help run-memory postgres-up postgres-down migrate-up postgres-init run-postgres docker-up docker-down test

help:
	@echo "run-memory    - run the service locally with in-memory storage"
	@echo "postgres-up   - start PostgreSQL in Docker"
	@echo "migrate-up    - apply migrations to PostgreSQL"
	@echo "postgres-init - start PostgreSQL and apply migrations"
	@echo "run-postgres  - run the service locally with PostgreSQL"
	@echo "docker-up     - run the full stack in Docker"
	@echo "docker-down   - stop the Docker stack"

run-memory:
	go run ./cmd/short-link --storage=memory --http-addr=$(HTTP_ADDR)

postgres-up:
	docker compose up -d postgres

postgres-down:
	docker compose stop postgres

migrate-up:
	docker compose run --rm migrate

postgres-init: postgres-up migrate-up

run-postgres:
	go run ./cmd/short-link --storage=postgres --postgres-dsn='$(POSTGRES_DSN)' --http-addr=$(HTTP_ADDR)

docker-up:
	docker compose up --build

docker-down:
	docker compose down

test:
	go test ./...
