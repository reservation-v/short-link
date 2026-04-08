POSTGRES_DSN ?= postgres://postgres:postgres@localhost:5433/shortener?sslmode=disable

.PHONY: help run-memory run-postgres docker-up docker-down reset test

help:
	@echo "run-memory    - run the service locally with in-memory storage"
	@echo "run-postgres  - run the service locally with PostgreSQL"
	@echo "docker-up     - run the full stack in Docker"
	@echo "docker-down   - stop the Docker stack"
	@echo "reset         - stop the Docker stack and remove volumes"

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
