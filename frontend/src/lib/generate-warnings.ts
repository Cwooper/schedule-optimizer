import type { GenerateResponse, CourseResult } from "@/lib/api"

export interface GenerateWarning {
  type: "warning" | "info"
  message: string
}

// Messages for individual courses (named)
const statusMessages: Record<string, (name: string) => string> = {
  not_exists: (name) => `${name} doesn't exist`,
  not_offered: (name) => `${name} is not offered this term`,
  blocked: (name) => `All ${name} sections conflict with your blocked times`,
  crn_filtered: (name) => `No ${name} sections match your CRN filters`,
  async_only: (name) => `${name} only has async sections (shown separately)`,
}

// Summary messages for multiple courses with the same status
const statusSummaries: Record<string, (count: number) => string> = {
  not_exists: (n) => `${n} courses don't exist`,
  not_offered: (n) => `${n} courses are not offered this term`,
  blocked: (n) => `${n} courses have all sections blocked`,
  crn_filtered: (n) => `${n} courses have no sections matching CRN filters`,
  async_only: (n) => `${n} courses only have async sections (shown separately)`,
}

function getCourseIssues(courseResults: CourseResult[]): GenerateWarning[] {
  const issues = courseResults.filter((r) => r.status !== "found")
  if (issues.length === 0) return []

  // Group issues by status
  const grouped = new Map<string, CourseResult[]>()
  for (const issue of issues) {
    const group = grouped.get(issue.status)
    if (group) {
      group.push(issue)
    } else {
      grouped.set(issue.status, [issue])
    }
  }

  // One warning per status group
  const warnings: GenerateWarning[] = []
  for (const [status, group] of grouped) {
    if (group.length === 1) {
      const getMessage = statusMessages[status]
      if (getMessage) {
        warnings.push({ type: "warning", message: getMessage(group[0].name) })
      }
    } else {
      const getSummary = statusSummaries[status]
      if (getSummary) {
        warnings.push({ type: "warning", message: getSummary(group.length) })
      }
    }
  }

  return warnings
}

export function getGenerateWarnings(data: GenerateResponse): GenerateWarning[] {
  const warnings: GenerateWarning[] = []

  warnings.push(...getCourseIssues(data.courseResults))

  if (data.schedules.length === 0) {
    warnings.push({
      type: "warning",
      message: "No valid schedules found for the selected courses",
    })
  } else if (data.stats.totalGenerated > data.schedules.length) {
    warnings.push({
      type: "info",
      message: `Showing top ${data.schedules.length.toLocaleString()} of ${data.stats.totalGenerated.toLocaleString()} schedules`,
    })
  }

  return warnings
}
