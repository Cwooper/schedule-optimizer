# Stats Dashboard - Research & Planning Notes

## Current State (Feb 11, 2026)

### Backend Grade Service: Complete
- Branch: `feat/gpa`
- Excel import, CRN-join mapping, 4-level aggregation, in-memory lookups all working
- GPA/GPASource fields already flow through cache, search, and generator API responses
- Frontend doesn't consume them yet

### Coverage Stats (after HTML entity fix)

| Level | Total | Matched | Coverage | Notes |
|-------|-------|---------|----------|-------|
| Subjects | 117 | 114 | 97.4% | 3 missing (ITAL, PORT, RDG) - not in grade data |
| Instructors (in grade range) | 2,486 | 2,262 | 90.9% | Unmapped are mostly MUS applied lesson profs |
| Courses (recent terms) | 3,359 | 2,612 | 77.8% | Missing are new courses, MUS lessons, etc. |
| Course+Prof combos (recent) | 7,036 | 4,738 | 67.3% | +48 recovered by HTML entity fix |

Remaining gaps are expected: S/U-only courses (internships, practica), MUS individual lessons, new course numbers post-2025, new hires.

### Bug Fixed This Session
`mapInstructor` wasn't unescaping HTML entities before map lookup. Banner stores names like `O&#39;Neil` and `Monta&ntilde;o`, but the mapping table stores clean names (`O'Neil`, `Montaño`). Fix: apply `html.UnescapeString` in `mapInstructor` before lookup. Affected 21 instructors / 935 sections.

---

## Interesting Data Findings

### Hardest Courses (letter-graded, 100+ students)
| Course | GPA | Students | A Rate | F/W Rate |
|--------|-----|----------|--------|----------|
| PHIL 201 | 1.73 | 498 | 16.1% | 46.4% |
| MATH 112 | 2.25 | 12,900 | 17.0% | 16.8% |
| MATH 114 | 2.31 | 10,937 | 16.3% | 18.9% |

### CSCI Weed-Out Gauntlet
- CSCI 103 (2.64) → 145 (2.83) → 247 (2.77) → 301 (2.78) — all cluster ~2.8
- Senior capstone (491/492/493) jumps to 3.7+ — "you made it, here's your A"

### Biggest Professor Spreads (same course, wildly different outcomes)
| Course | GPA Spread | Lowest | Highest | # Profs |
|--------|-----------|--------|---------|---------|
| MATH 114 | 2.27 | 1.30 | 3.57 | 90 |
| ANTH 201 | 1.87 | 2.01 | 3.88 | 17 |
| CSCI 301 | 1.46 | 2.29 | 3.75 | 13 |
| PSY 101 | 1.26 | 1.77 | 3.03 | 10 |

### "Did Anybody Even Get an A?"
- HIST 333: 0 straight A's out of 106 students (0.9% A/A- rate)
- LANG 305: 3.5% A/A- rate (3.67 GPA though — everyone gets B+/A-)

### Department Difficulty
- Hardest: HIST (2.98), LBRL (3.00), ECON/MATH (3.04)
- CSCI: 3.27 (mid-pack)
- Easiest: SEC (3.86), ELIT (3.85), HNRS (3.81)

---

## Dashboard Design Direction

### Recommended Layout: Course Explorer + Highlights

**Two areas on the stats tab:**

1. **Highlights/Leaderboards** (engagement hook, landing state)
   - Hardest courses, easiest courses, biggest professor spreads
   - "Did anybody get an A?" list
   - Department comparison
   - NOT professor-specific leaderboards (too controversial)

2. **Course Explorer** (the practical tool)
   - Subject dropdown → sortable course list with GPA badges
   - Click course → full grade distribution bar chart + professor comparison table
   - Professor comparison is contextual (within a course), not standalone

### User Scenarios
1. "How hard is this class?" → Course GPA + grade distribution
2. "Which professor should I pick?" → Professor comparison within a course
3. "Did anybody even get an A?!" → Raw grade counts, percentages
4. "What are the easy electives in CSCI?" → Browse courses in a subject

### Tech Stack
- shadcn/ui components: Card, Select/Combobox, Table, Badge
- shadcn charts (Recharts) for grade distribution bar charts
- `pnpm dlx shadcn@latest add chart` to install

### Implementation Order
1. Backend stats API endpoints (wire stats.Service into handlers)
2. Frontend: install chart components, add API types for GPA fields
3. Course explorer (subject → course → distribution + professor table)
4. Highlights/leaderboards

---

## Code Review Items Completed This Session

All from the initial code review of the grade data feature:

1. **Dead code removed** - `import.go:82-83` no-op professor check
2. **O(n^2) → O(n)** - `response.go` course GPA computation (two-pass with accumulator)
3. **Transaction wrapping** - `importExcel`, `computeMappings`, `computeAggregates` all wrapped in SQLite transactions (atomicity + major perf improvement)
4. **SQL WHERE clause fix** - `GetGradeBannerJoinData` no longer excludes empty-professor rows from subject mapping evidence
5. **Tests added** - 24 test cases for grades package (computeGPA, computePassRate, pickBest, lookups, mappings, idempotency, LoadFromDB)
6. **HTML entity bug fix** - `mapInstructor` now unescapes HTML entities before map lookup
