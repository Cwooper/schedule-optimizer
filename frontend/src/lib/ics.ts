// iCalendar (RFC 5545) export, written by hand: the `ics` package emits UTC
// or floating times and no VTIMEZONE, and weekly classes need TZID-local
// times so they stay at 9:00 across the November DST change.

import type { HydratedSection, MeetingTime, RegistrationPhaseName, TermCalendar } from "@/lib/api"
import { addDays, parseISODate, weekday } from "@/lib/dates"
import { FINAL_GUESS_REASONS } from "@/lib/finals"
import { hashString } from "@/lib/schedule-utils"
import { stripTrailingNote } from "@/lib/utils"

export interface CalendarExportOptions {
  classes: boolean
  finals: boolean
  holidays: boolean
  registration: Record<RegistrationPhaseName, boolean>
  deadlines: boolean
}

export const DEFAULT_CALENDAR_EXPORT: CalendarExportOptions = {
  classes: true,
  finals: true,
  holidays: true,
  registration: { I: false, II: false, III: false },
  deadlines: false,
}

export interface IcsInput {
  termName: string
  calendar: TermCalendar
  sections: HydratedSection[]
  options: CalendarExportOptions
  now: Date
}

const TZID = "America/Los_Angeles"
const UID_DOMAIN = "schedule.cwooper.dev"
const BYDAY = ["SU", "MO", "TU", "WE", "TH", "FR", "SA"]
const DAY_LETTERS = ["U", "M", "T", "W", "R", "F", "S"]
const EXPORTED_DEADLINES = new Set(["add", "drop", "withdraw"])

// The US rule since 2007, matching the DST rules of the zone the events name.
const VTIMEZONE = [
  "BEGIN:VTIMEZONE",
  `TZID:${TZID}`,
  "BEGIN:DAYLIGHT",
  "TZOFFSETFROM:-0800",
  "TZOFFSETTO:-0700",
  "TZNAME:PDT",
  "DTSTART:19700308T020000",
  "RRULE:FREQ=YEARLY;BYMONTH=3;BYDAY=2SU",
  "END:DAYLIGHT",
  "BEGIN:STANDARD",
  "TZOFFSETFROM:-0700",
  "TZOFFSETTO:-0800",
  "TZNAME:PST",
  "DTSTART:19701101T020000",
  "RRULE:FREQ=YEARLY;BYMONTH=11;BYDAY=1SU",
  "END:STANDARD",
  "END:VTIMEZONE",
]

export function icsFilename(termName: string): string {
  return `${termName.toLowerCase().replace(/[^a-z0-9]+/g, "-")}.ics`
}

/** Whether buildIcs would emit any event. */
export function hasIcsEvents(input: Omit<IcsInput, "now">): boolean {
  return buildEvents({ ...input, now: new Date(0) }).length > 0
}

export function buildIcs(input: IcsInput): string {
  const events = buildEvents(input)
  const timed = events.some((e) => e.some((line) => line.includes(`;TZID=${TZID}`)))
  const lines = [
    "BEGIN:VCALENDAR",
    "VERSION:2.0",
    `PRODID:-//Schedule Optimizer//${UID_DOMAIN}//EN`,
    "CALSCALE:GREGORIAN",
    "METHOD:PUBLISH",
    ...(timed ? VTIMEZONE : []),
    ...events.flat(),
    "END:VCALENDAR",
  ]
  return lines.map(foldLine).join("\r\n") + "\r\n"
}

type Event = string[]
type MakeEvent = (uid: string, lines: string[]) => Event

function buildEvents({ termName, calendar, sections, options, now }: IcsInput): Event[] {
  const stamp = utcStamp(now)
  const event: MakeEvent = (uid, lines) => [
    "BEGIN:VEVENT",
    `UID:${calendar.term}-${uid}@${UID_DOMAIN}`,
    `DTSTAMP:${stamp}`,
    ...lines,
    "END:VEVENT",
  ]
  return [
    ...(options.classes ? sections.flatMap((s) => classEvents(event, s, calendar)) : []),
    ...(options.finals ? sections.flatMap((s) => finalEvents(event, s)) : []),
    ...dateEvents(event, termName, calendar, options),
  ]
}

function classEvents(event: MakeEvent, section: HydratedSection, calendar: TermCalendar): Event[] {
  return section.meetingTimes.flatMap((meeting) => {
    const lines = classEventLines(section, meeting, calendar)
    if (!lines) return []
    const days = meeting.days.map((on, i) => (on ? DAY_LETTERS[i] : "")).join("")
    return [event(`${section.crn}-${days}-${meeting.startTime}`, lines)]
  })
}

function finalEvents(event: MakeEvent, section: HydratedSection): Event[] {
  const final = section.final
  if (!final) return []
  const mark = final.certain ? "" : "? "
  const summary = `${mark}${section.subject} ${section.courseNumber} - Final Exam`
  return [
    event(`${section.crn}-final`, [
      `DTSTART;TZID=${TZID}:${localStamp(final.date, final.start)}`,
      `DTEND;TZID=${TZID}:${localStamp(final.date, final.end)}`,
      `SUMMARY:${escapeText(summary)}`,
      ...(final.reason ? [`DESCRIPTION:${escapeText(FINAL_GUESS_REASONS[final.reason])}`] : []),
    ]),
  ]
}

