import { useQuery, useMutation, keepPreviousData } from "@tanstack/react-query"
import { toast } from "sonner"
import {
  getTerms,
  getCalendar,
  getSubjects,
  getCourse,
  getCRN,
  generateSchedules,
  validateCourses,
  searchCourses,
  getAnnouncement,
  getStatsCourses,
  getStatsCourse,
  getStatsSection,
  ApiError,
  type GenerateRequest,
  type SearchRequest,
} from "@/lib/api"
import { getGenerateWarnings } from "@/lib/generate-warnings"
import { useAppStore, type GenerationParams } from "@/stores/app-store"

export function useAnnouncement() {
  return useQuery({
    queryKey: ["announcement"],
    queryFn: getAnnouncement,
    staleTime: 5 * 60 * 1000,
  })
}

// Statistics answer 503 while the grade data loads after a restart, so
// they keep retrying (with the default backoff, about two minutes in all).
function retryWhileLoading(failures: number, error: Error): boolean {
  if (error instanceof ApiError && error.status === 503) return failures < 8
  return failures < 1
}

// The course list changes only when the grade index rebuilds, which after
// the first scrapes rarely changes it.
export function useStatsCourses() {
  return useQuery({
    queryKey: ["stats-courses"],
    queryFn: getStatsCourses,
    staleTime: Infinity,
    retry: retryWhileLoading,
  })
}

export function useStatsCourse(subject: string, number: string) {
  return useQuery({
    queryKey: ["stats-course", subject, number],
    queryFn: () => getStatsCourse(subject, number),
    enabled: Boolean(subject && number),
    staleTime: 60 * 60 * 1000,
    retry: retryWhileLoading,
  })
}

export function useStatsSection(term: string, crn: string) {
  return useQuery({
    queryKey: ["stats-section", term, crn],
    queryFn: () => getStatsSection(term, crn),
    enabled: Boolean(term && crn),
    staleTime: 60 * 60 * 1000,
    retry: retryWhileLoading,
  })
}

export function useTerms() {
  return useQuery({
    queryKey: ["terms"],
    queryFn: getTerms,
  })
}

export function useCalendar(term: string) {
  return useQuery({
    queryKey: ["calendar", term],
    queryFn: () => getCalendar(term),
    enabled: Boolean(term),
    // The registrar calendar is scraped daily.
    staleTime: 60 * 60 * 1000,
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

/**
 * Results apply in the mutation-level callbacks, which run even after the
 * builder that started the request unmounts (the mobile drawer closing),
 * and through the store's sequence so only the latest request lands.
 */
export function useGenerateSchedules() {
  const { mutate } = useMutation({
    mutationFn: ({ req }: { req: GenerateRequest; seq: number }) => generateSchedules(req),
    meta: { handlesError: true },
    onSuccess: (data, { seq }) => {
      if (!useAppStore.getState().settleGeneration(seq, data)?.showWarnings) return
      for (const w of getGenerateWarnings(data)) {
        if (w.type === "info") toast.info(w.message)
        else toast.warning(w.message)
      }
    },
    onError: (error, { seq }) => {
      if (!useAppStore.getState().settleGeneration(seq)) return
      toast.error(error.message || "Failed to generate schedules. Please try again.")
    },
  })
  const isPending = useAppStore((s) => s.pendingGeneration !== null)
  const generate = (req: GenerateRequest, params: GenerationParams, showWarnings: boolean) =>
    mutate({ req, seq: useAppStore.getState().beginGeneration(params, showWarnings) })
  return { generate, isPending }
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
export function hasValidSearchFilter(req: SearchRequest): boolean {
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
