// iCalendar (RFC 5545) export, written by hand: the `ics` package emits UTC
// or floating times and no VTIMEZONE, and weekly classes need TZID-local
// times so they stay at 9:00 across the November DST change.

import type { HydratedSection, MeetingTime, RegistrationPhaseName, TermCalendar } from "@/lib/api"
import { addDays, weekday } from "@/lib/dates"
import { FINAL_GUESS_REASONS } from "@/lib/finals"
import { hashString } from "@/lib/schedule-utils"
import { stripTrailingNote, termAbbreviation } from "@/lib/utils"

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
  calendar: TermCalendar
  sections: HydratedSection[]
  options: CalendarExportOptions
  now: Date
}

const TZID = "America/Los_Angeles"
const UID_DOMAIN = new URL(import.meta.env.VITE_SITE_URL).host
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

function buildEvents({ calendar, sections, options, now }: IcsInput): Event[] {
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
    ...dateEvents(event, calendar, options),
  ]
}

/**
 * One event per meeting time. Rows that differ only by room (a lecture split
 * across PH 210 and PH 210B) are one event in both rooms, since duplicate
 * UIDs make calendar apps keep only one of them.
 */
function classEvents(event: MakeEvent, section: HydratedSection, calendar: TermCalendar): Event[] {
  const groups = new Map<string, MeetingTime[]>()
  for (const m of section.meetingTimes) {
    const days = m.days.map((on, i) => (on ? DAY_LETTERS[i] : "")).join("")
    // Every part comes from Banner, so the UID survives re-exports, and the
    // dates keep a split term's sessions apart.
    const key = [section.crn, days, m.startTime, m.endTime, m.startDate, m.endDate]
      .filter((part): part is string => !!part)
      .map((part) => part.replaceAll("-", ""))
      .join("-")
    groups.set(key, [...(groups.get(key) ?? []), m])
  }
  return [...groups].flatMap(([uid, rows]) => {
    const locations = [...new Set(rows.map(location).filter(Boolean))]
    const lab = rows.some((r) => r.lab)
    const lines = classEventLines(section, rows[0], locations.join(", "), lab, calendar)
    return lines ? [event(uid, lines)] : []
  })
}

function location(meeting: MeetingTime): string {
  return [meeting.building, meeting.room]
    .filter((part) => part && part !== "ASNC" && part !== "SYNC")
    .join(" ")
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
      ...(final.location ? [`LOCATION:${escapeText(final.location)}`] : []),
      ...(final.reason ? [`DESCRIPTION:${escapeText(FINAL_GUESS_REASONS[final.reason])}`] : []),
    ]),
  ]
}

/**
 * WWU's own dates: holidays, registration openings, and deadlines. Titles
 * are short and name the term, since these events sit among other terms'
 * in a calendar; the registrar's full wording goes in the description.
 */
function dateEvents(event: MakeEvent, calendar: TermCalendar, options: CalendarExportOptions): Event[] {
  const term = termAbbreviation(calendar.term)
  const allDay = (uid: string, start: string, end: string, title: string, description: string) =>
    event(uid, [
      `DTSTART;VALUE=DATE:${compactDate(start)}`,
      `DTEND;VALUE=DATE:${compactDate(addDays(end, 1))}`,
      `SUMMARY:${escapeText(term + " - " + title)}`,
      `DESCRIPTION:${escapeText(description)}`,
      "TRANSP:TRANSPARENT",
    ])
  const holidays = options.holidays ? calendar.holidays : []
  const phases = calendar.registration.filter((r) => options.registration[r.phase])
  const deadlines = options.deadlines
    ? calendar.deadlines.filter((d) => EXPORTED_DEADLINES.has(d.kind))
    : []
  return [
    ...holidays.map((h) =>
      allDay(`holiday-${h.start}-${hashString(h.name)}`, h.start, h.end, stripTrailingNote(h.name), h.name)
    ),
    ...phases.map((r) =>
      allDay(`registration-${r.phase}`, r.start, r.start, `Phase ${r.phase} registration`, r.name)
    ),
    ...deadlines.map((d) =>
      allDay(`deadline-${d.date}-${hashString(d.name)}`, d.date, d.date, d.title, d.name)
    ),
  ]
}

/**
 * A weekly class from its first meeting on or after the term starts through
 * the day before finals week (or its own end date if earlier), skipping
 * holidays. Returns null for an unscheduled meeting or an empty range.
 */
function classEventLines(
  section: HydratedSection,
  meeting: MeetingTime,
  location: string,
  lab: boolean,
  calendar: TermCalendar
): string[] | null {
  const { startTime, endTime } = meeting
  const days = meeting.days.flatMap((on, i) => (on ? [i] : []))
  if (!startTime || !endTime || days.length === 0) return null

  const from = maxDate(meeting.startDate ?? calendar.start, calendar.start)
  const lastClassDay = calendar.finals ? addDays(calendar.finals.start, -1) : calendar.end
  const until = minDate(meeting.endDate ?? lastClassDay, lastClassDay)
  let first = from
  while (!days.includes(weekday(first)) && first <= until) first = addDays(first, 1)
  if (first > until) return null

  const excluded = new Set<string>()
  for (const h of calendar.holidays) {
    for (let day = maxDate(h.start, first); day <= minDate(h.end, until); day = addDays(day, 1)) {
      if (days.includes(weekday(day))) excluded.add(localStamp(day, startTime))
    }
  }

  const description = [lab && "Lab", `CRN ${section.crn}`, section.instructor].filter(Boolean).join("\n")
  const summary = `${section.subject} ${section.courseNumber} - ${section.title}`
  return [
    `DTSTART;TZID=${TZID}:${localStamp(first, startTime)}`,
    `DTEND;TZID=${TZID}:${localStamp(first, endTime)}`,
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

const offsetFormat = new Intl.DateTimeFormat("en-US", { timeZone: TZID, timeZoneName: "longOffset" })

/** Minutes Pacific clocks are ahead of UTC at an instant ("GMT-07:00" is -420). */
function pacificOffsetMinutes(ms: number): number {
  const name = offsetFormat.formatToParts(ms).find((p) => p.type === "timeZoneName")?.value ?? ""
  const match = /^GMT([+-])(\d{2}):(\d{2})$/.exec(name)
  if (!match) throw new Error(`unexpected Pacific offset "${name}"`)
  const [, sign, hours, minutes] = match
  return (sign === "-" ? -1 : 1) * (Number(hours) * 60 + Number(minutes))
}

/**
 * RFC 5545 requires UNTIL in UTC when DTSTART carries a TZID; this is the
 * last second of the Pacific day, so a class that evening still counts.
 */
function pacificEndOfDayUTC(iso: string): string {
  const [y, m, d] = iso.split("-").map(Number)
  const wall = Date.UTC(y, m - 1, d, 23, 59, 59)
  // The offset at 23:59:59 standard time is the one in effect at that wall
  // time too: DST changes at 2:00, never between the two.
  const offset = pacificOffsetMinutes(wall + 8 * 3_600_000)
  return utcStamp(new Date(wall - offset * 60_000))
}

export function escapeText(text: string): string {
  return text
    .replace(/\\/g, "\\\\")
    .replace(/;/g, "\\;")
    .replace(/,/g, "\\,")
    .replace(/\r\n|\r|\n/g, "\\n")
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
