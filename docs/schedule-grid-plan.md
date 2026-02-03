# Schedule Grid Overhaul — Phased Plan

## Context

Current grid (`frontend/src/components/schedule/ScheduleGrid.tsx`, 213 lines) is functional but has a TODO for "styling, interactions, edge cases." This plan breaks the work into iterative phases that can be tested independently.

**Data insight (from SQLite):** 95%+ of classes start 8am-5pm, end 9am-5pm. Only 1 class at 7am, 137 classes ending at 9pm. Default visible range should be **8am-5pm**.

---

## Phase 1: Grid Visual Polish

**Files:** `ScheduleGrid.tsx`, possibly `index.css`

### 1a. Color system overhaul
- Current colors use `bg-{color}-500/20` (very light wash). Make backgrounds more saturated/visible -- something like `bg-{color}-500/30` or `/35` in light mode, stronger in dark mode
- Ensure good contrast in both themes -- test all 8 colors against light and dark backgrounds
- Consider making the left border thicker (`border-l-3`) for better visual anchoring

### 1b. Course block content improvements
- Currently shows: `CSCI 247` + `Building Room`
- Better layout depending on block height:
  - **Tall blocks (50+ min):** Show subject+number, title (truncated), instructor, building/room
  - **Medium blocks (40-50 min):** Show subject+number, instructor or building
  - **Short blocks (<40 min):** Show subject+number only, maybe tooltip for details
- Use the `title` and `instructor` fields from `HydratedSection` (already available in the data)

### 1c. Grid background/lines
- Slightly darker alternating day column backgrounds for visual separation
- Clearer hour lines (currently `border-dashed border-muted` -- may be too subtle)
- Consider half-hour lines as lighter dashes

### 1d. Time label improvements
- Current labels are small (`text-xs`) and right-aligned in a 3rem column
- Ensure they align cleanly with hour lines

**Test:** Generate a schedule with 3-5 courses, verify readability in both light/dark mode. Check short (50-min MWF) and long (75-min TTh) blocks.

---

## Phase 2: Dynamic Grid Sizing

**Files:** `ScheduleGrid.tsx`, `ScheduleView.tsx`

### 2a. Compute visible time range
- Scan all course blocks to find earliest start and latest end
- Add 1-hour padding on each side
- Enforce minimum range of **8am-5pm** (covers 95%+ of classes per DB data)
- When no courses: show 8am-5pm default
- Include blocked times in the range calculation (Phase 5)

### 2b. Update grid rendering
- Replace hardcoded `START_HOUR=7` / `END_HOUR=22` with computed values
- Recalculate `HOURS` array, grid height, and block positioning based on dynamic range
- The `minHeight` style (`HOURS.length * 3rem`) adjusts automatically

### 2c. Height fitting
- Goal: grid fits in viewport without scrolling for typical schedules (8am-5pm = 9 hours)
- If range exceeds viewport, allow scrolling (keeps current `overflow-auto`)
- Consider using `vh`-based sizing or flex-grow to fill available space

**Test:** Generate schedules with different time spreads -- early morning classes, evening classes, compact midday-only schedules. Verify grid crops appropriately.

---

## Phase 3: Clickable Schedule Label + Stats Dialog

**Files:** `ScheduleView.tsx`, new `ScheduleStatsDialog.tsx`

### 3a. Make "Schedule X of Y" clickable
- Wrap the label in a button/clickable element
- On click, open a dialog showing schedule details

### 3b. Schedule Stats Dialog
- Show score breakdown from `weights` array (already returned per schedule):
  - Gap score, Start time score, End time score, Seats score
  - Visual bars or simple labeled values
- Show total credits for the schedule
- Show per-course info (subject, CRN, instructor, open/closed)
- Potentially: sort controls to reorder all schedules by a specific weight (gaps, start time, etc.)

### 3c. Sort functionality
- Allow user to re-sort the schedules array by any weight
- Current sort is by overall score -- let users prioritize e.g., "latest start" or "fewest gaps"
- Update `currentScheduleIndex` to 0 when sort changes

**Test:** Click schedule label, verify dialog shows correct weights. Sort by different criteria, verify schedule order changes.

---

## Phase 4: Three-Dots Dropdown Menu

**Files:** `ScheduleGrid.tsx` or `ScheduleView.tsx`

### 4a. Add menu button
- Three-dots icon (`MoreVertical` from lucide) positioned in the top-right corner of the grid area
- Uses shadcn `DropdownMenu` component (may need to add: `pnpm dlx shadcn@latest add dropdown-menu`)

