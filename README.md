# Schedule Optimizer

A course scheduling tool for Western Washington University students. Add your courses, set preferences, and generate optimized, conflict-free schedules instantly.

**[Try it live](https://cwooper.me/schedule-optimizer)**

> **Disclaimer:** This project is not affiliated with Western Washington University. It is an independent initiative developed for educational and personal use. All data is for informational purposes only and should not be considered official or binding.

## Features

### Schedule Generation

- Add courses by subject + number (e.g., "CSCI 247") or by CRN
- Mark courses as **required** or **optional** with configurable min/max bounds
- **Pin specific sections** to lock in a preferred CRN
- **Blocked times** — paint time slots as unavailable with named groups, custom colors, opacity, and hatching patterns
- Browse generated schedules with scoring based on gaps, start time, and end time preferences
- View async/TBD courses separately

Under the hood, schedule generation uses a **bitmask-based conflict detection** algorithm with backtracking and pruning. Each section is encoded as a 450-bit time mask (5 days x 90 ten-minute slots from 7am-10pm), enabling O(1) conflict checks. On a modern CPU, conflict detection runs in ~3ns and a full 10-course generation completes in under 2ms.

### Course Search

- Search across a single term, an academic year, or all historical data
- Filter by subject, course number, title, instructor, open seats, and credit range
- Results scored by relevance and term recency
- Add courses or specific CRNs directly from search results to the schedule builder

### Data Pipeline

- Automated scraping from WWU's Banner API with concurrent page fetches
- Background job scheduler with four job types: bootstrap, historical backfill, active registration polling, and daily pre-registration scrapes
- Course-centric data model supporting cross-term search — organized by unique course identity, not term hierarchy
- SQLite with WAL mode for concurrent reads during scraping

### Frontend

- Dark/light theme with system preference detection
- Mobile-responsive layout with drawer navigation
- Export schedules as PNG (with native share API on mobile)
- Announcement banners and in-app feedback submission
- All state persisted to localStorage (courses, filters, schedules, preferences)

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.24, Gin, SQLite (WAL mode) |
| SQL | sqlc (type-safe codegen), golang-migrate |
| Frontend | React 19, TypeScript, Vite |
| Styling | Tailwind CSS v4, shadcn/ui |
| State | Zustand (persisted), TanStack Query |
| Testing | Go test + benchmarks, Vitest |
| Logging | slog + tint |

## Architecture

```
Banner API → Scraper → SQLite → Cache (in-memory) → REST API → React Frontend
     ↑
Jobs Service (scheduled scraping)
```

### Backend

```
backend/
├── cmd/server/          # Entry point
├── migrations/          # SQL migration files (golang-migrate)
├── internal/
│   ├── api/             # HTTP handlers
│   ├── cache/           # In-memory schedule cache (active terms)
│   ├── config/          # Environment configuration
│   ├── db/              # SQLite connection setup
│   ├── generator/       # Bitmask conflict detection + backtracking
│   ├── jobs/            # Background job scheduler (scrape scheduling)
│   ├── scraper/         # Banner API client + data extraction
│   ├── search/          # Course search with scoring
│   ├── server/          # Router, middleware, graceful shutdown
│   ├── static/          # Embedded frontend serving
│   └── store/           # sqlc-generated query code
├── Makefile
└── sqlc.yaml
```

### Frontend

```
frontend/src/
├── components/
│   ├── ui/                  # shadcn/ui primitives
│   ├── schedule/            # Grid, time blocks, course dialogs, blocked times
│   └── schedule-builder/    # Course input, rows, previews
├── hooks/                   # API hooks, theme, export, drag-to-paint
├── stores/                  # Zustand store (persisted to localStorage)
└── lib/                     # API client, schedule utils, warnings
```

### API

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/health` | Health check |
| `GET` | `/api/terms` | Available academic terms |
| `GET` | `/api/subjects` | Subject codes (optionally by term) |
| `GET` | `/api/course/:subject/:courseNumber` | Course details + sections |
| `GET` | `/api/search` | Filtered course search |
| `GET` | `/api/crn/:crn` | CRN lookup |
| `POST` | `/api/courses/validate` | Batch validate courses |
| `POST` | `/api/generate` | Generate schedule combinations |
| `GET` | `/api/announcement` | Active announcement |
| `POST` | `/api/feedback` | Submit feedback |

## Getting Started

### Prerequisites

- Go 1.24+
- Node.js 20+ and pnpm
- SQLite 3
- [sqlc](https://sqlc.dev/) — `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- [golang-migrate](https://github.com/golang-migrate/migrate) — `go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`

### Backend

```bash
cd backend
cp .env.example .env        # configure environment
make migrate-up             # apply database migrations
make run                    # start server (default port 48920)
```

### Frontend

```bash
cd frontend
pnpm install
pnpm dev                    # dev server at http://localhost:5173
```

### Production Build

```bash
cd frontend && pnpm build   # compile frontend
cd ../backend && make build # builds binary with embedded frontend
./bin/server                # single binary serves everything
```

### Useful Commands

```bash
# Backend
make test                   # run all Go tests
make bench                  # run benchmarks
make sqlc                   # regenerate query code after editing queries.sql

# Frontend
pnpm lint                   # eslint
pnpm test                   # vitest
pnpm format                 # prettier
```

### Environment Variables

```bash
PORT=48920                          # server port
ENVIRONMENT=development             # development | production
DATABASE_PATH=data/schedule.db      # SQLite database location
CORS_ALLOWED_ORIGINS=http://localhost:5173
SCRAPER_CONCURRENCY=4               # parallel page fetches
JOBS_ENABLED=true                   # background scraping
JOBS_ACTIVE_SCRAPE_HOURS=8          # hours between active term scrapes
JOBS_DAILY_SCRAPE_HOUR=3            # hour (0-23) for daily scrapes
JOBS_PAST_TERM_YEARS=5              # years of historical data to backfill
```

## Testing

The project uses table-driven tests throughout. Backend test coverage spans all major packages:

```
internal/generator   — unit + benchmark (bitmask, backtracking, scoring)
internal/search      — unit + fuzz + benchmark
internal/jobs        — unit (scheduler, scrape jobs, term detection)
internal/api         — integration (handler endpoints)
internal/cache       — unit
internal/scraper     — unit
internal/db          — unit
internal/config      — unit
internal/static      — unit
```

Frontend tests cover the Zustand store, schedule utilities, and warning generation (118 tests).

## Authors

**Cooper Morgan** — Author & Maintainer
[cwooper.me](https://cwooper.me) · [@Cwooper](https://github.com/Cwooper)

**Konnor Kooi** — Previous Contributor
[konnorkooi.com](https://konnorkooi.com) · [@konnorkooi](https://github.com/konnorkooi)

Previous contributors include Rory Bates, Arne Wiseman, and Ben Huynh.

## License

MIT License — see [LICENSE](LICENSE) for details.
