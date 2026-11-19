import { describe, expect, it } from "vitest"
import type { HydratedSection, MeetingTime } from "./api"
import { computeScheduleStats } from "./schedule-stats"
import { DAY_LETTERS } from "./schedule-utils"

function meeting(days: string, startTime?: string, endTime?: string): MeetingTime {
  return { days: DAY_LETTERS.map((d) => days.includes(d)), startTime, endTime }
}

function section(meetingTimes: MeetingTime[], extra: Partial<HydratedSection> = {}): HydratedSection {
  return {
    crn: "10000",
    term: "202710",
    subject: "CSCI",
    courseNumber: "301",
    title: "Data Structures",
    credits: 5,
    instructor: "",
    meetingTimes,
    enrollment: 0,
    maxEnrollment: 0,
    seatsAvailable: 0,
    waitCount: 0,
    isOpen: true,
    ...extra,
  }
}

describe("computeScheduleStats", () => {
  it("returns empty stats for an empty schedule", () => {
    expect(computeScheduleStats([])).toEqual({
      start: null,
      end: null,
      gapMins: 0,
      gpa: null,
      gpaKnown: 0,
      sections: 0,
      open: 0,
      credits: 0,
      days: [false, false, false, false, false],
    })
  })

  it("merges overlapping meetings before measuring gaps", () => {
    const stats = computeScheduleStats([
      section([meeting("M", "0900", "1050")]),
      section([meeting("M", "1000", "1120")]),
      section([meeting("M", "1300", "1350")]),
    ])
    // Only 11:20 to 1:00; the overlap is not a negative gap.
    expect(stats.gapMins).toBe(100)
    expect(stats.end).toEqual({ day: 0, mins: 13 * 60 + 50 })
  })

  it("counts a meeting contained in another as part of its block", () => {
    const stats = computeScheduleStats([
      section([meeting("T", "0800", "1200")]),
      section([meeting("T", "0900", "0950")]),
      section([meeting("T", "1230", "1320")]),
    ])
    expect(stats.gapMins).toBe(30)
  })

  it("ignores passing time of 15 minutes but counts 16", () => {
    const passing = computeScheduleStats([
      section([meeting("MWF", "0900", "0950")]),
      section([meeting("MWF", "1005", "1055")]),
    ])
    expect(passing.gapMins).toBe(0)

    const gap = computeScheduleStats([
      section([meeting("MWF", "0900", "0950")]),
      section([meeting("MWF", "1006", "1055")]),
    ])
    expect(gap.gapMins).toBe(3 * 16)
  })

  it("ignores weekend meetings", () => {
    const stats = computeScheduleStats([
      section([meeting("W", "1000", "1050"), meeting("SU", "0700", "2100")]),
    ])
    expect(stats.start).toEqual({ day: 2, mins: 600 })
    expect(stats.end).toEqual({ day: 2, mins: 650 })
    expect(stats.days).toEqual([false, false, true, false, false])
  })

  it("leaves untimed sections out of times but counts their credits, GPA, and seats", () => {
    const stats = computeScheduleStats([
      section([meeting("TR", "1000", "1150")], { credits: 4, gpa: 3.0 }),
      section([meeting("", undefined, undefined)], { credits: 3, gpa: 3.5, isOpen: false }),
      section([], { credits: 2 }),
    ])
    expect(stats.start).toEqual({ day: 1, mins: 600 })
    expect(stats.end).toEqual({ day: 1, mins: 710 })
    expect(stats.days).toEqual([false, true, false, true, false])
    expect(stats.credits).toBe(9)
    expect(stats.gpa).toBeCloseTo(3.25)
    expect(stats.gpaKnown).toBe(2)
    expect(stats.sections).toBe(3)
    expect(stats.open).toBe(2)
  })

  it("averages GPA over sections with data only", () => {
    const partial = computeScheduleStats([
      section([], { gpa: 3.2 }),
      section([], { gpa: 3.6 }),
      section([], { gpa: 0 }),
      section([]),
    ])
    expect(partial.gpa).toBeCloseTo(3.4)
    expect(partial.gpaKnown).toBe(2)

    const none = computeScheduleStats([section([]), section([])])
    expect(none.gpa).toBeNull()
    expect(none.gpaKnown).toBe(0)
  })

  it("picks the first weekday when the earliest start or latest end ties", () => {
    const stats = computeScheduleStats([
      section([meeting("TR", "0800", "0850")]),
      section([meeting("TR", "1400", "1550")]),
      section([meeting("MW", "0900", "1550")]),
    ])
    expect(stats.start).toEqual({ day: 1, mins: 480 })
    expect(stats.end).toEqual({ day: 0, mins: 950 })
  })
})
