import { describe, expect, it } from "vitest"
import {
  axisDomain,
  axisTicks,
  currentTermCode,
  distributionColumns,
  jitter,
  quartersGrid,
  termName,
  termShort,
} from "./stats"

const counts = (set: Record<number, number>) =>
  Array.from({ length: 17 }, (_, i) => set[i] ?? 0)

describe("distributionColumns", () => {
  it("shows letters and W for a letter-graded course", () => {
    expect(distributionColumns(counts({ 0: 5, 12: 1 }))).toEqual([
      0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
    ])
  })
  it("drops letters for an S/U-only course and keeps W last", () => {
    expect(distributionColumns(counts({ 15: 20, 16: 1 }))).toEqual([15, 16, 12])
  })
  it("appends P/NP when a letter course has any", () => {
    expect(distributionColumns(counts({ 3: 9, 14: 1 })).slice(-3)).toEqual([
      13, 14, 12,
    ])
  })
})

describe("term labels", () => {
  it.each([
    ["202340", "Fall 2023", "Fall 23"],
    ["202710", "Winter 2027", "Win 27"],
    ["202630", "Summer 2026", "Sum 26"],
  ])("%s", (code, name, short) => {
    expect(termName(code)).toBe(name)
    expect(termShort(code)).toBe(short)
  })
})

describe("axisDomain", () => {
  it("keeps 2.0 to 4.0 when every GPA is above 2", () => {
    expect(axisDomain("gpa", [2.4, 3.9])).toEqual({ lo: 2, hi: 4, step: 0.5 })
  })
  it("extends down to the half step below the lowest GPA", () => {
    expect(axisDomain("gpa", [1.74, 3.1])).toEqual({
      lo: 1.5,
      hi: 4,
      step: 0.5,
    })
  })
  it("runs pass rates from 80% unless one is lower", () => {
    expect(axisDomain("pass", [0.93])).toEqual({ lo: 0.8, hi: 1, step: 0.1 })
    expect(axisDomain("pass", [0.42]).lo).toBeCloseTo(0.4)
  })
  it("ticks without float drift", () => {
    expect(axisTicks({ lo: 0.4, hi: 1, step: 0.2 })).toEqual([0.4, 0.6, 0.8, 1])
  })
})

describe("jitter", () => {
  it("is stable and in range", () => {
    expect(jitter("20234040981")).toBe(jitter("20234040981"))
    for (let crn = 40000; crn < 40200; crn++) {
      const j = jitter(`202340${crn}`)
      expect(j).toBeGreaterThanOrEqual(-1)
      expect(j).toBeLessThanOrEqual(1)
    }
  })
  it("spreads consecutive CRNs apart", () => {
    const gaps = Array.from({ length: 50 }, (_, i) =>
      Math.abs(jitter(`202340${40000 + i}`) - jitter(`202340${40001 + i}`))
    )
    expect(gaps.reduce((a, b) => a + b) / gaps.length).toBeGreaterThan(0.3)
  })
})

describe("quartersGrid", () => {
  it("ends at the academic year of the next term and marks it upcoming", () => {
    // Fall 2026 is academic year 2027; Winter 2027 too.
    const { years, rows } = quartersGrid(
      ["202540", "202610", "202640", "202710"],
      "202710",
      "202640"
    )
    expect(years).toEqual([2022, 2023, 2024, 2025, 2026, 2027])
    const fall = rows.find((r) => r.label === "F")!
    const winter = rows.find((r) => r.label === "W")!
    expect(fall.cells.slice(-2)).toEqual(["taught", "taught"])
    expect(winter.cells.slice(-2)).toEqual(["taught", "upcoming"])
  })
  it("ends at the current academic year without a next term", () => {
    expect(quartersGrid([], undefined, "202620").years.at(-1)).toBe(2026)
  })
})

describe("currentTermCode", () => {
  it("follows the month", () => {
    expect(currentTermCode(new Date(2026, 10, 19))).toBe("202640")
    expect(currentTermCode(new Date(2027, 0, 2))).toBe("202710")
  })
})
