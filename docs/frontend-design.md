# Frontend Design Document

This document captures the UI design, state management, and API contract for the Schedule Optimizer v2 frontend.

## Layout Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│ [Logo] Schedule Optimizer                     [About] [Help] [☀/☾]  │
├─────────────────────────────────────────────────────────────────────┤
│              [ Schedule | Search | Statistics ]                     │
├─────────────┬───────────────────────────────────────────────────────┤
│   Sidebar   │                    Main Content                       │
│  (fullscreen on mobile, collapsable)                                │
└─────────────┴───────────────────────────────────────────────────────┘
```

### Header
- Logo + title (links to home/schedule)
- Navigation: About, Help pages
- Theme toggle (light/dark)

### Tab Navigation
- **Schedule**: Build and view generated schedules
- **Search**: Find courses across terms
- **Statistics**: Usage analytics (TBD)

### Sidebar (Persistent)
Visible on all tabs, slides in/out on mobile.

```
┌─────────────────────────┐
│ Quarter                 │
│ [Winter ▼] [2025 ▼]     │
├─────────────────────────┤
│ Credits                 │
│ Min: [  ]  Max: [  ]    │
├─────────────────────────┤
│ Add Course              │
│ [Subject|CRN]           │  ← Toggle
│ [CSCI ▼] [247] [+]      │  ← Subject mode
│   - or -                │
│ [CRN input   ] [+]      │  ← CRN mode
├─────────────────────────┤
│ Courses          Pin    │  ← Table header, tooltip on hover
│─────────────────────────│
│ CSCI 247          ○   × │
│ ▼ ENG 101         ●   × │
│   41328 (Smith)   ○   × │
│   41329 (Jones)   ●   × │
│ ⚠️ MATH 204 (F24) ○   × │  ← Wrong quarter for this CRN
├─────────────────────────┤
│      [ Generate ]       │
└─────────────────────────┘
```

---

## Course Slot Model

Each entry in the course list is a **slot** that the schedule generator will try to fill.

### Slot Properties

| Property | Type | Description |
|----------|------|-------------|
| `id` | string | Unique identifier (UUID) |
| `subject` | string | Subject code (e.g., "CSCI") |
| `code` | string | Course code (e.g., "247") |
| `displayName` | string | Human-readable name (might just combine "subject code" |
| `required` | boolean | Course-level pin (must include in schedule) |
| `sections` | SectionFilter[] | Optional CRN filter list |

### Section Filter Properties

| Property | Type | Description |
|----------|------|-------------|
| `crn` | string | Course Reference Number |
| `term` | string | Term this CRN belongs to |
| `instructor` | string | Instructor name (display) |
| `required` | boolean | CRN-level pin (must be this exact section) |

### Pin Semantics

The "pin" icon (● filled / ○ outline) indicates whether something is **required**.

**Course-level pin:**
- `required: false` → "Include this course if it fits"
- `required: true` → "Must have this course in every generated schedule"

**CRN-level pin (when sections are filtered):**
- `required: false` → "Consider this section as an option"
- `required: true` → "Must have this exact section"

**Hierarchical behavior:**
- Pin on course = at least one section required
- Pin on specific CRN = that exact section required
- Both can be true: "Must have ENG 101, specifically with Jones"

### Adding Courses

**Via Subject + Code:**
1. User selects subject, enters code, clicks +
2. Validate course exists for selected term (API call)
3. Create slot with all sections available
4. `sections` remains empty (no filter)

**Via CRN:**
1. User enters CRN, clicks +
2. Lookup CRN metadata (API call) → get subject, code, instructor, term
3. If slot for same course exists → add CRN to `sections` filter
4. If no slot exists → create new slot with this CRN in `sections`

### Quarter Mismatch Handling

CRNs are bound to specific terms. When sidebar quarter ≠ CRN's term:

- Show warning icon (⚠️) on the slot
- Tooltip: "This CRN is for Fall 2024, not Winter 2025"
- Slot is **excluded** from generation (grayed out)
- Generate button shows count: "Generating (2 courses excluded)"

---

## Page Designs

### Schedule Page

```
┌───────────────────────────────────────────────────┐
│            [◀]  Schedule 2/10  [▶]                │
├───────────────────────────────────────────────────┤
│                                                   │
│   Time   Mon    Tue    Wed    Thu    Fri     [⋮]  │
│  ┌─────┬──────┬──────┬──────┬──────┬──────┐  [ ]  │
│  │ 8am │      │      │      │      │      │  [ ]  │
│  │ 9am │ CSCI │      │ CSCI │      │ CSCI │  [ ]  │
│  │     │ 247  │      │ 247  │      │ 247  │  [ ]  │
│  │10am │      │ ENG  │      │ ENG  │      │       │
│  │     │      │ 101  │      │ 101  │      │       │
│  │11am │      │      │      │      │      │       │
│  └─────┴──────┴──────┴──────┴──────┴──────┘       │
│                                                   │
└───────────────────────────────────────────────────┘
```

**Right-side menu (⋮) (stretch goal):**
- Download as PNG
- Download as ICS
- Add custom course (TBD)
- View on campus map (TBD)
- Reset/clear

**Schedule navigation:**
- Shows X/Y where Y = total valid schedules
- Left/right arrows to browse
- Could add keyboard shortcuts (←/→)

### Search Page

```
┌─────────────────────────────────────────────────────────────┐
│ Subject     Code          Quarter Scope                     │
│ [All ▼]     [247    ]     (●) Quarter  ( ) Year  ( ) All    │
│                           [Winter ▼] [2025 ▼]               │
│                                                             │
│ Title            Professor           [ ] No Waitlist        │
│ [           ]    [            ]      Credits: [1]─●─●─[5]   │
├─────────────────────────────────────────────────────────────┤
│ ┌─────────────────────────────────────────────────────┐     │
│ │ CSCI 247 - Computer Systems               [+]       │     │
│ │ See-Mong Tan · 41328 · Winter 2026                  │     │
│ │ 5 cr · 18/20 seats                                  │     │
│ └─────────────────────────────────────────────────────┘     │
│ ┌─────────────────────────────────────────────────────┐     │
│ │ ENG 101 - Introduction to English         [+]       │     │
│ │ Some Professor · 41249 · Fall 2026                  │     │
│ │ 5 cr · 80/102 seats                                 │     │
│ └─────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────┘
```

**Search filters:**
- Subject dropdown (optional, "All" default)
- Code input (supports wildcards like "2*" for 200-level)
- Quarter scope: specific quarter, this academic year, or all time
- Title text search (fuzzy)
- Professor name search
- No waitlist checkbox
- Credit range slider
- Start time (earliest class start, e.g., "no classes before 9am")
- End time (latest class end, e.g., "no classes after 3pm")

**Results:**
- Card per section showing: title, professor, CRN, term, credits, seats
- Hover/click shows [+] to add to sidebar
- Clicking [+] adds CRN to current slot list

---

## API Contract

Base URL: `/api`

### GET /terms

Returns available terms for dropdowns.

**Response:**
```json
{
  "terms": [
    {"code": "202620", "name": "Summer 2026", "phase": "future"},
    {"code": "202520", "name": "Winter 2025", "phase": "active"},
    {"code": "202510", "name": "Fall 2024", "phase": "past"}
  ],
  "current": "202520"
}
```

Phase indicates: `future` (not yet open), `active` (registration open), `past` (archived).
We use the `active` quarter as the "currently selected" by default. Backend is authoritative
on the defaults and timing selection.

### GET /subjects

Returns subject list for dropdowns.

**Response:**
```json
{
  "subjects": [
    {"code": "CSCI", "name": "Computer Science"},
    {"code": "ENG", "name": "English"}
  ]
}
```

### GET /courses

Search for courses/sections.

**Query params:**
- `term` (required if scope=quarter, e.g., 202540)
- `year` (required if scope=year, e.g., 2025)
- `subject` (optional, e.g., CSCI)
- `code` (optional, supports wildcards)
- `title` (optional, fuzzy search)
- `instructor` (optional, fuzzy search)
- `no_waitlist` (optional, boolean)
- `min_credits` / `max_credits` (optional)
- `limit` (default 50)
- `offset` (default 0)

**Response:**
```json
{
  "sections": [
    {
      "crn": "41328",
      "term": "202520",
      "subject": "CSCI",
      "code": "247",
      "title": "Computer Systems",
      "credits": 5,
      "instructor": "Seo-Young Tan",
      "seats": 20,
      "enrolled": 18,
      "waitlist": 0,
      "meetings": [
        {"days": "MWF", "start": "0900", "end": "0950", "building": "CF", "room": "225"}
      ]
    }
  ],
  "total": 150,
  "hasMore": true
}
```

### GET /crn/:crn

Lookup a specific CRN. Used for validating CRN input and fetching display metadata.
Rethink this. Not sure which is better to check:
 - A. Check if section exists, if so, use it
 - B. If exists is true, then check section

**Query params:**
- `term` (optional, searches current term if omitted)

**Response (found):**
```json
{
  "exists": true,
  "section": {
    "crn": "41328",
    "term": "202520",
    "subject": "CSCI",
    "code": "247",
    "title": "Computer Systems",
    "instructor": "See-Mong Tan",
    "credits": 5
  }
}
```

**Response (not found):**
```json
{
  "exists": false,
  "section": null
}
```

### POST /generate

Generate valid schedule combinations. (Already implemented)

**Request:**
```json
{
  "term": "202520",
  "courses": [
    {
      "subject": "CSCI",
      "code": "247",
      "required": true,
      "crns": null
    },
    {
      "subject": "ENG",
      "code": "101",
      "required": true,
      "crns": ["41328", "41329"]
    }
  ],
  "minCourses": 1,
  "maxCourses": 8,
}
```

**Response:**
```json
{
  "schedules": [
    {
      xxx
      "score": 0.75
    }
  ],
  "total": 10,
  "excluded": [
    {"subject": "MATH", "code": "204", "reason": "wrong_term"}
    {"subject": "MATH", "code": "231", "reason": "asynchronous"}
  ]
}
```

---

## Backend Changes Needed

The current backend API needs updates to support the frontend model.

### /generate endpoint changes

**Current backend (types.go):**
```go
type GenerateRequest struct {
    Term         string        `json:"term"`
    Courses      []string      `json:"courses"`      // ["CSCI:247", "ENG:101"]
    ForcedCRNs   []string      `json:"forcedCrns"`   // CRNs that MUST be included
    BlockedTimes []BlockedTime `json:"blockedTimes"`
    MinCourses   int           `json:"minCourses"`
    MaxCourses   int           `json:"maxCourses"`
}
```

**Proposed changes:**
```go
type GenerateRequest struct {
    Term         string         `json:"term"`
    Courses      []CourseInput  `json:"courses"`      // Richer per-course options
    BlockedTimes []BlockedTime  `json:"blockedTimes"`
    MinCourses   int            `json:"minCourses"`
    MaxCourses   int            `json:"maxCourses"`
}

