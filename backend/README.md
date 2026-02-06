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
│   ├── generator/            # Schedule generation (bitmask + backtracking)
│   │   ├── service.go        # Generate() entry point
│   │   ├── bitmask.go        # O(1) conflict detection
│   │   ├── backtrack.go      # Recursive enumeration
│   │   └── scorer.go         # Gap, Start, End scoring
│   └── testutil/             # Shared test utilities
├── sqlc.yaml                 # sqlc configuration
└── Makefile
```

### Data Flow

```
Banner API → Scraper → SQLite → Cache (in-memory) → API → Frontend
```

## Schedule Generator

The schedule generator uses bitmask-based conflict detection with backtracking to enumerate all valid schedule combinations.

### Algorithm

- **Bitmask conflict detection**: O(1) conflict check via bitwise AND on `[8]uint64` (512 bits for 450 time slots)
- **10-minute granularity**: 7am-10pm = 90 slots/day × 5 days = 450 bits
- **Backtracking with pruning**: Generates schedules in order of course count, stops early when limit reached
- **Scoring**: Gap (minimize gaps between classes), Start (prefer later starts), End (prefer earlier ends)

### Performance

Benchmarks comparing v2 (bitmask) vs old implementation (O(n²) precomputed conflict matrix):

| Test | Sections | Old (main) | New (v2) | Speedup |
|------|----------|------------|----------|---------|
| 5 courses | 100 | 7.0ms | 1.5ms | **4.7x** |
| 8 courses | 136 | 15.4ms | 1.7ms | **9.0x** |
| 10 courses | 160 | 14.5ms | 1.8ms | **8.1x** |
| 13 courses | 169 | 13.3ms | 1.8ms | **7.4x** |

*Tested with synthetic course data, 20k schedule limit, on same hardware.*

### Limits

| Constant | Value | Description |
|----------|-------|-------------|
| `MaxInputCourses` | 13 | Maximum courses in request |
| `MaxSchedulesToGenerate` | 20,000 | Safety limit during generation |
| `MaxSchedulesToReturn` | 2,000 | Schedules returned to client (sorted by score) |
| `DefaultMaxCourses` | 8 | Default max courses per schedule |

## API Endpoints

API endpoints are under development. See GitHub issues #12 (API Contract) and #20 (Search Service).

### Health
- `GET /health` - Health check

### Schedule Generation
- `POST /generate` - Generate schedule combinations for requested courses

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

## Admin Operations

### Announcements

Announcements are managed directly via `sqlite3`. The frontend fetches the active announcement and shows a dismissable banner (tracked per announcement ID in localStorage).

```bash
# Set a new announcement (types: 'info', 'warning', 'beta')
sqlite3 data/schedule.db <<'SQL'
INSERT INTO announcements(title, body, type)
VALUES('Beta Notice', 'This is a beta release, expect bugs!', 'beta');
SQL

# Update: deactivate old, insert new
sqlite3 data/schedule.db <<'SQL'
UPDATE announcements SET active = 0;
INSERT INTO announcements(title, body, type)
VALUES('Maintenance', 'Scheduled maintenance tonight at 10pm.', 'warning');
SQL

# Clear all announcements
sqlite3 data/schedule.db "UPDATE announcements SET active = 0;"
```

> **Note:** Use heredocs (`<<'SQL'`) to avoid shell escaping issues with `!` and other special characters.

### Viewing Feedback

User feedback is stored in the `feedback` table with session IDs linking to analytics.

```bash
# View recent feedback
sqlite3 data/schedule.db "SELECT id, session_id, message, created_at FROM feedback ORDER BY id DESC LIMIT 20;"

# Count feedback entries
sqlite3 data/schedule.db "SELECT COUNT(*) FROM feedback;"
```

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
