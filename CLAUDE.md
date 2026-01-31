# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Schedule Optimizer is a course scheduling tool for WWU students. The main branch contains the v1 implementation. The v2 branch is a full rewrite:

- **Backend:** Go (Gin + SQLite + sqlc)
- **Frontend:** React 19 + TypeScript + Vite + Tailwind v4
  - **UI Components:** shadcn/ui (new-york style, neutral base)
  - **State Management:** Zustand (persisted to localStorage)
  - **Data Fetching:** TanStack Query
  - **No routing:** Tab state stored in localStorage, defaults to "schedule"

## Development Philosophy

- **Plan before implementing.** Discuss architecture and design decisions before writing code. Use plan mode for non-trivial features.
- **Test-heavy development.** Write tests alongside implementation. Benchmark tests are critical for schedule generation and scraping performance.
- **Code review on completion.** After implementing a task, spin off an agent to perform an unbiased code review of the changes. Code reviews must check:
  - Architecture alignment (does this belong here?)
  - File and function lengths (keep functions focused, split large files)
  - Scope creep (only implement what was requested)
  - Are the test cases indicative of the actual code, or are they mocking unrealistic scenarios
- **Course-centric, not term-centric.** Data structures should support cross-term search. Organize by unique course ID, not term→CRN hierarchy.
- **No autonomous commits.** Do not create commits or sign commits. The user will decide when and what to commit.

## Prerequisites

Install these on a fresh server:

```bash
# Go 1.24+
# https://go.dev/doc/install

# SQLite3
sudo apt install sqlite3

# sqlc (SQL code generator)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# golang-migrate (database migrations)
# With sqlite3 support - see https://github.com/golang-migrate/migrate
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Frontend (when implemented)
# Node.js 20+ and pnpm
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
│   │   ├── response.go       # Banner API response types (TODO: validate against real API)
│   │   └── store.go          # Persistence helpers
│   ├── jobs/                 # Background job scheduler
│   │   ├── service.go        # Job interface + Service (runs jobs on interval)
│   │   ├── scrape.go         # BootstrapJob, PastTermBackfillJob, ActiveScrapeJob, DailyScrapeJob
│   │   └── term.go           # Term phase detection (40-day registration estimate)
│   ├── cache/                # In-memory schedule cache
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
Banner API → Scraper → SQLite → Cache (in-memory) → API → Frontend
     ↑
Jobs Service (scheduled scraping)
```

### Key Design Decisions

- **SQLite with WAL mode** for concurrent reads during scraping
- **golang-migrate** for schema migrations
- **sqlc** for type-safe queries (regenerate after changing queries.sql)
- **Separate concerns:**
  - `store/` - sqlc queries and models
  - `cache/` - in-memory cache for schedule generation (active terms only)

## External Data Sources

WWU Banner API requires session initialization:
1. GET `/classSearch/getTerms` (gets cookies)
2. POST `/term/search` with term (sets context, needs ~1s to process)
3. GET `/searchResults/searchResults` (paginated course data, 500 per page)

See `internal/scraper/scraper.go` package doc for detailed design decisions (timeouts, concurrency, partial failure handling). See `WWU-Scraping-URLs.md` for endpoint documentation.

**TODO:** Validate `internal/scraper/response.go` types against real Banner API responses before production use.

## Environment Variables

```bash
PORT=48920
ENVIRONMENT=development
DATABASE_PATH=data/schedule.db
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
SCRAPER_CONCURRENCY=4           # concurrent page fetches (default 4)

# Jobs Service
JOBS_ENABLED=true               # enable background job scheduler (default true)
JOBS_ACTIVE_SCRAPE_HOURS=8      # hours between active registration scrapes (default 8)
JOBS_DAILY_SCRAPE_HOUR=3        # hour (0-23) for daily pre-registration scrapes (default 3)
JOBS_LOG_RETENTION_DAYS=90      # days to keep logs before pruning (default 90, TODO: not implemented)
JOBS_PAST_TERM_YEARS=5          # years of past terms to scrape (default 5)
```

## v2 Roadmap (GitHub Issues)

### Completed
- [x] #9 - Frontend Mockup/Wireframes
- [x] #10 - Design Schedule Optimizer Icon/Logo

### Backend
- [x] #13 - Set up new Go Backend with Gin
- [x] #14 - Implement SQLite Database Layer
- [x] #15 - Implement Structured Logging with slog *(tint for colored dev logs)*
- [ ] #16 - Implement Analytics/Statistics Collection
- [x] #17 - Migrate Data Pipeline to SQLite *(scraper done, needs real API validation)*
- [x] #18 - Implement Jobs Service (scheduled scraping)
- [x] #19 - Schedule Generation Logic *(bitmask-based conflict detection, 4-9x faster than v1)*
- [ ] #20 - Implement Course Search Service (NLP, advanced search)
- [ ] #21 - Backend Tests
- [ ] #34 - Explore ML Features
- [ ] #37 - Extra Dates and Times *(jobs use 40-day registration estimate, TODO: scrape actual dates)*

### Frontend
- [ ] #11 - Frontend SEO
- [ ] #22 - Frontend Initialization (Vite + TypeScript + Tailwind v4)
- [ ] #23 - Component Library
- [ ] #24 - Build API Client
- [ ] #25 - Schedule Builder
- [ ] #26 - Course Search Mode
- [ ] #27 - Implement Stats Dashboard
- [ ] #28 - Add Mobile Responsive Layout
- [ ] #32 - Custom Courses in Schedule
- [ ] #33 - Campus Map Feature

### API & Infrastructure
- [ ] #12 - Define API Contract
- [ ] #30 - v2 Documentation
- [ ] #31 - Deploy v2
- [ ] #35 - Share Schedule Feature
- [ ] #36 - Campus Map

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

Use modern Go idioms. Run `go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -test ./...` to check.

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
