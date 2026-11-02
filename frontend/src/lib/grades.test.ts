import { describe, it, expect } from "vitest"
import { describeGrades } from "./grades"

describe("describeGrades", () => {
  it("prefers a professor GPA", () => {
    expect(describeGrades(3.456, 0.9, "course_professor")).toMatchObject({ kind: "professor", label: "3.46 GPA" })
  })

  it("falls back to the course GPA for other sources", () => {
    expect(describeGrades(2.5, undefined, "course")).toMatchObject({ kind: "course", label: "2.50 GPA" })
  })

  it("uses the pass rate for S/U courses, including 0%", () => {
    expect(describeGrades(undefined, 0.874)).toMatchObject({ kind: "pass", label: "87% pass" })
    expect(describeGrades(0, 0)).toMatchObject({ kind: "pass", label: "0% pass" })
  })

  it("reports no data", () => {
    expect(describeGrades()).toMatchObject({ kind: "none", label: "N/A" })
  })
})
