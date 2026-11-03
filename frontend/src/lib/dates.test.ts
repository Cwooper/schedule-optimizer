import { afterAll, beforeAll, describe, expect, it, vi } from "vitest"
import { pacificToday } from "./dates"

describe("pacificToday", () => {
  // A student east of Pacific: their local date turns over hours earlier.
  beforeAll(() => {
    vi.stubEnv("TZ", "America/New_York")
  })
  afterAll(() => {
    vi.unstubAllEnvs()
  })

  it.each([
    ["2026-11-03T05:30:00Z", "2026-11-02"], // 00:30 in New York, 21:30 in Bellingham
    ["2026-11-03T08:00:00Z", "2026-11-03"], // midnight Pacific
    ["2026-07-01T06:59:59Z", "2026-06-30"], // daylight time
  ])("%s is %s at WWU", (instant, want) => {
    expect(pacificToday(new Date(instant))).toBe(want)
  })
})
