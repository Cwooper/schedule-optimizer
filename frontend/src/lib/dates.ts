// The API sends calendar dates as YYYY-MM-DD. They are WWU (Pacific) dates,
// handled here as local-midnight Dates so arithmetic never crosses a zone.

const MONTHS = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"]
const WEEKDAYS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"]

export function parseISODate(iso: string): Date {
  const [y, m, d] = iso.split("-").map(Number)
  return new Date(y, m - 1, d)
}

export function toISODate(date: Date): string {
  const m = String(date.getMonth() + 1).padStart(2, "0")
  const d = String(date.getDate()).padStart(2, "0")
  return `${date.getFullYear()}-${m}-${d}`
}

export function addDays(iso: string, days: number): string {
  const date = parseISODate(iso)
  date.setDate(date.getDate() + days)
  return toISODate(date)
}

/** Whole calendar days from a to b; Math.round absorbs a DST hour. */
export function daysBetween(a: string, b: string): number {
  return Math.round((parseISODate(b).getTime() - parseISODate(a).getTime()) / 86_400_000)
}

export function weekday(iso: string): number {
  return parseISODate(iso).getDay()
}

/** "Nov 3" */
export function formatMonthDay(iso: string): string {
  const date = parseISODate(iso)
  return `${MONTHS[date.getMonth()]} ${date.getDate()}`
}

/** "Tue, Nov 3" */
export function formatWeekdayDate(iso: string): string {
  return `${WEEKDAYS[weekday(iso)]}, ${formatMonthDay(iso)}`
}

/** "Tue" */
export function formatWeekday(iso: string): string {
  return WEEKDAYS[weekday(iso)]
}

/** "12/7" */
export function formatNumericDate(iso: string): string {
  const date = parseISODate(iso)
  return `${date.getMonth() + 1}/${date.getDate()}`
}

/** "Dec 7-11", or "Nov 30-Dec 4" across months. */
export function formatDayRange(start: string, end: string): string {
  if (start === end) return formatMonthDay(start)
  const s = parseISODate(start)
  const e = parseISODate(end)
  if (s.getMonth() === e.getMonth()) return `${formatMonthDay(start)}-${e.getDate()}`
  return `${formatMonthDay(start)}-${formatMonthDay(end)}`
}

/** "Tue, Nov 3 - Mon, Nov 16", or one date when start is end. */
export function formatWeekdayRange(start: string, end: string): string {
  if (start === end) return formatWeekdayDate(start)
  return `${formatWeekdayDate(start)} - ${formatWeekdayDate(end)}`
}
