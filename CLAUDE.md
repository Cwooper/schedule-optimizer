# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Schedule Optimizer is a course scheduling tool for WWU students. `main` holds the v2 rewrite, now deployed. The legacy v1 lives in git history (and the retained `v2` branch). Stack:

- **Backend:** Go (Gin + SQLite + sqlc)
- **Frontend:** React 19 + TypeScript + Vite + Tailwind v4
  - **UI Components:** shadcn/ui (new-york style, neutral base)
  - **State Management:** Zustand (persisted to localStorage)
  - **Data Fetching:** TanStack Query
  - **No routing:** Tab state stored in localStorage, defaults to "schedule"

## Development Philosophy

- **Act on the small, discuss the big.** Just implement small, well-scoped, or clearly-specified tasks — skip the brainstorm-first ritual. For architectural changes or non-trivial features, talk through approach and tradeoffs first (use plan mode) before writing code.
- **Test-heavy development.** Write tests alongside implementation. Benchmark tests are critical for schedule generation and scraping performance.
- **Self-review non-trivial work.** After a non-trivial change, spin off an agent for an unbiased code review that checks: architecture alignment (does this belong here?), file/function lengths (keep functions focused, split large files), scope creep (only what was requested), and whether tests exercise real code paths rather than mocking unrealistic scenarios.
- **Course-centric, not term-centric.** Data structures should support cross-term search. Organize by unique course ID, not term→CRN hierarchy.
- **Git workflow.** Work on feature branches and commit freely as you go. Open PRs against `main` only after checking in first — the user squash-merges them. Don't add AI attribution or co-author trailers to commits.
- **Bump the version.** Don't forget to bump `frontend/package.json` `version` (rendered in the footer) with each piece of work — patch for fixes/small changes, minor for features. Fold the bump into the feature branch so shipped work never goes out with a stale version.

## Prerequisites

Install these on a fresh server:

```bash
# Go 1.26+
# https://go.dev/doc/install

# SQLite3
sudo apt install sqlite3

# sqlc (SQL code generator)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# golang-migrate (database migrations)
# With sqlite3 support - see https://github.com/golang-migrate/migrate
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Frontend
# Node.js 22.13+ and pnpm (the floor for Vitest 5 and ESLint 10)
```

## Build & Run Commands

### Backend (Go)

```bash
cd backend

# Database migrations (run before first start)
make migrate-up
# Or manually:
migrate -path migrations -database "sqlite3://data/schedule.db" up

# Run server
go run cmd/server/main.go

# Build binary
go build -o bin/server ./cmd/server

# Generate sqlc code (after modifying queries.sql)
sqlc generate

# Run tests
go test ./...

# Run benchmarks
go test -bench=. ./...
```

### Makefile targets

```bash
make build          # Build binary
make run            # Run server
make test           # Run tests
make sqlc           # Regenerate sqlc code
make migrate-up     # Apply migrations
make migrate-down   # Rollback migrations
make migrate-create name=add_foo  # Create new migration
```

### Frontend

```bash
cd frontend
pnpm install
pnpm dev        # development server (http://localhost:5173)
pnpm build      # production build
pnpm lint       # run eslint
pnpm test       # run tests
pnpm test:watch # run tests in watch mode
```

## Architecture

### Backend Structure

