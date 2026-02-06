.PHONY: build run test sqlc migrate-up migrate-down migrate-create build-frontend

# Database path
DB_PATH ?= data/schedule.db
MIGRATIONS_PATH = migrations

# Build frontend (compiles to internal/static/dist)
build-frontend:
	cd ../frontend && pnpm build

# Build the server (includes frontend)
build: build-frontend
	go build -o bin/server ./cmd/server

# Build server only (assumes frontend already built)
build-server:
	go build -o bin/server ./cmd/server

# Run the server
run:
	go run ./cmd/server

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Run benchmarks
bench:
	go test -bench=. ./...

# Generate sqlc code
sqlc:
	sqlc generate

# Run database migrations
migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "sqlite3://$(DB_PATH)" up

migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "sqlite3://$(DB_PATH)" down

migrate-drop:
	migrate -path $(MIGRATIONS_PATH) -database "sqlite3://$(DB_PATH)" drop -f

# Create a new migration (usage: make migrate-create name=add_foo_table)
migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)

# Tidy dependencies
tidy:
	go mod tidy
