import { describe, it, expect } from "vitest"
import { getGenerateWarnings } from "./generate-warnings"
import type { GenerateResponse } from "./api"

function makeResponse(
  overrides: Partial<GenerateResponse> = {}
): GenerateResponse {
  return {
    courses: {},
    sections: {},
    schedules: [],
    asyncs: [],
    courseResults: [],
    stats: { totalGenerated: 0, timeMs: 1 },
    ...overrides,
  }
}

describe("getGenerateWarnings", () => {
  it("returns empty array when all courses found and schedules exist", () => {
    const data = makeResponse({
      schedules: [{ crns: ["123"], score: 1, weights: [] }],
      courseResults: [{ name: "CSCI 241", status: "found", count: 3 }],
      stats: { totalGenerated: 1, timeMs: 1 },
    })
    expect(getGenerateWarnings(data)).toEqual([])
  })

  it.each([
    ["not_exists", "CSCI 247 doesn't exist"],
    ["not_offered", "CSCI 247 is not offered this term"],
    ["blocked", "All CSCI 247 sections conflict with your blocked times"],
    ["crn_filtered", "No CSCI 247 sections match your CRN filters"],
    ["async_only", "CSCI 247 only has async sections (shown separately)"],
  ] as const)("returns a warning for a %s course", (status, message) => {
    const data = makeResponse({
      schedules: [{ crns: ["123"], score: 1, weights: [] }],
      courseResults: [{ name: "CSCI 247", status }],
    })
    expect(getGenerateWarnings(data)).toEqual([
      expect.objectContaining({ type: "warning", message }),
    ])
  })

  it("summarizes multiple courses with same status", () => {
    const data = makeResponse({
      schedules: [{ crns: ["123"], score: 1, weights: [] }],
      courseResults: [
        { name: "CSCI 999", status: "not_exists" },
        { name: "MATH 888", status: "not_exists" },
      ],
    })
    const warnings = getGenerateWarnings(data)
    expect(warnings).toHaveLength(1)
    expect(warnings[0].message).toBe("2 courses don't exist")
  })

  it("produces separate warnings for different status types", () => {
    const data = makeResponse({
      schedules: [{ crns: ["123"], score: 1, weights: [] }],
      courseResults: [
        { name: "CSCI 999", status: "not_exists" },
        { name: "MATH 101", status: "not_offered" },
      ],
    })
    const warnings = getGenerateWarnings(data)
    expect(warnings).toHaveLength(2)
    expect(warnings[0].message).toBe("CSCI 999 doesn't exist")
    expect(warnings[1].message).toBe("MATH 101 is not offered this term")
  })

  it("mixes summary and individual for different statuses", () => {
    const data = makeResponse({
      schedules: [{ crns: ["123"], score: 1, weights: [] }],
      courseResults: [
        { name: "CSCI 999", status: "not_exists" },
        { name: "MATH 888", status: "not_exists" },
        { name: "PHYS 101", status: "async_only" },
      ],
    })
    const warnings = getGenerateWarnings(data)
    expect(warnings).toHaveLength(2)
    expect(warnings[0].message).toBe("2 courses don't exist")
    expect(warnings[1].message).toBe(
      "PHYS 101 only has async sections (shown separately)"
    )
  })

  it("ignores found courses in issue list", () => {
    const data = makeResponse({
      schedules: [{ crns: ["123"], score: 1, weights: [] }],
      courseResults: [
        { name: "CSCI 241", status: "found", count: 3 },
        { name: "CSCI 999", status: "not_exists" },
      ],
    })
    const warnings = getGenerateWarnings(data)
    expect(warnings).toHaveLength(1)
    expect(warnings[0].message).toBe("CSCI 999 doesn't exist")
  })

  it("returns no_schedules warning when schedules is empty", () => {
    const data = makeResponse({
      schedules: [],
      courseResults: [{ name: "CSCI 241", status: "found", count: 3 }],
    })
    const warnings = getGenerateWarnings(data)
    expect(warnings.some((w) => w.message.includes("No valid schedules"))).toBe(
      true
    )
  })

  it("returns truncated info when totalGenerated > schedules returned", () => {
    const data = makeResponse({
      schedules: [{ crns: ["123"], score: 1, weights: [] }],
      courseResults: [{ name: "CSCI 241", status: "found", count: 3 }],
      stats: { totalGenerated: 5000, timeMs: 1 },
    })
    const warnings = getGenerateWarnings(data)
    expect(warnings).toHaveLength(1)
    expect(warnings[0].type).toBe("info")
    expect(warnings[0].message).toContain("Showing top 1 of 5,000")
  })

  it("does not show truncated when totalGenerated equals schedules length", () => {
    const data = makeResponse({
      schedules: [{ crns: ["123"], score: 1, weights: [] }],
      courseResults: [{ name: "CSCI 241", status: "found", count: 3 }],
      stats: { totalGenerated: 1, timeMs: 1 },
    })
    const warnings = getGenerateWarnings(data)
    expect(warnings).toHaveLength(0)
  })

  it("can return both course issues and no_schedules warning", () => {
    const data = makeResponse({
      schedules: [],
      courseResults: [{ name: "CSCI 999", status: "not_exists" }],
    })
    const warnings = getGenerateWarnings(data)
    expect(warnings).toHaveLength(2)
    expect(warnings[0].message).toContain("CSCI 999")
    expect(warnings[1].message).toContain("No valid schedules")
  })
})
