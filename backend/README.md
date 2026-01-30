# Schedule Optimizer Backend

Go backend for WWU Schedule Optimizer using Gin, SQLite, and sqlc.

## Prerequisites

### Required Software

```bash
# Go 1.24+
# See https://go.dev/doc/install
go version  # should show go1.24 or later

# SQLite3
sudo apt install sqlite3  # Debian/Ubuntu
# or
brew install sqlite3      # macOS

# sqlc - SQL code generator
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# golang-migrate - Database migrations
# Must be built with sqlite3 support
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Verify installations:
```bash
go version
sqlite3 --version
sqlc version
migrate -version
```

## Quick Start

```bash
# 1. Copy environment configuration
cp .env.example .env

# 2. Apply database migrations (creates SQLite database)
make migrate-up

# 3. Run the server
make run
```

Server runs at http://localhost:48920 by default.

## Build Commands

| Command | Description |
|---------|-------------|
| `make build` | Build binary to `bin/server` |
| `make run` | Run server with `go run` |
| `make test` | Run all tests |
| `make test-coverage` | Run tests with coverage report |
| `make bench` | Run benchmark tests |
| `make sqlc` | Regenerate sqlc code after modifying `queries.sql` |
| `make migrate-up` | Apply pending migrations |
| `make migrate-down` | Rollback last migration |
| `make migrate-create name=foo` | Create new migration files |
| `make tidy` | Tidy Go module dependencies |

## Configuration

Environment variables (set in `.env` or export):

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `48920` | Server port |
| `ENVIRONMENT` | `development` | `development` or `production` |
| `DATABASE_PATH` | `data/schedule.db` | SQLite database file path |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000,http://localhost:5173` | Comma-separated CORS origins |

## Architecture

```
backend/
├── cmd/server/main.go        # Entry point, route definitions
├── migrations/               # golang-migrate SQL files
├── internal/
│   ├── config/               # Environment configuration
│   ├── db/                   # Database connection (WAL mode)
│   ├── store/                # sqlc generated code + queries
│   │   ├── queries.sql       # Named SQL queries (schema from migrations)
│   │   └── *.go              # Generated (don't edit)
│   ├── cache/                # In-memory cache for schedule generation
│   └── testutil/             # Shared test utilities
├── sqlc.yaml                 # sqlc configuration
└── Makefile
```

### Data Flow

```
Banner API → Scraper → SQLite → Cache (in-memory) → API → Frontend
```

## API Endpoints

API endpoints are under development. See GitHub issues #12 (API Contract) and #20 (Search Service).

### Health
- `GET /health` - Health check

## Database

SQLite with WAL mode for concurrent read performance.

### Migrations

Schema changes use golang-migrate:

```bash
# Create new migration
make migrate-create name=add_new_table

# Apply migrations
make migrate-up

# Rollback
make migrate-down
```

### sqlc Workflow

After modifying `internal/store/queries.sql`:

```bash
make sqlc  # Regenerates internal/store/*.go
```

**Note:** sqlc reads the schema directly from the migrations folder (configured in `sqlc.yaml`).

## Development

### Adding a New Query

1. Add query to `internal/store/queries.sql`:
   ```sql
   -- name: GetFoo :one
   SELECT * FROM foo WHERE id = ?;
   ```

2. Regenerate: `make sqlc`

3. Use in code:
   ```go
   result, err := queries.GetFoo(ctx, id)
   ```

### Adding a Schema Change

1. Create migration: `make migrate-create name=add_bar_column`
2. Edit `migrations/NNNNNN_add_bar_column.up.sql` and `.down.sql`
3. Apply: `make migrate-up`
4. Run `make sqlc` to regenerate code with new schema

## Testing

```bash
make test          # Run all tests
make test-coverage # With coverage
make bench         # Run benchmarks
```

See `CLAUDE.md` in the project root for testing strategy details.
