import { useQuery, useMutation } from "@tanstack/react-query"
import {
  getTerms,
  getSubjects,
  validateCourse,
  getCRN,
  generateSchedules,
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

export function useValidateCourse(term: string, subject: string, courseNumber: string) {
  return useQuery({
    queryKey: ["validate-course", term, subject, courseNumber],
    queryFn: () => validateCourse(term, subject, courseNumber),
    enabled: Boolean(term && subject && courseNumber),
  })
}

export function useCRN(crn: string, term?: string) {
  return useQuery({
    queryKey: ["crn", crn, term],
    queryFn: () => getCRN(crn, term),
    enabled: Boolean(crn),
  })
}

export function useGenerateSchedules() {
  return useMutation({
    mutationFn: (req: GenerateRequest) => generateSchedules(req),
  })
}
