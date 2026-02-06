# Schedule Grid Overhaul — Phased Plan

## Context

Current grid (`frontend/src/components/schedule/ScheduleGrid.tsx`, 213 lines) is functional but has a TODO for "styling, interactions, edge cases." This plan breaks the work into iterative phases that can be tested independently.

**Data insight (from SQLite):** 95%+ of classes start 8am-5pm, end 9am-5pm. Only 1 class at 7am, 137 classes ending at 9pm. Default visible range should be **8am-5pm**.

---

## Phase 1: Grid Visual Polish ✅

**Files:** `ScheduleGrid.tsx`, `index.css`

- Color system: `bg-{color}-500/40` light / `/35` dark with `border-l-3` anchoring
- Height-dependent content: subject+number always, title ≥70min, instructor ≥80min, building ≥50min
- Alternating day columns (`bg-muted/20`), solid hour lines, dashed half-hour lines
- Time labels: 11px with am/pm suffix, tabular-nums, centered on hour lines

---

## Phase 2: Dynamic Grid Sizing ✅

**Files:** `ScheduleGrid.tsx`, `ScheduleView.tsx`

- `computeTimeRange()` scans course blocks + blocked times, 10min padding, 30-min snap
- Default range 8am–4pm, minimum 8-hour span
- Fully dynamic `gridStartMin`/`gridEndMin`, `flex-1 overflow-auto` for viewport fitting

---

## Phase 3: Schedule Nav Bar Enhancements ✅

**Files:** `ScheduleView.tsx`, `ScheduleGrid.tsx`, `app-store.ts`, `index.css`

**Deferred:** A clickable "Schedule X of Y" stats dialog was considered but deferred — the grid and course block clicks already surface per-course details, and weight display is blocked on scorer rework (see Phase 7). Revisit if needed. Inline credits display in the nav bar is also a candidate for later.

### Completed:
- **Dropdown menu** (`MoreVertical` icon) absolutely positioned right side of nav bar
  - Menu items: Blocked Times, Download PNG, Export ICS, Share Link, Campus Map (all disabled until wired)
  - Uses shadcn `DropdownMenu` component
- **Regenerate button** (`RefreshCw` icon) absolutely positioned left side of nav bar
  - Turns amber when schedule is stale, disabled when no courses
  - Tooltip: contextual ("Course list changed...", "Regenerate schedules", "Add courses to regenerate")
  - CSS spin animation (`animate-spin-once`) on click via key-based remount
  - Replaces the old `cornerContent` stale indicator from the grid corner cell
- **Cleaned up ScheduleGrid**: removed dead `cornerContent` and `menuContent` props, unused `ReactNode` import
- **Fixed pre-existing lint issues**: `dragRef.current` assignment moved to `useEffect`, semicolon-led cast replaced with local variable
- **Added `BlockedTime` type** export to `app-store.ts` (was imported but never defined)
- **Persisted schedule state**: `generateResult`, `generatedWithParams`, and `currentScheduleIndex` now saved to localStorage so schedules survive page reload. Worst-case localStorage size ~520KB (2000 schedules × 8 courses).

---

## Phase 4: Blocked Times (Grid Toggle Mode) ✅

**Files:** `ScheduleGrid.tsx`, `ScheduleView.tsx`, `app-store.ts`, `api.ts`, `ScheduleBuilder.tsx`

### 4a. Store additions
- Add `blockedTimes: BlockedTime[]` to app store (persisted)
- Add actions: `addBlockedTime`, `removeBlockedTime`, `clearBlockedTimes`
- Include blocked times in stale detection fingerprint
- Type: `{ day: number; startTime: string; endTime: string }`

### 4b. Grid toggle mode
- When "Edit Blocked Times" is clicked from the nav bar dropdown (Phase 3):
  - Course blocks dim/fade out (opacity)
  - Existing blocked times show as solid hatched/red-tinted regions
  - Click-and-drag on empty cells paints new blocked time regions
  - Click existing blocked region to remove it
  - Snap to 10-minute or 30-minute increments (TBD)
  - A visible indicator (banner/badge) shows "Editing Blocked Times" with a Done button

### 4c. Normal mode blocked time display
- Blocked times render as subtle hatched/striped regions behind course blocks
- Semi-transparent so they don't dominate the view
- Different visual treatment from course blocks (no border, no text, just pattern)

### 4d. Wire to generate request
- Add `blockedTimes` field to frontend `GenerateRequest` type in `api.ts`
- Pass store's blocked times into the generate mutation call in `ScheduleBuilder.tsx`
- Backend already handles `BlockedTime` in `generator/types.go` and filters via bitmask

### 4e. Drag interaction implementation
- Track mousedown/touchstart on grid cells
- Calculate day and time from pointer position (reverse the positioning math)
- On mousemove/touchmove, expand selection rectangle
- On mouseup/touchend, commit as a new blocked time
- Visual feedback during drag (preview rectangle)

**Test:** Toggle into edit mode, paint blocked times, toggle back. Generate -- verify blocked time regions are excluded. Add a class that conflicts with blocked time, verify it's filtered out.

