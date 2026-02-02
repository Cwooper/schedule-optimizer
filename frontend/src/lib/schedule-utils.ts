import type {
  GenerateCourseInfo,
  GenerateSectionInfo,
  GenerateResponse,
  HydratedSection,
  ScheduleRef,
} from "./api"

/**
 * Hydrate a single CRN into a full HydratedSection by joining course and section data.
 * Returns null if section or course not found.
 */
export function hydrateSection(
  crn: string,
  courses: Record<string, GenerateCourseInfo>,
  sections: Record<string, GenerateSectionInfo>
): HydratedSection | null {
  const section = sections[crn]
  if (!section) return null

  const course = courses[section.courseKey]
  if (!course) return null

  return {
    crn: section.crn,
    term: section.term,
    subject: course.subject,
    courseNumber: course.courseNumber,
    title: course.title,
    credits: course.credits,
    instructor: section.instructor,
    meetingTimes: section.meetingTimes,
    enrollment: section.enrollment,
    maxEnrollment: section.maxEnrollment,
    seatsAvailable: section.seatsAvailable,
    waitCount: section.waitCount,
    isOpen: section.isOpen,
  }
}

/**
 * Hydrate a schedule reference into full section data.
 * Filters out any CRNs that can't be resolved.
 */
export function hydrateSchedule(
  schedule: ScheduleRef,
  courses: Record<string, GenerateCourseInfo>,
  sections: Record<string, GenerateSectionInfo>
): {
  courses: HydratedSection[]
  score: number
  weights: { name: string; value: number }[]
} {
  const hydratedCourses: HydratedSection[] = []

  for (const crn of schedule.crns) {
    const hydrated = hydrateSection(crn, courses, sections)
    if (hydrated) {
      hydratedCourses.push(hydrated)
    }
  }

  return {
    courses: hydratedCourses,
    score: schedule.score,
    weights: schedule.weights,
  }
}

/**
 * Hydrate all async sections from a generate response.
 */
export function hydrateAsyncs(response: GenerateResponse): HydratedSection[] {
  const hydrated: HydratedSection[] = []

  for (const crn of response.asyncs) {
    const section = hydrateSection(crn, response.courses, response.sections)
    if (section) {
      hydrated.push(section)
    }
  }

  return hydrated
}

/**
 * Sort sections: open sections first, then by CRN.
 * Mutates the array in place and returns it.
 */
export function sortSectionsByAvailability(
  sections: HydratedSection[]
): HydratedSection[] {
  return sections.sort((a, b) => {
    if (a.isOpen !== b.isOpen) return a.isOpen ? -1 : 1
    return a.crn.localeCompare(b.crn)
  })
}