```
backend/
├── cmd/server/main.go        # Entry point, route definitions, graceful shutdown
├── migrations/               # golang-migrate SQL files
│   ├── 000001_initial_schema.up.sql
│   └── 000001_initial_schema.down.sql
├── internal/
│   ├── config/               # Environment configuration
│   ├── db/                   # Database connection only
│   │   └── db.go
│   ├── store/                # sqlc generated code + queries
│   │   ├── queries.sql       # Query definitions (schema comes from migrations)
│   │   └── *.go              # Generated (don't edit)
│   ├── scraper/              # Banner API scraper
│   │   ├── scraper.go        # Main ScrapeTerm logic (see package doc for design)
│   │   ├── client.go         # HTTP client with cookie jar
│   │   ├── response.go       # Banner API response types
│   │   └── store.go          # Persistence helpers
│   ├── api/                  # HTTP handlers, one file per area (terms, sections, calendar, generate, feedback)
│   ├── calendar/             # Registrar calendar scraper (see package doc for design)
│   │   ├── calendar.go       # ScrapeAll: term dates, finals schedule, important dates
│   │   ├── schedule.go       # Term phases from scraped dates, estimates as fallback
│   │   ├── store.go          # Transactional writes: term dates upserted, finals and important dates replaced per term
│   │   ├── term.go           # LoadTerm: a term's registration phases, holidays, deadlines, finals week (GET /api/calendar/:term)
│   │   └── finals.go         # FinalFor: section meetings -> exam slot, certain or a guess with a reason
│   ├── meeting/              # meeting.Time, the meeting payload every section response shares
│   ├── termcode/             # Term code (YYYYQQ) parsing and helpers
│   ├── jobs/                 # Background job scheduler
│   │   ├── service.go        # Job interface + Service (runs due jobs in registration order)
│   │   ├── calendar.go       # CalendarJob: Banner term list + registrar calendar, on startup and daily
│   │   └── scrape.go         # ScrapeJob: each term due by its phase and last_scraped_at; failures retry hourly
│   ├── cache/                # In-memory schedule cache (sections carry their computed final)
│   └── generator/            # Schedule generation (bitmask + backtracking)
│       ├── service.go        # Generate() entry point, course group building
│       ├── bitmask.go        # O(1) conflict detection via [8]uint64
│       ├── backtrack.go      # Recursive enumeration with pruning
│       ├── scorer.go         # Gap, Start, End scoring
│       └── types.go          # Request/Response types
├── sqlc.yaml
├── Makefile
└── .env.example
```

### Frontend Structure

```
frontend/
├── src/
│   ├── components/          # React components
│   │   └── ui/              # shadcn/ui components (generated)
│   ├── hooks/               # Custom React hooks
│   │   └── use-api.ts       # TanStack Query hooks for API calls
│   ├── lib/
│   │   ├── api.ts           # Typed fetch functions for backend API
│   │   ├── dates.ts         # YYYY-MM-DD helpers (API dates are Pacific calendar dates)
│   │   ├── term-status.ts   # Sidebar term status line and key dates rows
│   │   ├── finals.ts        # Finals week grid data, guess reason labels
│   │   ├── ics.ts           # Hand-written RFC 5545 export (TZID times, EXDATE holidays)
│   │   └── utils.ts         # cn() helper for Tailwind class merging
│   ├── stores/
│   │   └── app-store.ts     # Zustand store (persisted to localStorage)
│   ├── App.tsx
│   ├── main.tsx             # Entry point with QueryClientProvider
│   └── index.css            # Tailwind + shadcn/ui CSS variables
├── public/
│   ├── favicon-light.svg    # Favicon for light mode
│   └── favicon-dark.svg     # Favicon for dark mode
├── components.json          # shadcn/ui configuration
├── vite.config.ts
├── vitest.config.ts         # Test configuration
└── package.json
```

### Frontend Patterns

- **Adding shadcn components:** `pnpm dlx shadcn@latest add <component>`
- **Import alias:** Use `@/` for imports from `src/` (e.g., `@/components/ui/button`)
- **API calls:** Use hooks from `@/hooks/use-api.ts`, not raw fetch
- **State:** Use `useAppStore()` from `@/stores/app-store.ts`

### Data Flow

```
Banner API        → Scraper  ─┐
registrar.wwu.edu → Calendar ─┴→ SQLite → Cache (in-memory) → API → Frontend
                       ↑
          Jobs Service (scheduled scraping)
```

### Key Design Decisions

- **SQLite with WAL mode** for concurrent reads during scraping
- **golang-migrate** for schema migrations
- **sqlc** for type-safe queries (regenerate after changing queries.sql)
- **Separate concerns:**
  - `store/` - sqlc queries and models
  - `cache/` - in-memory cache for schedule generation: a term loads on its first request as an immutable snapshot and is dropped after a scrape, the calendar job, or the grade import rewrites its data, so the next request reloads it

## External Data Sources

WWU Banner API requires session initialization:
1. GET `/classSearch/getTerms` (gets cookies)
2. POST `/term/search` with term (sets context, needs ~1s to process)
3. GET `/searchResults/searchResults` (paginated course data, 500 per page)

