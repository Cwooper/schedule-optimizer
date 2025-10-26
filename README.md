# Schedule Optimizer

Schedule Optimizer is a personal project to help students at WWU plan their
courses. You can try it live at 
[cwooper.me/schedule-optimizer](https://cwooper.me/schedule-optimizer)!

**Disclaimer:** This project is not affiliated with Western Washington University.
It is an independent initiative developed solely for educational and personal use.
All data provided is for informational purposes only and should not be considered
official or binding.

## Features

### Core Features

- Scrape Course Data from WWU
- Store sources in SQLite database with full metadata
- Merge GPA data from CSV into course records
- Update data on a schedule
- Support multiple quarters by being "course-first" in design

### Web Scraping

- Smart scraping for as-needed course data
- Stores structured course data
- Slow scraping at off times to not overload servers
- Does not update old courses
- More aggressive for current/new courses

### Schedule Generation

- Generate non-conflicting schedules using iterative tree-pruning algorithm
- Specify course code (e.g., "CSCI 247") which maps to multiple CRNs
- Force specific CRNs during generation
- Configure min/max courses in generated schedules
- (optional) minimal backend weighting by:
  - start time
  - end time
  - gaps
  - GPA
- Return multiple valid schedules

### Course Search

- Advanced Search based on course subject, code, title, professor, etc.
- Filter by quarter or search for all time (at least the data that I have)
- Ability to add courses to schedule (specific CRN)

### Analytics and Statistics

- Track search queries
- Track schedule generations
- Display popular subjects/courses

### Frontend Features

- Custom schedule additions
- Saving schedule as a png
- Browser caching for course list, settings, search, and possibly schedule

## Future Enhancements

### Campus Map

- Campus map showing class building locations
- Shortest path between classes using pre-computed building graph
- Highlight buildings for selected schedules on specific days
- Walking time estimates between back-to-back classes

### Schedule Sharing

- Share schedules via URL or JSON
- Shortened URLs

### Extra Dates and Times

- Scrape extra data for:
  - Final Exam times
  - Term Dates
  - Registration Times
  - When does registration for certain users
  - Holidays
- Could export as ICS file based on Term dates, Holidays, and Exam times
- Could extend this into notifying users when their registration opens

### Machine Learning

- Predict future classes based on past searches
- Difficult ratings derived from GPA distrubitions, variation, bias, professor
- Could implement Gemini Cloud LLM with function calling for AI-search

## Tech Stack

### Backend

- Go
- REST API with GIN
- SQLite database for course data and logs
- log/slog
- Smart Coordinator that rescrapes data
  - if process receives SIGUSER1
  - on daily 24 hour ticker
- Docker

### Frontend

- pnpm
- Vite
- React/TS
- Tailwind
- shadcn/ui
- @konnorkooi/schedule-glance
- Excalidraw for wireframing

## Proposed Layout

```sh
schedule-optimizer/
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── api/
│   │   ├── services/
│   │   ├── repository/
│   │   │   ├── database.go      # Schema + migrations
│   │   │   ├── course_repo.go
│   │   │   └── log_repo.go
│   │   ├── models/models.go
│   │   └── config/config.go
│   ├── data/                     # gitignored
│   ├── go.mod
│   └── Dockerfile
│
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── ui/              # shadcn
│   │   │   ├── schedule/
│   │   │   ├── search/
│   │   │   └── analytics/
│   │   ├── hooks/               # React Query hooks
│   │   ├── api/
│   │   │   ├── client.ts
│   │   │   └── types.ts
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── package.json
│   ├── vite.config.ts
│   └── Dockerfile
│
├── docker-compose.yml
└── README.md
```

## v1 Authors and Contributions

**Cooper Morgan** ([@cwooper](https://github.com/cwooper))

- Initial concept and design
- Backend development and architecture
- Docker containerization
- Frontend React components
- Website: [cwooper.me](https://cwooper.me)

**Konnor Kooi** ([@konnorkooi](https://github.com/konnorkooi))

- Frontend development
- Created [`schedule-glance`](https://www.npmjs.com/package/@konnorkooi/schedule-glance) npm module
- Schedule visualization components
- Website: [konnorkooi.com](https://konnorkooi.com)

Previous contributors include Rory Bates, Arne Wiseman, and Ben Huynh

## License

This project is licensed under the MIT License - see the LICENSE file for details.