type CourseInput struct {
    Subject  string         `json:"subject"`           // "CSCI"
    Code     string         `json:"code"`              // "247"
    Required bool           `json:"required"`          // Course-level pin: must include
    Sections []SectionInput `json:"sections,omitempty"` // nil = all sections allowed
}

type SectionInput struct {
    CRN      string `json:"crn"`
    Required bool   `json:"required"`  // CRN-level pin: must be this exact section
}
```

**Example request:**
```json
{
  "term": "202520",
  "courses": [
    {
      "subject": "CSCI",
      "code": "247",
      "required": true,
      "sections": null
    },
    {
      "subject": "ENG",
      "code": "101",
      "required": true,
      "sections": [
        {"crn": "41328", "required": false},
        {"crn": "41329", "required": true}
      ]
    }
  ]
}
```
This means: "Must have CSCI 247 (any section). Must have ENG 101, and it must be section 41329 specifically (41328 is just for comparison)."

**Changes needed in generator:**
- Update `GenerateRequest` struct with new types
- Handle course-level `required` flag (vs optional "include if fits")
- Handle `sections` filtering (only consider listed CRNs)
- Handle CRN-level `required` flag (force specific section)
- Remove top-level `ForcedCRNs` (now per-section)

### New endpoints needed

**GET /api/terms** - Return available terms with phase info
- Query: `GetTerms` exists in store
- Need: Handler + response formatting with phase detection

**GET /api/subjects** - Return subject list
- Query: `GetDistinctSubjects` or `GetDistinctSubjectsByTerm` exist
- Need: Handler + include subject descriptions

**GET /api/courses** - Search sections
- Queries: Various `GetSections*` exist
- Need: New handler with filtering logic (wildcards, fuzzy search, time filters)
- Add `start_time` / `end_time` query params for time filtering

**GET /api/crn/:crn** - Lookup single CRN
- Query: `GetSectionByTermAndCRN` exists
- Need: Handler + response formatting

---

## LocalStorage Schema

Key: `scheduleOptimizer`

```typescript
interface AppState {
  // Sidebar state
  term: string                    // Selected term code
  minCredits: number | null
  maxCredits: number | null
  slots: CourseSlot[]

