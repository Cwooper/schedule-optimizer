import { describe, expect, it } from "vitest"
import { formatCredits, formatMeetingTime } from "./meetings"

const days = (...on: number[]) =>
  Array.from({ length: 7 }, (_, i) => on.includes(i))

describe("formatMeetingTime", () => {
  it.each([
    [
      { days: days(1, 3), startTime: "1000", endTime: "1050" },
      "MW 10:00am-10:50am",
    ],
    // A weekend meeting has set times too.
    [{ days: days(6), startTime: "1200", endTime: "1650" }, "S 12:00pm-4:50pm"],
    [{ days: days() }, "Times arranged with instructor"],
    [{ days: days(), room: "ASNC" }, "Asynchronous, no set times"],
  ])("%j", (meeting, want) => {
    expect(formatMeetingTime(meeting)).toBe(want)
  })
})

describe("formatCredits", () => {
  it.each([
    [4, undefined, "4"],
    [4, 4, "4"],
    [1, 3, "1-3"],
  ])("%d to %s", (lo, hi, want) => {
    expect(formatCredits(lo, hi)).toBe(want)
  })
})
