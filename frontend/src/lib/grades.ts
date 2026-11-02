export type GradeKind = "professor" | "course" | "pass" | "none"

export interface GradeSummary {
  kind: GradeKind
  label: string
  tooltip: string
}

/** Picks the best available grade stat: GPA, then S/U pass rate, then none. */
export function describeGrades(gpa?: number, passRate?: number, gpaSource?: string): GradeSummary {
  if (gpa) {
    if (gpaSource === "course_professor") {
      return { kind: "professor", label: `${gpa.toFixed(2)} GPA`, tooltip: "GPA for this professor teaching this course" }
    }
    return { kind: "course", label: `${gpa.toFixed(2)} GPA`, tooltip: "Average GPA for this course" }
  }
  if (passRate != null) {
    return {
      kind: "pass",
      label: `${Math.round(passRate * 100)}% pass`,
      tooltip: "Historical pass rate for this S/U graded course",
    }
  }
  return { kind: "none", label: "N/A", tooltip: "No historical grade data available" }
}
