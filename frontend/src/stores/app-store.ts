import { create } from "zustand"
import { persist } from "zustand/middleware"

// --- Types ---

export type Tab = "schedule" | "search" | "statistics"
export type Theme = "light" | "dark" | "system"

export interface SectionFilter {
  crn: string
  term: string
  instructor: string | null
  required: boolean
}

export interface CourseSlot {
  id: string
  subject: string
  courseNumber: string
  displayName: string
  required: boolean
  sections: SectionFilter[] | null // null = all sections allowed
}

// StoredSchedule is the persisted version of a generated schedule.
// Differs from api.ts GeneratedSchedule which has full course details.
// TODO: we probably want to store the course slot details rather than
// just a string/section list. This means we have to refetch course details
// currently.
export interface StoredSchedule {
  sections: string[] // CRN list
  totalCredits: number
  score: number
}

// --- Store State ---

interface AppState {
  // Navigation
  tab: Tab

  // Sidebar state
  term: string
  selectedSubject: string
  minCourses: number | null
  maxCourses: number | null
  slots: CourseSlot[]

  // UI state
  theme: Theme
  sidebarCollapsed: boolean

  // Generated schedules (cleared on slot change)
  schedules: StoredSchedule[] | null
  currentScheduleIndex: number

  // Actions
  setTab: (tab: Tab) => void
  setTerm: (term: string) => void
  setSelectedSubject: (subject: string) => void
  setCourseBounds: (min: number | null, max: number | null) => void
  addSlot: (slot: CourseSlot) => void
  removeSlot: (id: string) => void
  updateSlot: (id: string, updates: Partial<CourseSlot>) => void
  clearSlots: () => void
  setTheme: (theme: Theme) => void
  setSidebarCollapsed: (collapsed: boolean) => void
  setSchedules: (schedules: StoredSchedule[] | null) => void
  setCurrentScheduleIndex: (index: number) => void
}

// --- Store ---

export const useAppStore = create<AppState>()(
  persist(
    (set) => ({
      // Initial state
      tab: "schedule",
      term: "",
      selectedSubject: "",
      minCourses: null,
      maxCourses: null,
      slots: [],
      theme: "system",
      sidebarCollapsed: false,
      schedules: null,
      currentScheduleIndex: 0,

      // Actions
      setTab: (tab) => set({ tab }),

      setTerm: (term) =>
        set({
          term,
          // Clear schedules when term changes
          schedules: null,
          currentScheduleIndex: 0,
        }),

      setSelectedSubject: (subject) => set({ selectedSubject: subject }),

      setCourseBounds: (min, max) =>
        set({
          minCourses: min,
          maxCourses: max,
        }),

      addSlot: (slot) =>
        set((state) => ({
          slots: [...state.slots, slot],
          // Clear schedules when slots change
          schedules: null,
          currentScheduleIndex: 0,
        })),

      removeSlot: (id) =>
        set((state) => ({
          slots: state.slots.filter((s) => s.id !== id),
          schedules: null,
          currentScheduleIndex: 0,
        })),

      updateSlot: (id, updates) =>
        set((state) => ({
          slots: state.slots.map((s) =>
            s.id === id ? { ...s, ...updates } : s
          ),
          schedules: null,
          currentScheduleIndex: 0,
        })),

      clearSlots: () =>
        set({
          slots: [],
          schedules: null,
          currentScheduleIndex: 0,
        }),

      setTheme: (theme) => set({ theme }),

      setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),

      setSchedules: (schedules) =>
        set({
          schedules,
          currentScheduleIndex: 0,
        }),

      setCurrentScheduleIndex: (index) => set({ currentScheduleIndex: index }),
    }),
    {
      name: "schedule-optimizer",
      partialize: (state) => ({
        // Only persist these fields
        tab: state.tab,
        term: state.term,
        selectedSubject: state.selectedSubject,
        minCourses: state.minCourses,
        maxCourses: state.maxCourses,
        slots: state.slots,
        theme: state.theme,
        sidebarCollapsed: state.sidebarCollapsed,
        // Don't persist generated schedules - they should be regenerated
      }),
    }
  )
)
