import type { HydratedSection } from "./api"
import { parseTime } from "./schedule-utils"

// Matches passingTimeMins in backend/internal/generator/scorer.go, so these
// gaps are the ones the Gaps weight scores.
const PASSING_TIME_MINS = 15

/** A time on a weekday; day 0 is Monday. */
export interface DayTime {
  day: number
  mins: number
}

export interface ScheduleStats {
  /** Earliest Mon-Fri start and latest end, first weekday on a tie; null with no timed meetings. */
  start: DayTime | null
  end: DayTime | null
  /** Weekly minutes between classes, not counting passing time. */
  gapMins: number
  /** Mean GPA of the sections with grade data, or null when none have it. */
  gpa: number | null
  gpaKnown: number
  sections: number
  open: number
  credits: number
  /** Mon-Fri, whether the day has a class. */
  days: boolean[]
}

type Span = [start: number, end: number]

function weekdaySpans(sections: HydratedSection[]): Span[][] {
  const week: Span[][] = [[], [], [], [], []]
  for (const m of sections.flatMap((s) => s.meetingTimes)) {
    if (!m.startTime || !m.endTime) continue
    const span: Span = [parseTime(m.startTime), parseTime(m.endTime)]
    for (const [day, spans] of week.entries()) {
      if (m.days[day + 1]) spans.push(span)
    }
  }
  return week
}

function summarizeDay(spans: Span[]): { first: number; last: number; gap: number } {
  spans.sort((a, b) => a[0] - b[0])
  let last = spans[0][1]
  let gap = 0
  for (const [start, end] of spans.slice(1)) {
    // Sorted by start, so last is the end of the merged block so far.
    if (start - last > PASSING_TIME_MINS) gap += start - last
    last = Math.max(last, end)
  }
  return { first: spans[0][0], last, gap }
}

export function computeScheduleStats(sections: HydratedSection[]): ScheduleStats {
  const week = weekdaySpans(sections).map((spans) => (spans.length > 0 ? summarizeDay(spans) : null))
  let start: DayTime | null = null
  let end: DayTime | null = null
  let gapMins = 0
  for (const [day, ds] of week.entries()) {
    if (!ds) continue
    gapMins += ds.gap
    if (!start || ds.first < start.mins) start = { day, mins: ds.first }
    if (!end || ds.last > end.mins) end = { day, mins: ds.last }
  }

  const gpas = sections.flatMap((s) => (s.gpa != null && s.gpa > 0 ? [s.gpa] : []))
  return {
    start,
    end,
    gapMins,
    gpa: gpas.length > 0 ? gpas.reduce((a, b) => a + b) / gpas.length : null,
    gpaKnown: gpas.length,
    sections: sections.length,
    open: sections.filter((s) => s.isOpen).length,
    credits: sections.reduce((sum, s) => sum + s.credits, 0),
    days: week.map((ds) => ds != null),
  }
}

/** Minutes as "2h 10m", dropping a zero part. */
export function formatDuration(mins: number): string {
  const h = Math.floor(mins / 60)
  const m = mins % 60
  if (h === 0) return `${m}m`
  return m === 0 ? `${h}h` : `${h}h ${m}m`
}
