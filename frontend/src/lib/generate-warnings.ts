import type { GenerateResponse, CourseResult } from "@/lib/api"

export interface GenerateWarning {
  type: "warning" | "info"
  message: string
}

const statusMessages: Record<string, (name: string) => string> = {
  not_exists: (name) => `${name} doesn't exist`,
  not_offered: (name) => `${name} is not offered this term`,
  blocked: (name) => `All ${name} sections conflict with your blocked times`,
  crn_filtered: (name) => `No ${name} sections match your CRN filters`,
  async_only: (name) => `${name} only has async sections (shown separately)`,
}

function getCourseIssues(courseResults: CourseResult[]): GenerateWarning | null {
  const issues = courseResults.filter((r) => r.status !== "found")
  if (issues.length === 0) return null

  if (issues.length === 1) {
    const issue = issues[0]
    const getMessage = statusMessages[issue.status]
    if (!getMessage) return null
    return { type: "warning", message: getMessage(issue.name) }
  }

  const lines = issues
    .map((issue) => {
      const getMessage = statusMessages[issue.status]
      return getMessage ? `- ${getMessage(issue.name)}` : null
    })
    .filter(Boolean)

  if (lines.length === 0) return null

  return {
    type: "warning",
    message: `Course issues:\n${lines.join("\n")}`,
  }
}

export function getGenerateWarnings(data: GenerateResponse): GenerateWarning[] {
  const warnings: GenerateWarning[] = []

  const courseIssue = getCourseIssues(data.courseResults)
  if (courseIssue) warnings.push(courseIssue)

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
