import type {
  GenerateCourseInfo,
  GenerateSectionInfo,
  GenerateResponse,
  HydratedSection,
  ScheduleRef,
} from "./api"
import type { BlockedTimeBlock } from "@/stores/app-store"

// ── Grid constants ──────────────────────────────────────────────────────

export const DAYS = ["Mon", "Tue", "Wed", "Thu", "Fri"] as const

export const GRID = {
  TIME_COL: "2.5rem",
  TIME_COL_PX: 40,
  DEFAULT_START_MIN: 465,   // 7:45am (15min buffer so 8am label isn't clipped)
  DEFAULT_END_MIN: 975,     // 4:15pm (15min buffer so 4pm label isn't clipped)
  PADDING_MIN: 10,
  MIN_HOURS: 8,
  DAY_COUNT: 5,
  SNAP_MINUTES: 10,
  MIN_DRAG_MINUTES: 10,
  BLOCK_INSET_PX: 2,
  REM_PER_HOUR_MOBILE: 3,
  REM_PER_HOUR_DESKTOP: 4,
} as const

// ── Color palette ───────────────────────────────────────────────────────

export const COURSE_COLORS = [
  "bg-blue-500/40 dark:bg-blue-500/35 border-blue-500 text-blue-950 dark:text-blue-100",
  "bg-emerald-500/40 dark:bg-emerald-500/35 border-emerald-500 text-emerald-950 dark:text-emerald-100",
  "bg-rose-500/40 dark:bg-rose-500/35 border-rose-500 text-rose-950 dark:text-rose-100",
  "bg-orange-500/40 dark:bg-orange-500/35 border-orange-500 text-orange-950 dark:text-orange-100",
  "bg-violet-500/40 dark:bg-violet-500/35 border-violet-500 text-violet-950 dark:text-violet-100",
  "bg-teal-500/40 dark:bg-teal-500/35 border-teal-500 text-teal-950 dark:text-teal-100",
  "bg-amber-500/40 dark:bg-amber-500/35 border-amber-500 text-amber-950 dark:text-amber-100",
  "bg-fuchsia-500/40 dark:bg-fuchsia-500/35 border-fuchsia-500 text-fuchsia-950 dark:text-fuchsia-100",
]

// ── Blocked time constants ──────────────────────────────────────────────

export const BLOCKED_PRESET_COLORS = [
  { key: "slate", hex: "#64748b", label: "Slate" },
  { key: "rose", hex: "#f43f5e", label: "Rose" },
  { key: "amber", hex: "#f59e0b", label: "Amber" },
  { key: "emerald", hex: "#10b981", label: "Emerald" },
  { key: "sky", hex: "#0ea5e9", label: "Sky" },
  { key: "violet", hex: "#8b5cf6", label: "Violet" },
  { key: "orange", hex: "#f97316", label: "Orange" },
  { key: "pink", hex: "#ec4899", label: "Pink" },
]

// ── Blocked time helpers ────────────────────────────────────────────────

/** Convert 0-100 opacity percentage to a two-character hex string ("00"–"ff") */
export function opacityToHex(value: number): string {
  const clamped = Math.max(0, Math.min(100, Math.round(value)))
  const byte = Math.round((clamped / 100) * 255)
  return byte.toString(16).padStart(2, "0")
}

/** CSS gradient for the hatched pattern overlay, uses --hatch-color CSS variable */
export const HATCH_GRADIENT =
  "repeating-linear-gradient(45deg, transparent, transparent 4px, var(--hatch-color) 4px, var(--hatch-color) 8px)"

// ── Time helpers ────────────────────────────────────────────────────────

export function hashString(str: string): number {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i)
    hash = (hash << 5) - hash + char
    hash = hash & hash
  }
  return Math.abs(hash)
}

export function parseTime(time: string): number {
  if (!time) return 0
  const clean = time.replace(":", "").padStart(4, "0")
  const hours = parseInt(clean.slice(0, 2), 10) || 0
  const mins = parseInt(clean.slice(2, 4), 10) || 0
  return hours * 60 + mins
}

export function formatHour(hour: number): string {
  if (hour === 0 || hour === 12) return "12"
  return hour > 12 ? `${hour - 12}` : `${hour}`
}

export function formatAmPm(hour: number): string {
  return hour >= 12 ? "pm" : "am"
}

/** Convert total minutes (e.g. 570) to stored time format "0930", clamped to 00:00–23:50 */
export function minsToTimeStr(mins: number): string {
  const clamped = Math.max(0, Math.min(23 * 60 + 50, mins))
  const h = String(Math.floor(clamped / 60)).padStart(2, "0")
  const m = String(clamped % 60).padStart(2, "0")
  return `${h}${m}`
}

