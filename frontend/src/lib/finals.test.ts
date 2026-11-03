import { describe, expect, it } from "vitest"
import type { HydratedSection } from "./api"
import { finalsWeekDays } from "./finals"

function exam(date: string): HydratedSection {
  return {
    crn: date,
    term: "202640",
    subject: "CSCI",
    courseNumber: "247",
    title: "Computer Systems I",
    credits: 5,
    instructor: "",
    meetingTimes: [],
    final: { date, start: "0800", end: "1000", certain: true },
    enrollment: 0,
    maxEnrollment: 0,
    seatsAvailable: 0,
    waitCount: 0,
    isOpen: true,
  }
}

const fallWeek = ["2026-12-07", "2026-12-08", "2026-12-09", "2026-12-10", "2026-12-11"]

describe("finalsWeekDays", () => {
  it("uses the term calendar's week when known", () => {
    expect(finalsWeekDays([exam("2026-12-09")], { start: "2026-12-07", end: "2026-12-11" })).toEqual(fallWeek)
  })

  it("falls back to the week of the earliest exam", () => {
    // Out of order, and none on the Monday.
    expect(finalsWeekDays([exam("2026-12-10"), exam("2026-12-08")], null)).toEqual(fallWeek)
  })

  it("is empty with no week and no exams", () => {
    expect(finalsWeekDays([], null)).toEqual([])
  })
})