---

## Phase 5: Generation Feedback (Sonners) ✅

**Files:** `ScheduleBuilder.tsx`, `ScheduleView.tsx`

### 5a. No schedules generated
- When generation succeeds but `schedules.length === 0`, show a sonner explaining why
- Use the `courseResults` array to build a specific message:
  - "CSCI 247: all sections conflict with blocked times" (`status: "blocked"`)
  - "MATH 204: not offered this term" (`status: "not_offered"`)
  - "All combinations have time conflicts" (when courses exist but no valid combos)

### 5b. Partial results feedback
- When some courses couldn't be included, show an informational sonner
- e.g., "2 of 5 courses had no available sections — generated with remaining 3"

### 5c. Edge case messaging
- No courses added yet (already handled by disabled Generate button, but could add tooltip)
- Term not selected
- All courses are async-only (no scheduleable sections)

**Test:** Generate with a known impossible combination (conflicting required courses), verify sonner appears with meaningful message. Generate with a blocked-out course, verify specific feedback.

---

## Phase 6: Pinned Section Indicator ✅

**Files:** `ScheduleGrid.tsx`, `ScheduleView.tsx`

### 6a. Detect pinned sections
- A section is "pinned" when the user has filtered to specific CRNs in a course slot (`slot.sections !== null`)
- Pass pinned CRN set from the store into ScheduleGrid so blocks can check membership

### 6b. Visual indicator on course blocks
- Options to consider:
  - **Pin icon:** Small pin icon (lucide `Pin`) in the top-right corner of pinned course blocks
  - **Border style:** Different border treatment (e.g., solid thick border vs. normal)
  - **Badge/dot:** Small colored dot overlay
- Should be subtle enough not to clutter but noticeable enough to inform
- Needs to work at all block sizes (50min MWF through 170min labs)

### 6c. Tooltip/context
- On hover over a pinned block, tooltip could mention "Pinned section" or "Filtered to this CRN"
- Helps users understand why a specific section was chosen vs. alternatives

**Test:** Pin a specific section of a course, generate schedules, verify the pin indicator appears on that course's blocks. Verify non-pinned courses don't show the indicator.

---

## Phase 7: Scoring Weights & Preferences

**Files:** `ScheduleBuilder.tsx`, `app-store.ts`, `api.ts`, backend `generator/scorer.go`, backend `generator/types.go`

**Requires backend changes.** Current scorer uses linear start/end time functions that are essentially inverses of each other. This phase reworks scoring to use user-configurable preferences and exposes controls in the frontend.

### 7a. Backend: Rework scorer curves
- Replace linear start/end time scoring with quadratic/bell curves centered on user-specified ideal times
- Reference: v1 implementation (main branch) used quadratic curves with ideal start ~10am, ideal end ~2pm
- This makes start and end scores independent — "I want to start at 10am" and "I want to be done by 3pm" can coexist without being inverse
- Accept ideal start/end times as part of the generate request

### 7b. Backend: Expose weight configuration
- Accept user weight multipliers in the generate request (e.g., `{ gaps: 1.5, startTime: 1.0, endTime: 0.5, seats: 2.0 }`)
- Default multipliers if not provided (backward compatible)
- Allow setting ideal start/end times (default 10am / 2pm)

### 7c. Frontend: Weights button next to Generate
- Add a button (e.g., `SlidersHorizontal` icon) next to the Generate button in `ScheduleBuilder.tsx`
- Opens a popover or dialog with:
  - Sliders or inputs for each weight multiplier (gaps, start, end, seats)
  - Ideal start time picker (default 10am)
  - Ideal end time picker (default 2pm)
- Store preferences in app store (persisted)
- Pass preferences in the generate request

### 7d. Frontend: Store additions
- Add `scoringPreferences` to app store (persisted):
  - `weights: { gaps: number, startTime: number, endTime: number, seats: number }`
  - `idealStartTime: string` (e.g., "1000")
  - `idealEndTime: string` (e.g., "1400")
- Include in stale detection fingerprint
- Defaults match backend defaults

**Test:** Adjust ideal start to 8am, generate schedules, verify early-morning schedules rank higher. Set gaps weight to 0, verify gap-heavy schedules aren't penalized. Verify start and end scores are independent (not inverse).

---

## Files Modified (Summary)

| File | Phases |
|------|--------|
| `components/schedule/ScheduleGrid.tsx` | ~~1~~, ~~2~~, ~~3~~, 4, 6 |
| `components/schedule/ScheduleView.tsx` | ~~2~~, ~~3~~, 4, 6 |
| `stores/app-store.ts` | ~~3~~, 4, 7 |
| `lib/api.ts` | 4, 7 |
| `components/ScheduleBuilder.tsx` | 4, 5, 7 |
| `components/ui/dropdown-menu.tsx` | ~~3~~ (new, shadcn) |
| `index.css` | ~~3~~ |
| backend `generator/scorer.go` | 7 |
| backend `generator/types.go` | 7 |