export function floorTo30(min: number): number {
  return Math.floor(min / 30) * 30
}

export function ceilTo30(min: number): number {
  return Math.ceil(min / 30) * 30
}

export function computeTimeRange(
  timeSpans: { startMin: number; endMin: number }[]
): { startMin: number; endMin: number } {
  let earliestMin = Infinity
  let latestMin = -Infinity

  for (const span of timeSpans) {
    earliestMin = Math.min(earliestMin, span.startMin)
    latestMin = Math.max(latestMin, span.endMin)
  }

  if (!isFinite(earliestMin)) {
    return { startMin: GRID.DEFAULT_START_MIN, endMin: GRID.DEFAULT_END_MIN }
  }

  let startMin = floorTo30(earliestMin - GRID.PADDING_MIN)
  let endMin = ceilTo30(latestMin + GRID.PADDING_MIN)

  startMin = Math.max(0, Math.min(startMin, GRID.DEFAULT_START_MIN))
  endMin = Math.min(24 * 60, Math.max(endMin, GRID.DEFAULT_END_MIN))

  if (endMin - startMin < GRID.MIN_HOURS * 60) {
    endMin = Math.min(24 * 60, startMin + GRID.MIN_HOURS * 60)
  }

  // Ensure hour labels at the edges aren't clipped by overflow-hidden.
  // When the range starts/ends exactly on an hour, the label (centered via
  // -translate-y-1/2) has half its height outside the container.
  const firstHour = Math.ceil(startMin / 60)
  if (firstHour * 60 === startMin) startMin = Math.max(0, startMin - 15)
  const lastHour = Math.floor(endMin / 60)
  if (lastHour * 60 === endMin) endMin = Math.min(24 * 60, endMin + 15)

  return { startMin, endMin }
}

// ── Overlap prevention ──────────────────────────────────────────────────

export interface TimeRange {
  day: number
  startMin: number
  endMin: number
}

/** Clamp a proposed range at existing block boundaries (create & resize). */
function clampRange(
  proposed: TimeRange,
  sameDayBlocks: TimeRange[],
  anchor?: number
): TimeRange | null {
  let { startMin, endMin } = proposed

  for (const block of sameDayBlocks) {
    if (endMin <= block.startMin || startMin >= block.endMin) continue

    if (anchor !== undefined) {
      if (anchor <= startMin) {
        endMin = Math.min(endMin, block.startMin)
      } else {
        startMin = Math.max(startMin, block.endMin)
      }
    } else {
      const distToStart = Math.abs(endMin - block.startMin)
      const distToEnd = Math.abs(startMin - block.endMin)
      if (distToStart <= distToEnd) {
        endMin = Math.min(endMin, block.startMin)
      } else {
        startMin = Math.max(startMin, block.endMin)
      }
    }
  }

  if (endMin - startMin <= 0) return null
  return { day: proposed.day, startMin, endMin }
}

/** Snap a proposed range to sit adjacent to overlapping blocks (move). */
function snapRange(
  proposed: TimeRange,
  sameDayBlocks: TimeRange[],
  dur: number
): TimeRange | null {
  const hasOverlap = sameDayBlocks.some(
    (b) => proposed.startMin < b.endMin && proposed.endMin > b.startMin
  )
  if (!hasOverlap) return proposed

  type Candidate = { startMin: number; endMin: number; dist: number }
  const candidates: Candidate[] = []
  const center = proposed.startMin + dur / 2

  for (const block of sameDayBlocks) {
    if (proposed.startMin >= block.endMin || proposed.endMin <= block.startMin) continue
    const aboveStart = block.startMin - dur
    if (aboveStart >= 0) {
      candidates.push({
        startMin: aboveStart, endMin: block.startMin,
        dist: Math.abs(aboveStart + dur / 2 - center),
      })
    }
    const belowEnd = block.endMin + dur
    if (belowEnd <= 24 * 60) {
      candidates.push({
        startMin: block.endMin, endMin: belowEnd,
        dist: Math.abs(block.endMin + dur / 2 - center),
      })
    }
  }

  if (candidates.length === 0) return null
  candidates.sort((a, b) => a.dist - b.dist)

  for (const c of candidates) {
    const overlapsOther = sameDayBlocks.some(
      (b) => c.startMin < b.endMin && c.endMin > b.startMin
    )
    if (!overlapsOther) {
      return { day: proposed.day, startMin: c.startMin, endMin: c.endMin }
    }
  }
  return null
}

/**
 * Clamp or snap a proposed time range to avoid overlapping with existing blocks.
 *
 * - **"clamp" mode** (create & resize): Truncates at the nearest block boundary.
 * - **"snap" mode** (move): Snaps adjacent to the overlapping block.
 */
