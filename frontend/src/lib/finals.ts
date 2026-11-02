import type { DateRange, FinalGuessReason, HydratedSection, MeetingTime } from "@/lib/api"
import { addDays, weekday } from "@/lib/dates"

export const FINAL_GUESS_REASONS: Record<FinalGuessReason, string> = {
  lab: "Has a lab",
  multiple: "Multiple meeting times",
  weekly: "Meets once a week; may have no final",
}

/**
 * Turns each section's final into a one-meeting section on its exam weekday,
 * so the week grid draws exams with the same blocks and colors as classes.
 */
export function examSections(sections: HydratedSection[]): HydratedSection[] {
  return sections.flatMap((section) => {
    const final = section.final
    if (!final) return []
    const days = Array<boolean>(7).fill(false)
    days[weekday(final.date)] = true
    const exam: MeetingTime = { days, startTime: final.start, endTime: final.end, building: "", room: "" }
    return [{ ...section, meetingTimes: [exam] }]
  })
}

/**
 * Finals week's five weekdays: the term calendar's week when known, else the
 * week of the earliest exam.
 */
export function finalsWeekDays(exams: HydratedSection[], week: DateRange | null): string[] {
  let monday = week?.start
  if (!monday) {
    const first = exams.map((e) => e.final?.date ?? "").filter(Boolean).sort()[0]
    if (!first) return []
    monday = addDays(first, 1 - weekday(first))
  }
  return Array.from({ length: 5 }, (_, i) => addDays(monday, i))
}