### 4b. Menu structure (skeleton)
- Items listed but some disabled/coming-soon:
  - **Edit Blocked Times** -> toggles grid into blocked-times mode (Phase 5)
  - **Download as PNG** -> (coming soon)
  - **Export to Calendar (.ics)** -> (coming soon)
  - **Add Custom Course** -> (coming soon, links to #32)
  - **View on Campus Map** -> (coming soon, links to #33)

### 4c. Positioning
- The current `cornerContent` prop in ScheduleGrid puts content in the top-left time-label cell
- The three-dots menu should be top-right of the grid, outside the day columns
- May need to add a separate prop or position it in `ScheduleView` above the grid

**Test:** Menu opens, items render, disabled items show as grayed out.

---

## Phase 5: Blocked Times (Grid Toggle Mode)

**Files:** `ScheduleGrid.tsx`, `ScheduleView.tsx`, `app-store.ts`, `api.ts`, `ScheduleBuilder.tsx`

### 5a. Store additions
- Add `blockedTimes: BlockedTime[]` to app store (persisted)
- Add actions: `addBlockedTime`, `removeBlockedTime`, `clearBlockedTimes`
- Include blocked times in stale detection fingerprint
- Type: `{ day: number; startTime: string; endTime: string }`

### 5b. Grid toggle mode
- When "Edit Blocked Times" is clicked from three-dots menu:
  - Course blocks dim/fade out (opacity)
  - Existing blocked times show as solid hatched/red-tinted regions
  - Click-and-drag on empty cells paints new blocked time regions
  - Click existing blocked region to remove it
  - Snap to 10-minute or 30-minute increments (TBD)
  - A visible indicator (banner/badge) shows "Editing Blocked Times" with a Done button

### 5c. Normal mode blocked time display
- Blocked times render as subtle hatched/striped regions behind course blocks
- Semi-transparent so they don't dominate the view
- Different visual treatment from course blocks (no border, no text, just pattern)

### 5d. Wire to generate request
- Add `blockedTimes` field to frontend `GenerateRequest` type in `api.ts`
- Pass store's blocked times into the generate mutation call in `ScheduleBuilder.tsx`
- Backend already handles `BlockedTime` in `generator/types.go` and filters via bitmask

### 5e. Drag interaction implementation
- Track mousedown/touchstart on grid cells
- Calculate day and time from pointer position (reverse the positioning math)
- On mousemove/touchmove, expand selection rectangle
- On mouseup/touchend, commit as a new blocked time
- Visual feedback during drag (preview rectangle)

**Test:** Toggle into edit mode, paint blocked times, toggle back. Generate -- verify blocked time regions are excluded. Add a class that conflicts with blocked time, verify it's filtered out.

---

## Phase 6: Generation Feedback (Sonners)

**Files:** `ScheduleBuilder.tsx`, `ScheduleView.tsx`

### 6a. No schedules generated
- When generation succeeds but `schedules.length === 0`, show a sonner explaining why
- Use the `courseResults` array to build a specific message:
  - "CSCI 247: all sections conflict with blocked times" (`status: "blocked"`)
  - "MATH 204: not offered this term" (`status: "not_offered"`)
  - "All combinations have time conflicts" (when courses exist but no valid combos)

### 6b. Partial results feedback
- When some courses couldn't be included, show an informational sonner
- e.g., "2 of 5 courses had no available sections — generated with remaining 3"

### 6c. Edge case messaging
- No courses added yet (already handled by disabled Generate button, but could add tooltip)
- Term not selected
- All courses are async-only (no scheduleable sections)

**Test:** Generate with a known impossible combination (conflicting required courses), verify sonner appears with meaningful message. Generate with a blocked-out course, verify specific feedback.

---

## Phase 7: Pinned Section Indicator

**Files:** `ScheduleGrid.tsx`, `ScheduleView.tsx`

### 7a. Detect pinned sections
- A section is "pinned" when the user has filtered to specific CRNs in a course slot (`slot.sections !== null`)
- Pass pinned CRN set from the store into ScheduleGrid so blocks can check membership

### 7b. Visual indicator on course blocks
- Options to consider:
  - **Pin icon:** Small pin icon (lucide `Pin`) in the top-right corner of pinned course blocks
  - **Border style:** Different border treatment (e.g., solid thick border vs. normal)
  - **Badge/dot:** Small colored dot overlay
- Should be subtle enough not to clutter but noticeable enough to inform
- Needs to work at all block sizes (50min MWF through 170min labs)

### 7c. Tooltip/context
- On hover over a pinned block, tooltip could mention "Pinned section" or "Filtered to this CRN"
- Helps users understand why a specific section was chosen vs. alternatives

**Test:** Pin a specific section of a course, generate schedules, verify the pin indicator appears on that course's blocks. Verify non-pinned courses don't show the indicator.

---

## Files Modified (Summary)

| File | Phases |
|------|--------|
| `components/schedule/ScheduleGrid.tsx` | 1, 2, 4, 5, 7 |
| `components/schedule/ScheduleView.tsx` | 2, 3, 4, 7 |
| `components/schedule/ScheduleStatsDialog.tsx` | 3 (new) |
| `stores/app-store.ts` | 5 |
| `lib/api.ts` | 5 |
| `components/ScheduleBuilder.tsx` | 5, 6 |
| `index.css` | 1 (maybe) |