export function clampToAvoidOverlap(
  proposed: TimeRange,
  existing: TimeRange[],
  mode: "clamp" | "snap",
  duration?: number,
  anchor?: number
): TimeRange | null {
  const sameDayBlocks = existing.filter((b) => b.day === proposed.day)
  if (sameDayBlocks.length === 0) return proposed

  if (mode === "clamp") return clampRange(proposed, sameDayBlocks, anchor)
  return snapRange(proposed, sameDayBlocks, duration ?? (proposed.endMin - proposed.startMin))
}

// ── Blocked time block utilities ─────────────────────────────────────

/** Convert BlockedTimeBlock[] to TimeRange[] for overlap utilities */
export function blocksToRanges(blocks: BlockedTimeBlock[]): TimeRange[] {
  return blocks.map((b) => ({
    day: b.day,
    startMin: parseTime(b.startTime),
    endMin: parseTime(b.endTime),
  }))
}

/** Get existing ranges excluding a specific block */
export function otherBlockRanges(blocks: BlockedTimeBlock[], excludeId: string): TimeRange[] {
  return blocksToRanges(blocks.filter((b) => b.id !== excludeId))
}

/**
 * Merge adjacent blocks on the same day within a group.
 * Two blocks are adjacent when prev.endTime === curr.startTime on the same day.
 * Merged block keeps the earlier block's ID.
 */
export function mergeAdjacentBlocks(blocks: BlockedTimeBlock[]): BlockedTimeBlock[] {
  if (blocks.length <= 1) return blocks

  const sorted = [...blocks].sort((a, b) => {
    if (a.day !== b.day) return a.day - b.day
    return parseTime(a.startTime) - parseTime(b.startTime)
  })

  const merged: BlockedTimeBlock[] = [sorted[0]]

  for (let i = 1; i < sorted.length; i++) {
    const prev = merged[merged.length - 1]
    const curr = sorted[i]

    if (prev.day === curr.day && parseTime(prev.endTime) === parseTime(curr.startTime)) {
      merged[merged.length - 1] = { ...prev, endTime: curr.endTime }
    } else {
      merged.push(curr)
    }
  }

  return merged
}

export function buildColorMap(courses: HydratedSection[]): Map<string, string> {
  const colorMap = new Map<string, string>()
  for (const course of courses) {
    const key = `${course.subject}:${course.courseNumber}`
    if (!colorMap.has(key)) {
      const colorIndex = hashString(key) % COURSE_COLORS.length
      colorMap.set(key, COURSE_COLORS[colorIndex])
    }
  }
  return colorMap
}

/**
 * Hydrate a single CRN into a full HydratedSection by joining course and section data.
 * Returns null if section or course not found.
 */
// Section info type that supports both generate and search responses (instructor optional)
export type SectionInfoLike = Omit<GenerateSectionInfo, 'instructor'> & { instructor?: string }

export function hydrateSection(
  sectionKey: string,
  courses: Record<string, GenerateCourseInfo>,
  sections: Record<string, SectionInfoLike>
): HydratedSection | null {
  const section = sections[sectionKey]
  if (!section) return null

  const course = courses[section.courseKey]
  if (!course) return null

  return {
    crn: section.crn,
    term: section.term,
    subject: course.subject,
    courseNumber: course.courseNumber,
    title: course.title,
    credits: course.credits,
    instructor: section.instructor ?? "",
    meetingTimes: section.meetingTimes,
    enrollment: section.enrollment,
    maxEnrollment: section.maxEnrollment,
    seatsAvailable: section.seatsAvailable,
    waitCount: section.waitCount,
    isOpen: section.isOpen,
    gpa: section.gpa,
    gpaSource: section.gpaSource,
    passRate: section.passRate,
  }
}

/**
 * Hydrate a schedule reference into full section data.
 * Filters out any CRNs that can't be resolved.
 */
export function hydrateSchedule(
  schedule: ScheduleRef,
  courses: Record<string, GenerateCourseInfo>,
  sections: Record<string, GenerateSectionInfo>
): {
  courses: HydratedSection[]
  score: number
  weights: { name: string; value: number }[]
} {
  const hydratedCourses: HydratedSection[] = []

  for (const crn of schedule.crns) {
    const hydrated = hydrateSection(crn, courses, sections)
    if (hydrated) {
      hydratedCourses.push(hydrated)
    }
  }

  return {
    courses: hydratedCourses,
    score: schedule.score,
    weights: schedule.weights,
  }
}

/**
 * Hydrate all async sections from a generate response.
 */
