import { useQuery, useMutation, keepPreviousData } from "@tanstack/react-query"
import {
  getTerms,
  getSubjects,
  getCourse,
  getCRN,
  generateSchedules,
  validateCourses,
  searchCourses,
  getAnnouncement,
  type GenerateRequest,
  type SearchRequest,
} from "@/lib/api"

export function useAnnouncement() {
  return useQuery({
    queryKey: ["announcement"],
    queryFn: getAnnouncement,
    staleTime: 5 * 60 * 1000,
  })
}

export function useTerms() {
  return useQuery({
    queryKey: ["terms"],
    queryFn: getTerms,
  })
}

export function useSubjects(term?: string) {
  return useQuery({
    queryKey: ["subjects", term],
    queryFn: () => getSubjects(term),
  })
}

export function useCourse(term: string, subject: string, courseNumber: string) {
  return useQuery({
    queryKey: ["course", term, subject, courseNumber],
    queryFn: () => getCourse(term, subject, courseNumber),
    enabled: Boolean(term && subject && courseNumber),
  })
}

export function useCRN(crn: string, term?: string) {
  return useQuery({
    queryKey: ["crn", crn, term],
    // term! is safe: enabled guard ensures term exists when queryFn runs
    queryFn: () => getCRN(crn, term!),
    enabled: Boolean(crn && term),
  })
}

export function useGenerateSchedules() {
  return useMutation({
    mutationFn: (req: GenerateRequest) => generateSchedules(req),
    meta: { handlesError: true },
  })
}

const MAX_VALIDATE_COURSES = 20

export function useValidateCourses(
  term: string,
  courses: { subject: string; courseNumber: string }[]
) {
  // Limit to first 20 courses to match backend limit
  const limitedCourses = courses.slice(0, MAX_VALIDATE_COURSES)
  return useQuery({
    queryKey: ["validate-courses", term, limitedCourses],
    queryFn: () => validateCourses({ term, courses: limitedCourses }),
    enabled: Boolean(term && limitedCourses.length > 0),
  })
}

/** Strip wildcard characters to match backend validation logic */
function stripWildcards(s: string): string {
  return s.replace(/[*%_]/g, "")
}

/**
 * Check if search request has at least one valid filter.
 * Mirrors backend validation: requires subject, courseNumber, title, or instructor
 * with minimum character counts after stripping wildcards.
 */
function hasValidSearchFilter(req: SearchRequest): boolean {
  const subjectValid = stripWildcards(req.subject ?? "").length >= 2
  const courseNumberValid = stripWildcards(req.courseNumber ?? "").length >= 1
  const titleValid = stripWildcards(req.title ?? "").length >= 2
  const instructorValid = stripWildcards(req.instructor ?? "").length >= 2
  return subjectValid || courseNumberValid || titleValid || instructorValid
}

export function useSearch(req: SearchRequest) {
  return useQuery({
    queryKey: ["search", req],
    queryFn: () => searchCourses(req),
    enabled: hasValidSearchFilter(req),
    placeholderData: keepPreviousData,
    staleTime: 30_000, // 30 seconds - search results don't change often
  })
}
