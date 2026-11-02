import type { Term, TermCalendar } from "@/lib/api"
import {
  addDays,
  daysBetween,
  formatDayRange,
  formatMonthDay,
  formatWeekdayRange,
  weekday,
} from "@/lib/dates"
import { stripTrailingNote } from "@/lib/utils"

export interface TermStatus {
  text: string
  /** Set when the status is about another term, which clicking selects. */
  switchTo?: string
}

/** A newer term's registration this close is worth pointing out mid-term. */
const UPCOMING_REGISTRATION_DAYS = 14

/**
 * One line for the sidebar: the next registration date while registering,
 * finals week in session, or a newer term's registration about to open.
 * Returns null when there is nothing worth saying, or the calendar it needs
 * has not loaded.
 */
export function termStatus(
  term: Term,
  calendar: TermCalendar | undefined,
  terms: Term[],
  today: string
): TermStatus | null {
  if (term.phase === "in-session" || term.phase === "past") {
    const next = upcomingRegistration(term, terms, today)
    if (next) return { text: registrationOpensText(next), switchTo: next.code }
  }
  if (term.phase === "past" || !calendar) return null
  if (term.phase === "pre-registration") return { text: countdownText(calendar, today) }

  const registration = term.phase === "registration" ? registrationText(calendar, today) : null
  if (registration) return { text: registration }
  if (today < calendar.start) return { text: `Classes start ${formatMonthDay(calendar.start)}` }
  if (calendar.finals) return { text: `Finals ${formatDayRange(calendar.finals.start, calendar.finals.end)}` }
  return { text: `Classes end ${formatMonthDay(calendar.end)}` }
}

function registrationOpensText(term: Term): string {
  if (term.phase === "registration") return `${term.name} registration open`
  return `${term.name} registration opens ${formatMonthDay(term.registrationOpens)}`
}

// TODO(issue 75): count down to the student's own
// registration time from earned credits, not just Phase I.
function countdownText(calendar: TermCalendar, today: string): string {
  const phaseOne = calendar.registration.find((r) => r.phase === "I")
  const opens = phaseOne?.start ?? calendar.registrationOpens
  const days = daysBetween(today, opens)
  const when = days === 1 ? "tomorrow" : `in ${days} days`
  return `${phaseOne ? "Phase I" : "Registration"} opens ${formatMonthDay(opens)} - ${when}`
}

function registrationText(calendar: TermCalendar, today: string): string | null {
  const phases = calendar.registration
  const current = phases.findIndex((r) => r.start <= today && today <= r.end)
  const next = phases.find((r) => r.start > today)
  if (current >= 0 && current < phases.length - 1 && next) {
    return `Phase ${next.phase} opens ${formatMonthDay(next.start)}`
  }
  if (current >= 0) {
    const r = phases[current]
    return `Phase ${r.phase} open until ${formatMonthDay(r.end)}`
  }
  if (next) return `Phase ${next.phase} opens ${formatMonthDay(next.start)}`
  return null
}

/** The newest newer term registering now, else the soonest opening soon. */
function upcomingRegistration(term: Term, terms: Term[], today: string): Term | null {
  const newer = terms.filter((t) => t.code > term.code)
  const registering = newer.filter((t) => t.phase === "registration")
  if (registering.length > 0) {
    return registering.reduce((a, b) => (b.code > a.code ? b : a))
  }
  const soon = newer.filter(
    (t) =>
      t.phase === "pre-registration" &&
      daysBetween(today, t.registrationOpens) <= UPCOMING_REGISTRATION_DAYS
  )
  if (soon.length === 0) return null
  return soon.reduce((a, b) => (b.registrationOpens < a.registrationOpens ? b : a))
}

export type KeyDateState = "past" | "current" | "upcoming"

export interface KeyDateRow {
  label: string
  dates: string
  note?: string
  state: KeyDateState
}

function stateOf(start: string, end: string, today: string): KeyDateState {
  if (today > end) return "past"
  if (today >= start) return "current"
  return "upcoming"
}

/** The weekday before finals week, or the term's last day without finals. */
export function lastDayOfClasses(calendar: TermCalendar): string {
  if (!calendar.finals) return calendar.end
  let day = addDays(calendar.finals.start, -1)
  while (weekday(day) === 0 || weekday(day) === 6) day = addDays(day, -1)
  return day
}

/** Shortens a holiday name to fit beside its date in the popover. */
function holidayLabel(name: string): string {
  return stripTrailingNote(name).replace("Martin Luther King Jr.", "MLK Jr.")
}

/** Rows for the key dates popover, in date order. */
export function keyDateRows(calendar: TermCalendar, today: string): KeyDateRow[] {
  const rows: KeyDateRow[] = []
  const range = (label: string, start: string, end: string, note?: string) =>
    rows.push({ label, dates: formatWeekdayRange(start, end), note, state: stateOf(start, end, today) })

  if (calendar.registration.length > 0) {
    for (const r of calendar.registration) range(`Phase ${r.phase}`, r.start, r.end)
  } else {
    range("Registration", calendar.registrationOpens, calendar.registrationOpens)
  }
  range("Classes", calendar.start, lastDayOfClasses(calendar))

  // A holiday inside another (Native American Heritage Day within
  // Thanksgiving Break) adds no day off.
  const holidays = calendar.holidays.filter(
    (h, i) =>
      !calendar.holidays.some(
        (o, j) =>
          j !== i &&
          o.start <= h.start &&
          h.end <= o.end &&
          // Of two identical ranges, keep the first.
          (o.start !== h.start || o.end !== h.end || j < i)
      )
  )
  holidays.forEach((h, i) =>
    range(i === 0 ? "No classes" : "", h.start, h.end, holidayLabel(h.name))
  )

  if (calendar.finals) range("Finals", calendar.finals.start, calendar.finals.end)
  return rows
}