export function hydrateAsyncs(response: GenerateResponse): HydratedSection[] {
  const hydrated: HydratedSection[] = []

  for (const crn of response.asyncs) {
    const section = hydrateSection(crn, response.courses, response.sections)
    if (section) {
      hydrated.push(section)
    }
  }

  return hydrated
}

export interface GroupedCourse {
  subject: string
  courseNumber: string
  title: string
  credits: number
  firstCrn: string
  sections: {
    crn: string
    term: string
    instructor?: string
    seatsAvailable: number
    isOpen: boolean
  }[]
}

/** Group hydrated sections by course for list display. */
export function groupSectionsByCourse(
  sections: HydratedSection[]
): GroupedCourse[] {
  const grouped = new Map<
    string,
    { course: HydratedSection; sections: HydratedSection[] }
  >()

  for (const section of sections) {
    const key = `${section.subject}:${section.courseNumber}`
    const existing = grouped.get(key)
    if (existing) {
      existing.sections.push(section)
    } else {
      grouped.set(key, { course: section, sections: [section] })
    }
  }

  return Array.from(grouped.values()).map(({ course, sections }) => ({
    subject: course.subject,
    courseNumber: course.courseNumber,
    title: course.title,
    credits: course.credits,
    firstCrn: sections[0].crn,
    sections: sections.map((s) => ({
      crn: s.crn,
      term: s.term,
      instructor: s.instructor || undefined,
      seatsAvailable: s.seatsAvailable,
      isOpen: s.isOpen,
    })),
  }))
}

/**
 * Sort sections contextually based on term spread.
 *
 * Single-term (schedule builder, single-term search):
 *   open first → CRN
 *
 * Multi-term (cross-term search):
 *   active term first → open first → instructor asc (empty last) → term desc
 *
 * Mutates the array in place and returns it.
 */
export function sortSectionsByAvailability(
  sections: HydratedSection[],
  activeTerm?: string
): HydratedSection[] {
  const isMultiTerm = new Set(sections.map((s) => s.term)).size > 1

  if (!isMultiTerm) {
    return sections.sort((a, b) => {
      if (a.isOpen !== b.isOpen) return a.isOpen ? -1 : 1
      return a.crn.localeCompare(b.crn)
    })
  }

  return sections.sort((a, b) => {
    // Active term first
    if (activeTerm) {
      const aActive = a.term === activeTerm
      const bActive = b.term === activeTerm
      if (aActive !== bActive) return aActive ? -1 : 1
    }
    // Open seats first
    if (a.isOpen !== b.isOpen) return a.isOpen ? -1 : 1
    // Instructor asc, empty last
    const aInst = a.instructor || ""
    const bInst = b.instructor || ""
    if (aInst !== bInst) {
      if (!aInst) return 1
      if (!bInst) return -1
      return aInst.localeCompare(bInst)
    }
    // Term desc (most recent first)
    return b.term.localeCompare(a.term)
  })
}

// ── Term/Academic Year utilities ─────────────────────────────────────────

const QUARTER_FALL = 40
const QUARTER_WINTER = 10
const QUARTER_SPRING = 20
const QUARTER_SUMMER = 30

/**
 * Parse a term code into year and quarter.
 * Term codes are YYYYQQ where QQ is 10 (Winter), 20 (Spring), 30 (Summer), 40 (Fall).
 */
export function parseTermCode(code: string): { year: number; quarter: number } | null {
  if (code.length !== 6) return null
  const year = parseInt(code.slice(0, 4), 10)
  const quarter = parseInt(code.slice(4), 10)
  if (isNaN(year) || isNaN(quarter)) return null
  if (![QUARTER_WINTER, QUARTER_SPRING, QUARTER_SUMMER, QUARTER_FALL].includes(quarter)) {
    return null
  }
  return { year, quarter }
}

/**
 * Get the academic year for a term code.
 * Academic year N runs from Fall(N-1) through Summer(N).
 * Example: Fall 2024 (202440) → academic year 2025
 */
export function getAcademicYear(termCode: string): number | null {
  const parsed = parseTermCode(termCode)
  if (!parsed) return null
  // Fall starts the next academic year
  return parsed.quarter === QUARTER_FALL ? parsed.year + 1 : parsed.year
}

/**
 * Format academic year for display.
 * Example: 2025 → "2024-2025"
 */
export function formatAcademicYear(academicYear: number): string {
  return `${academicYear - 1}-${academicYear}`
}

/**
 * Derive unique academic years from a list of term codes.
 * Returns years in descending order (newest first).
 */
export function getAcademicYearsFromTerms(termCodes: string[]): number[] {
  const years = new Set<number>()
  for (const code of termCodes) {
    const year = getAcademicYear(code)
    if (year !== null) years.add(year)
  }
  return Array.from(years).sort((a, b) => b - a)
}
