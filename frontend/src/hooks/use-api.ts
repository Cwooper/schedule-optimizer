import { useQuery, useMutation } from "@tanstack/react-query"
import {
  getTerms,
  getSubjects,
  getCourse,
  getCRN,
  generateSchedules,
  validateCourses,
  type GenerateRequest,
} from "@/lib/api"

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
