import { useAppStore } from "@/stores/app-store"

const API_BASE = "/api"

// --- Types ---

export interface Term {
  code: string
  name: string
}

export interface TermsResponse {
  terms: Term[]
  current: string
}

export interface Subject {
  code: string
  name: string
}

export interface SubjectsResponse {
  subjects: Subject[]
}

export interface CourseSectionInfo {
  crn: string
  instructor: string
  enrollment: number
  maxEnrollment: number
  seatsAvailable: number
  waitCount: number
  isOpen: boolean
}

export interface CourseInfo {
  subject: string
  courseNumber: string
  title: string
  credits: number
}

export type CourseResponse =
  | { course: null }
  | { course: CourseInfo; sections: CourseSectionInfo[]; sectionCount: number }

export interface SectionInfo {
  crn: string
  term: string
  subject: string
  courseNumber: string
  title: string
  instructor: string
  credits: number
  enrollment: number
  maxEnrollment: number
  seatsAvailable: number
  waitCount: number
  isOpen: boolean
  meetingTimes: MeetingTime[]
}

export interface CRNResponse {
  section: SectionInfo | null
}

export interface MeetingTime {
  days: boolean[] // [Sun, Mon, Tue, Wed, Thu, Fri, Sat]
  startTime: string
  endTime: string
  building: string
  room: string
}

// Wire format types - sent once per unique course/section

export interface GenerateCourseInfo {
  subject: string
  courseNumber: string
  title: string
  credits: number
}

export interface GenerateSectionInfo {
  crn: string
  term: string
  courseKey: string
  instructor: string
  enrollment: number
  maxEnrollment: number
  seatsAvailable: number
  waitCount: number
  isOpen: boolean
  meetingTimes: MeetingTime[]
}

export interface ScheduleRef {
  crns: string[]
  score: number
  weights: { name: string; value: number }[]
}

export interface CourseResult {
  name: string
  status: "found" | "async_only" | "blocked" | "crn_filtered" | "not_offered" | "not_exists"
  count?: number
}

export interface GenerateResponse {
  courses: Record<string, GenerateCourseInfo>
  sections: Record<string, GenerateSectionInfo>
  schedules: ScheduleRef[]
  asyncs: string[]
  courseResults: CourseResult[]
  stats: {
    totalGenerated: number
    timeMs: number
  }
}

// Hydrated types - for component consumption

export interface HydratedSection {
  crn: string
  term: string
  subject: string
  courseNumber: string
  title: string
  credits: number
  instructor: string
  meetingTimes: MeetingTime[]
  enrollment: number
  maxEnrollment: number
  seatsAvailable: number
  waitCount: number
  isOpen: boolean
}

export interface CourseSpec {
  subject: string
  courseNumber: string
  required: boolean
  allowedCrns?: string[]
}

export interface GenerateRequest {
  term: string
  courseSpecs: CourseSpec[]
  blockedTimes?: { day: number; startTime: string; endTime: string }[]
  minCourses?: number
  maxCourses?: number
}

export interface ValidateCoursesRequest {
  term: string
  courses: { subject: string; courseNumber: string }[]
}

export interface CourseValidationResult {
  subject: string
  courseNumber: string
  exists: boolean
  title?: string
  sectionCount?: number
}

export interface ValidateCoursesResponse {
  results: CourseValidationResult[]
}

// Search API types - normalized wire format

export interface SearchCourseInfo {
  subject: string
  courseNumber: string
  title: string
  credits: number
  creditsHigh?: number
}

export interface SearchSectionInfo {
  crn: string
  term: string
  courseKey: string
  instructor?: string
  instructorEmail?: string
  enrollment: number
  maxEnrollment: number
  seatsAvailable: number
  waitCount: number
  isOpen: boolean
  campus?: string
  scheduleType?: string
  meetingTimes: MeetingTime[]
}

export interface SearchCourseRef {
  courseKey: string
  sectionKeys: string[] // Format: "term:crn" for cross-term uniqueness
  relevanceScore?: number
}

export interface SearchRequest {
  term?: string
  year?: number
  subject?: string
  courseNumber?: string
  title?: string
  instructor?: string
  openSeats?: boolean
  minCredits?: number
  maxCredits?: number
}

export interface SearchStats {
  totalSections: number
  totalCourses: number
  timeMs: number
}

export interface SearchResponse {
  courses: Record<string, SearchCourseInfo>
  sections: Record<string, SearchSectionInfo>
  results: SearchCourseRef[]
  total: number
  warning?: string
  stats: SearchStats
}

// --- Errors ---

export class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = "ApiError"
    this.status = status
  }
}

// --- API Functions ---

async function fetchAPI<T>(
  endpoint: string,
  options?: RequestInit
): Promise<T> {
  const sessionId = useAppStore.getState().sessionId
  let res: Response
  try {
    res = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...(sessionId ? { "X-Session-ID": sessionId } : {}),
        ...options?.headers,
      },
    })
  } catch {
    throw new ApiError(
      "Unable to reach the server. Check your internet connection.",
      0
    )
  }

  if (!res.ok) {
    if (res.status >= 500) {
      throw new ApiError(
        "Something went wrong on the server. Try again later.",
        res.status
      )
    }
    const body = await res.json().catch(() => ({ error: "Request failed" }))
    throw new ApiError(body.error || `HTTP ${res.status}`, res.status)
  }

  return res.json()
}

export async function getTerms(): Promise<TermsResponse> {
  return fetchAPI<TermsResponse>("/terms")
}

export async function getSubjects(term?: string): Promise<SubjectsResponse> {
  const query = term ? `?term=${encodeURIComponent(term)}` : ""
  return fetchAPI<SubjectsResponse>(`/subjects${query}`)
}

export async function getCourse(
  term: string,
  subject: string,
  courseNumber: string
): Promise<CourseResponse> {
  const query = `?term=${encodeURIComponent(term)}`
  return fetchAPI<CourseResponse>(
    `/course/${encodeURIComponent(subject)}/${encodeURIComponent(courseNumber)}${query}`
  )
}

export async function getCRN(crn: string, term?: string): Promise<CRNResponse> {
  const query = term ? `?term=${encodeURIComponent(term)}` : ""
  return fetchAPI<CRNResponse>(`/crn/${encodeURIComponent(crn)}${query}`)
}

export async function generateSchedules(
  req: GenerateRequest
): Promise<GenerateResponse> {
  return fetchAPI<GenerateResponse>("/generate", {
    method: "POST",
    body: JSON.stringify(req),
  })
}

export async function validateCourses(
  req: ValidateCoursesRequest
): Promise<ValidateCoursesResponse> {
  return fetchAPI<ValidateCoursesResponse>("/courses/validate", {
    method: "POST",
    body: JSON.stringify(req),
  })
}

export async function searchCourses(
  req: SearchRequest
): Promise<SearchResponse> {
  const params = new URLSearchParams()
  if (req.term) params.set("term", req.term)
  if (req.year) params.set("year", String(req.year))
  if (req.subject) params.set("subject", req.subject)
  if (req.courseNumber) params.set("courseNumber", req.courseNumber)
  if (req.title) params.set("title", req.title)
  if (req.instructor) params.set("instructor", req.instructor)
  if (req.openSeats) params.set("openSeats", "true")
  if (req.minCredits !== undefined) params.set("minCredits", String(req.minCredits))
  if (req.maxCredits !== undefined) params.set("maxCredits", String(req.maxCredits))
  return fetchAPI<SearchResponse>(`/search?${params.toString()}`)
}
