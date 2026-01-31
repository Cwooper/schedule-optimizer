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

export interface ValidateCourseResponse {
  exists: boolean
  title?: string
  sectionCount?: number
}

export interface SectionInfo {
  crn: string
  term: string
  subject: string
  courseNumber: string
  title: string
  instructor: string
  credits: number
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

export interface ScheduleSection {
  crn: string
  term: string
  subject: string
  courseNumber: string
  title: string
  credits: number
  instructor: string
  meetingTimes: MeetingTime[]
}

export interface GeneratedSchedule {
  courses: ScheduleSection[]
  score: number
  weights: { name: string; value: number }[]
}

export interface CourseResult {
  name: string
  status: "found" | "async_only" | "not_found"
  count?: number
}

export interface GenerateResponse {
  schedules: GeneratedSchedule[]
  asyncs: SectionInfo[]
  courseResults: CourseResult[]
  stats: {
    totalGenerated: number
    timeMs: number
  }
}

export interface CourseInput {
  subject: string
  courseNumber: string
  required: boolean
  sections?: { crn: string; required: boolean }[] | null
}

export interface GenerateRequest {
  term: string
  courses: CourseInput[]
  minCourses?: number
  maxCourses?: number
}

// --- API Functions ---

async function fetchAPI<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: "Request failed" }))
    throw new Error(error.error || `HTTP ${res.status}`)
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

export async function validateCourse(
  term: string,
  subject: string,
  courseNumber: string
): Promise<ValidateCourseResponse> {
  const params = new URLSearchParams({ term, subject, courseNumber })
  return fetchAPI<ValidateCourseResponse>(`/course/validate?${params}`)
}

export async function getCRN(crn: string, term?: string): Promise<CRNResponse> {
  const query = term ? `?term=${encodeURIComponent(term)}` : ""
  return fetchAPI<CRNResponse>(`/crn/${encodeURIComponent(crn)}${query}`)
}

export async function generateSchedules(req: GenerateRequest): Promise<GenerateResponse> {
  return fetchAPI<GenerateResponse>("/generate", {
    method: "POST",
    body: JSON.stringify(req),
  })
}