/** WWU's own dates: holidays, registration openings, and deadlines. */
function dateEvents(
  event: MakeEvent,
  termName: string,
  calendar: TermCalendar,
  options: CalendarExportOptions
): Event[] {
  const holidays = options.holidays ? calendar.holidays : []
  const phases = calendar.registration.filter((r) => options.registration[r.phase])
  const deadlines = options.deadlines
    ? calendar.deadlines.filter((d) => EXPORTED_DEADLINES.has(d.kind))
    : []
  return [
    ...holidays.map((h) =>
      allDay(event, `holiday-${h.start}-${hashString(h.name)}`, h.start, h.end, stripTrailingNote(h.name))
    ),
    ...phases.map((r) =>
      allDay(event, `registration-${r.phase}`, r.start, r.start, `${termName} Phase ${r.phase} registration opens`)
    ),
    ...deadlines.map((d) =>
      allDay(event, `deadline-${d.date}-${hashString(d.name)}`, d.date, d.date, stripTrailingNote(d.name))
    ),
  ]
}

function allDay(
  event: MakeEvent,
  uid: string,
  start: string,
  end: string,
  summary: string
): Event {
  return event(uid, [
    `DTSTART;VALUE=DATE:${compactDate(start)}`,
    `DTEND;VALUE=DATE:${compactDate(addDays(end, 1))}`,
    `SUMMARY:${escapeText(summary)}`,
    "TRANSP:TRANSPARENT",
  ])
}

/**
 * A weekly class from its first meeting on or after the term starts through
 * the day before finals week (or its own end date if earlier), skipping
 * holidays. Returns null for an unscheduled meeting or an empty range.
 */
function classEventLines(
  section: HydratedSection,
  meeting: MeetingTime,
  calendar: TermCalendar
): string[] | null {
  const days = meeting.days.flatMap((on, i) => (on ? [i] : []))
  if (!meeting.startTime || !meeting.endTime || days.length === 0) return null

  const from = maxDate(meeting.startDate ?? calendar.start, calendar.start)
  const lastClassDay = calendar.finals ? addDays(calendar.finals.start, -1) : calendar.end
  const until = minDate(meeting.endDate ?? lastClassDay, lastClassDay)
  let first = from
  while (!days.includes(weekday(first)) && first <= until) first = addDays(first, 1)
  if (first > until) return null

  const excluded = new Set<string>()
  for (const h of calendar.holidays) {
    for (let day = maxDate(h.start, first); day <= minDate(h.end, until); day = addDays(day, 1)) {
      if (days.includes(weekday(day))) excluded.add(localStamp(day, meeting.startTime))
    }
  }

  const location = [meeting.building, meeting.room]
    .filter((part) => part && part !== "ASNC" && part !== "SYNC")
    .join(" ")
  const description = [`CRN ${section.crn}`, section.instructor].filter(Boolean).join("\n")
  const summary = `${section.subject} ${section.courseNumber} - ${section.title}`
  return [
    `DTSTART;TZID=${TZID}:${localStamp(first, meeting.startTime)}`,
    `DTEND;TZID=${TZID}:${localStamp(first, meeting.endTime)}`,
    `RRULE:FREQ=WEEKLY;BYDAY=${days.map((d) => BYDAY[d]).join(",")};UNTIL=${pacificEndOfDayUTC(until)}`,
    ...[...excluded].sort().map((stamp) => `EXDATE;TZID=${TZID}:${stamp}`),
    `SUMMARY:${escapeText(summary)}`,
    ...(location ? [`LOCATION:${escapeText(location)}`] : []),
    `DESCRIPTION:${escapeText(description)}`,
  ]
}

function minDate(a: string, b: string): string {
  return a < b ? a : b
}

function maxDate(a: string, b: string): string {
  return a > b ? a : b
}

function compactDate(iso: string): string {
  return iso.replaceAll("-", "")
}

/** "2026-09-23", "0900" -> "20260923T090000" */
function localStamp(iso: string, hhmm: string): string {
  return `${compactDate(iso)}T${hhmm.padStart(4, "0")}00`
}

function utcStamp(date: Date): string {
  return date.toISOString().replace(/[-:]/g, "").replace(/\.\d{3}/, "")
}

function nthSunday(year: number, month: number, n: number): number {
  const firstDay = new Date(year, month, 1).getDay()
  return 1 + ((7 - firstDay) % 7) + (n - 1) * 7
}

/** Whether 23:59 on a Pacific date is in daylight time (US rule since 2007). */
export function isPacificDaylightEvening(iso: string): boolean {
  const date = parseISODate(iso)
  const y = date.getFullYear()
  const start = new Date(y, 2, nthSunday(y, 2, 2))
  const end = new Date(y, 10, nthSunday(y, 10, 1))
  return date >= start && date < end
}

/**
 * RFC 5545 requires UNTIL in UTC when DTSTART carries a TZID; this is the
 * last second of the Pacific day, so a class that evening still counts.
 */
function pacificEndOfDayUTC(iso: string): string {
  const [y, m, d] = iso.split("-").map(Number)
  const offsetHours = isPacificDaylightEvening(iso) ? 7 : 8
  return utcStamp(new Date(Date.UTC(y, m - 1, d, 23 + offsetHours, 59, 59)))
}

export function escapeText(text: string): string {
  return text
    .replace(/\\/g, "\\\\")
    .replace(/;/g, "\\;")
    .replace(/,/g, "\\,")
    .replace(/\r?\n/g, "\\n")
}

const encoder = new TextEncoder()

/** Folds a content line at 75 octets without splitting a UTF-8 character. */
export function foldLine(line: string): string {
  const parts: string[] = []
  let current = ""
  let octets = 0
  for (const char of line) {
    const size = encoder.encode(char).length
    // Continuation lines spend one octet on their leading space.
    const limit = parts.length === 0 ? 75 : 74
    if (octets + size > limit) {
      parts.push(current)
      current = ""
      octets = 0
    }
    current += char
    octets += size
  }
  parts.push(current)
  return parts.join("\r\n ")
}
