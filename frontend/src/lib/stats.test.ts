import { describe, expect, it } from "vitest"
import {
  axisDomain,
  axisTicks,
  countTicks,
  currentTermCode,
  distributionColumns,
  dotLayout,
  formatDelta,
  quartersGrid,
  termName,
} from "./stats"
import type { StatsDot } from "./api"

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

describe("termName", () => {
  it.each([
    ["202340", "Fall 2023"],
    ["202710", "Winter 2027"],
    ["202630", "Summer 2026"],
  ])("%s", (code, name) => {
    expect(termName(code)).toBe(name)
  })
})

describe("formatDelta", () => {
  it.each([
    ["gpa", 3.1, 2.97, "+0.13"],
    ["gpa", 2.58, 2.97, "-0.39"],
    // 3.004 and 2.996 both show as 3.00, so the gap shows as none.
    ["gpa", 3.004, 2.996, "0.00"],
    // Shown as 2.58 and 2.97 even though the raw gap rounds to 0.38.
    ["gpa", 2.584, 2.966, "-0.39"],
    ["pass", 0.97, 0.94, "+3%"],
    ["pass", 0.9, 0.94, "-4%"],
  ] as const)("%s %d vs %d", (metric, value, course, want) => {
    expect(formatDelta(metric, value, course)).toBe(want)
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

describe("dotLayout", () => {
  const dot = (crn: string, value: number, mine = false): StatsDot => ({
    term: "202340",
    crn,
    professor: "",
    students: 20,
    value,
    mine,
  })
  const current = { term: "202340", crn: "c" }

  it("bins by tenths of GPA and stacks current, mine, then others", () => {
    const { points, xDomain } = dotLayout(
      [dot("o", 3.04), dot("m", 2.96, true), dot("c", 3.01), dot("x", 3.4)],
      "gpa",
      current,
      { width: 400, height: 150 }
    )
    expect(xDomain).toEqual([1.95, 4.05])
    const at = (crn: string) => points.find((p) => p.crn === crn)!
    // 3.04, 2.96, and 3.01 all round to the 3.0 bin.
    expect([at("c"), at("m"), at("o")].map((p) => [p.x, p.y])).toEqual([
      [3, 0.5],
      [3, 1.5],
      [3, 2.5],
    ])
    expect(at("x").x).toBeCloseTo(3.4)
    // The current section is drawn last, on top.
    expect(points.at(-1)!.crn).toBe("c")
  })

  it("wraps a tall bin into columns and keeps height in sections", () => {
    const dots = Array.from({ length: 100 }, (_, i) => dot(String(i), 3))
    const { points, yMax, r } = dotLayout(dots, "gpa", current, {
      width: 420,
      height: 150,
    })
    const cols = new Set(points.map((p) => p.x)).size
    expect(cols).toBeGreaterThan(1)
    // Every full row holds cols sections, so the top dot's row ends at 100.
    const top = Math.max(...points.map((p) => p.y))
    expect(top + cols / 2).toBe(Math.ceil(100 / cols) * cols)
    expect(yMax).toBeGreaterThanOrEqual(top + cols / 2)
    expect(r).toBeGreaterThan(1)
    // Columns stay inside the 0.1 bin.
    for (const p of points) expect(Math.abs(p.x - 3)).toBeLessThan(0.05)
  })

  it("bins pass rates by fifths of the axis step", () => {
    const { points } = dotLayout([dot("a", 0.931)], "pass", current, {
      width: 400,
      height: 150,
    })
    // An axis from 80% in tenths gives 2% bins; 93.1% falls in 94%.
    expect(points[0].x).toBeCloseTo(0.94)
  })
})

describe("countTicks", () => {
  it.each([
    [3.2, [0, 1, 2, 3]],
    [12.6, [0, 5, 10]],
    [16, [0, 5, 10, 15]],
    [150, [0, 50, 100, 150]],
    [0.6, [0]],
  ])("%d", (max, want) => {
    expect(countTicks(max)).toEqual(want)
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
    expect(fall.cells.slice(-2)).toEqual([
      { term: "202540", state: "taught" },
      { term: "202640", state: "taught" },
    ])
    expect(winter.cells.slice(-2).map((c) => c.state)).toEqual([
      "taught",
      "upcoming",
    ])
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