  // UI state
  theme: 'light' | 'dark' | 'system'
  sidebarCollapsed: boolean

  // Generated schedules (cleared on slot change)
  schedules: GeneratedSchedule[] | null
  currentScheduleIndex: number
}

interface CourseSlot {
  id: string                      // UUID
  subject: string
  code: string
  displayName: string
  required: boolean
  sections: SectionFilter[] | null  // null = all sections
}

interface SectionFilter {
  crn: string
  term: string
  instructor: string | null
  required: boolean
}

interface GeneratedSchedule {
  sections: string[]              // CRN list
  totalCredits: number
  score: number
}
```

**Persistence behavior:**
- Save on every state change (debounced)
- Load on app init
- Clear `schedules` when slots/term/credits change

---

## Component Hierarchy

```
App
├── Header
│   ├── Logo
│   ├── NavLinks (About, Help)
│   └── ThemeToggle
├── TabNav (Schedule, Search, Statistics)
├── Layout
│   ├── Sidebar
│   │   ├── TermSelector
│   │   ├── CreditRange
│   │   ├── AddCourse
│   │   │   ├── SubjectCodeInput
│   │   │   └── CrnInput
│   │   ├── CourseList
│   │   │   └── CourseSlot
│   │   │       └── SectionFilter
│   │   └── GenerateButton
│   └── MainContent
│       ├── SchedulePage
│       │   ├── ScheduleNav
│       │   ├── ScheduleGrid
│       │   └── ScheduleMenu
│       ├── SearchPage
│       │   ├── SearchFilters
│       │   └── SearchResults
│       │       └── SectionCard
│       └── StatsPage (TBD)
└── Modals/Tooltips
```

---

## Open Questions

1. **Search pagination** - Infinite scroll or page numbers?
 - Paginate
3. **Custom courses** - How do users add non-catalog courses (work, club meetings)?
4. **Mobile interaction** - How do users add courses from search on mobile (no hover)?
 - [+] button always visible
5. **Offline support** - Should the app work offline with cached data?
 - Lmao no, it doesn't even work without an api

---

## Future Features (Out of Scope for v1)

- Multi-quarter planning
- Share schedule via URL
- Campus map integration
- Export to calendar apps
- Professor ratings integration
- ML-based course recommendations