Additional endpoints:
- GET `/classSearch/get_subject?term=YYYYTT&max=1000` - Returns subject codes with descriptions
  ```json
  [{"code": "CSCI", "description": "Computer Science"}, ...]
  ```

See `internal/scraper/scraper.go` package doc for detailed design decisions (timeouts, concurrency, partial failure handling).

Term dates, registration phases, deadlines, holidays, and the finals schedule come from registrar.wwu.edu, since Banner exposes none of them. See the `internal/calendar` package doc for the pages and how they are parsed.

The API serves calendar data as JSON only: `GET /api/terms` adds each term's phase and dates, `GET /api/calendar/:term` returns its registration phases, holidays, deadlines, and finals week, and section payloads (generate, search, CRN lookup) carry an optional `final`. Calendar export is an `.ics` file built in the browser (`lib/ics.ts`); there are no hosted feeds, so nothing on the server has to track subscribers or stay in sync with a schedule the server never stores.

## Environment Variables

```bash
PORT=48920
ENVIRONMENT=development
DATABASE_PATH=data/schedule.db
CORS_ALLOWED_ORIGINS=http://localhost:5173
SCRAPER_CONCURRENCY=4           # concurrent page fetches (default 4)

# Jobs Service
JOBS_ENABLED=true               # enable background job scheduler (default true)
JOBS_ACTIVE_SCRAPE_HOURS=8      # hours between scrapes of terms in registration (default 8)
JOBS_DAILY_SCRAPE_HOUR=3        # Pacific hour (0-23) for daily pre-registration and in-session scrapes and the calendar refresh (default 3)
JOBS_LOG_RETENTION_DAYS=90      # days to keep logs before pruning (default 90, TODO: not implemented)
JOBS_PAST_TERM_YEARS=5          # years of past terms to scrape (default 5)
```

## Open Work

Tracked in GitHub issues: run `gh issue list`. Issue numbers are not copied here because the list goes stale as issues close.

## Testing Strategy

- **Unit tests** for all service functions
- **Integration tests** for API endpoints
- **Benchmark tests** for:
  - Schedule generation algorithm
  - Search/filter operations
  - Cache loading performance
  - Database query performance
- **Table-driven tests** preferred for Go code

## Go Style (Modern Patterns)

Use modern Go idioms. Run `go fix ./...` to apply the modernizers (`go fix -diff ./...` to preview).

**Avoid redundant comments:**

```go
// BAD: Comment restates the function name
// Get sections with primary instructor
sections, err := c.queries.GetSectionsWithInstructorByTerm(ctx, term)

// GOOD: No comment needed - function name is self-documenting
sections, err := c.queries.GetSectionsWithInstructorByTerm(ctx, term)

// GOOD: Comment explains WHY, not WHAT
// Pre-fetch all meetings to avoid N+1 queries when building course objects
meetings, err := c.queries.GetMeetingTimesByTerm(ctx, term)
```

**Avoid section headers in queries.sql** - they get copied into generated Go code and break intellisense:

```sql
-- BAD: This becomes a doc comment for the next function in generated code
-- ============================================================================
-- Schedule Cache Queries
-- ============================================================================

-- name: GetSectionsWithInstructorByTerm :many
SELECT ...

-- GOOD: Just the query, sqlc's "-- name:" comment is sufficient
-- name: GetSectionsWithInstructorByTerm :many
SELECT ...
```

**Avoid these legacy patterns:**

```go
// BAD: Legacy loop
for i := 0; i < n; i++ { ... }
// GOOD: Range over int (Go 1.22+)
for range n { ... }
// Or if you need the index:
for i := range n { ... }

// BAD: Legacy benchmark loop
for i := 0; i < b.N; i++ { ... }
// GOOD: b.Loop() (Go 1.24+)
for b.Loop() { ... }

// BAD: Manual slice search
found := false
for _, v := range slice {
    if v == target { found = true; break }
}
// GOOD: slices.Contains (Go 1.21+)
found := slices.Contains(slice, target)

// BAD: interface{}
func foo(x interface{}) { ... }
// GOOD: any (Go 1.18+)
func foo(x any) { ... }
```
