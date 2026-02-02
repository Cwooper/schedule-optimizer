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
  title?: string // e.g., "Computer Systems"
  required: boolean
  sections: SectionFilter[] | null // null = all sections allowed
}

// Re-export from api.ts for convenience
export type { HydratedSection, ScheduleRef, GenerateResponse } from "@/lib/api"
import type { GenerateResponse } from "@/lib/api"

// --- Store State ---

interface CourseDialogState {
  open: boolean
  selectedCrn?: string
  selectedCourseKey?: string
}

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

  // Generated schedules (not persisted)
  generateResult: GenerateResponse | null
  isGenerateResultStale: boolean
  currentScheduleIndex: number
  // Incremented on slot changes to detect stale mutation results
  slotsVersion: number

  // Course info dialog state (not persisted)
  courseDialog: CourseDialogState

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
  setGenerateResult: (result: GenerateResponse | null) => void
  setCurrentScheduleIndex: (index: number) => void
  getSlotsVersion: () => number
  openCourseDialog: (opts: { crn?: string; courseKey?: string }) => void
  closeCourseDialog: () => void
}

// --- Store ---

export const useAppStore = create<AppState>()(
  persist(
    (set, get) => ({
      // Initial state
      tab: "schedule",
      term: "",
      selectedSubject: "",
      minCourses: null,
      maxCourses: null,
      slots: [],
      theme: "system",
      sidebarCollapsed: false,
      generateResult: null,
      isGenerateResultStale: false,
      currentScheduleIndex: 0,
      slotsVersion: 0,
      courseDialog: { open: false },

      // Actions
      setTab: (tab) => set({ tab }),

      setTerm: (term) =>
        set((state) => ({
          term,
          // Mark schedules stale when term changes
          isGenerateResultStale: true,
          slotsVersion: state.slotsVersion + 1,
        })),

      setSelectedSubject: (subject) => set({ selectedSubject: subject }),

      setCourseBounds: (min, max) =>
        set({
          minCourses: min,
          maxCourses: max,
        }),

      addSlot: (slot) =>
        set((state) => ({
          slots: [...state.slots, slot],
          isGenerateResultStale: true,
          slotsVersion: state.slotsVersion + 1,
        })),

      removeSlot: (id) =>
        set((state) => ({
          slots: state.slots.filter((s) => s.id !== id),
          isGenerateResultStale: true,
          slotsVersion: state.slotsVersion + 1,
        })),

      updateSlot: (id, updates) =>
        set((state) => ({
          slots: state.slots.map((s) =>
            s.id === id ? { ...s, ...updates } : s
          ),
          isGenerateResultStale: true,
          slotsVersion: state.slotsVersion + 1,
        })),

      clearSlots: () =>
        set((state) => ({
          slots: [],
          generateResult: null,
          isGenerateResultStale: false,
          currentScheduleIndex: 0,
          slotsVersion: state.slotsVersion + 1,
        })),

      setTheme: (theme) => set({ theme }),

      setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),

      setGenerateResult: (result) =>
        set({
          generateResult: result,
          isGenerateResultStale: false,
          currentScheduleIndex: 0,
        }),

      setCurrentScheduleIndex: (index) =>
        set((state) => {
          const maxIndex = state.generateResult
            ? state.generateResult.schedules.length - 1
            : 0
          return { currentScheduleIndex: Math.max(0, Math.min(index, maxIndex)) }
        }),

      getSlotsVersion: () => get().slotsVersion,

      openCourseDialog: (opts) =>
        set({
          courseDialog: {
            open: true,
            selectedCrn: opts.crn,
            selectedCourseKey: opts.courseKey,
          },
        }),

      closeCourseDialog: () =>
        set((state) => ({
          courseDialog: { ...state.courseDialog, open: false },
        })),
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
