import { create } from "zustand"
import { persist } from "zustand/middleware"
import { genId } from "@/lib/utils"

// --- Types ---

export type Tab = "schedule" | "search" | "statistics"
export type Theme = "light" | "dark" | "system"

export interface BlockedTimeBlock {
  day: number
  startTime: string
  endTime: string
}

export interface BlockedTimeGroup {
  id: string
  title: string // user-set label; empty = auto "Blocked Time N" in dialog only
  description: string
  color: string | null // null = no color, string = palette key or hex
  hatched: boolean // show hatched overlay pattern
  opacity: number // 10-80, background opacity percentage when color is set
  enabled: boolean // ON = visible on grid + used in generation
  blocks: BlockedTimeBlock[]
}

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

/** Parameters that affect schedule generation - used for stale detection */
export interface GenerationParams {
  term: string
  minCourses: number
  maxCourses: number
  slotsFingerprint: string
  blockedTimesFingerprint: string
}

/**
 * Computes a content-based fingerprint of slots for stale detection.
 * Excludes slot.id so re-adding the same course clears stale state.
 * Sorts by course key for stable comparison regardless of UI order.
 */
export function computeSlotsFingerprint(slots: CourseSlot[]): string {
  const normalized = slots
    .map((s) => ({
      key: `${s.subject}-${s.courseNumber}`,
      required: s.required,
      sections:
        s.sections
          ?.map((sec) => ({ crn: sec.crn, required: sec.required }))
          .sort((a, b) => a.crn.localeCompare(b.crn)) ?? null,
    }))
    .sort((a, b) => a.key.localeCompare(b.key))
  return JSON.stringify(normalized)
}

export function computeBlockedTimesFingerprint(
  groups: BlockedTimeGroup[]
): string {
  const enabled = groups
    .filter((g) => g.enabled)
    .flatMap((g) => g.blocks)
    .map((b) => `${b.day}-${b.startTime}-${b.endTime}`)
    .sort()
  return JSON.stringify(enabled)
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

  // Blocked times
  blockedTimeGroups: BlockedTimeGroup[]
  editingBlockedTimeGroupId: string | null

  // UI state
  theme: Theme
  sidebarCollapsed: boolean

  // Generated schedules (not persisted)
  generateResult: GenerateResponse | null
  generatedWithParams: GenerationParams | null
  currentScheduleIndex: number
  // Incremented on slot changes to detect stale mutation results
  slotsVersion: number
  // Flag to request regeneration from outside ScheduleBuilder
  regenerateRequested: boolean

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
  addBlockedTimeGroup: () => string
  removeBlockedTimeGroup: (id: string) => void
  updateBlockedTimeGroup: (
    id: string,
    updates: Partial<BlockedTimeGroup>
  ) => void
  addBlockToGroup: (groupId: string, block: BlockedTimeBlock) => void
  removeBlockFromGroup: (groupId: string, blockIndex: number) => void
  setEditingBlockedTimeGroupId: (id: string | null) => void
  setTheme: (theme: Theme) => void
  setSidebarCollapsed: (collapsed: boolean) => void
  setGenerateResult: (
    result: GenerateResponse | null,
    params?: GenerationParams
  ) => void
  setCurrentScheduleIndex: (index: number) => void
  getSlotsVersion: () => number
  isGenerateResultStale: () => boolean
  openCourseDialog: (opts: { crn?: string; courseKey?: string }) => void
  closeCourseDialog: () => void
  requestRegenerate: () => void
  clearRegenerateRequest: () => void
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
      blockedTimeGroups: [],
      editingBlockedTimeGroupId: null,
      theme: "system",
      sidebarCollapsed: false,
      generateResult: null,
      generatedWithParams: null,
      currentScheduleIndex: 0,
      slotsVersion: 0,
      regenerateRequested: false,
      courseDialog: { open: false },

      // Actions
      setTab: (tab) => set({ tab }),

      setTerm: (term) =>
        set((state) => ({
          term,
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
          slotsVersion: state.slotsVersion + 1,
        })),

      removeSlot: (id) =>
        set((state) => ({
          slots: state.slots.filter((s) => s.id !== id),
          slotsVersion: state.slotsVersion + 1,
        })),

      updateSlot: (id, updates) =>
        set((state) => ({
          slots: state.slots.map((s) =>
            s.id === id ? { ...s, ...updates } : s
          ),
          slotsVersion: state.slotsVersion + 1,
        })),

      clearSlots: () =>
        set((state) => ({
          slots: [],
          generateResult: null,
          generatedWithParams: null,
          currentScheduleIndex: 0,
          slotsVersion: state.slotsVersion + 1,
        })),

      addBlockedTimeGroup: () => {
        const id = genId()
        const state = get()
        set({
          blockedTimeGroups: [
            ...state.blockedTimeGroups,
            {
              id,
              title: "",
              description: "",
              color: null,
              hatched: true,
              opacity: 20,
              enabled: true,
              blocks: [],
            },
          ],
        })
        return id
      },

      removeBlockedTimeGroup: (id) =>
        set((state) => ({
          blockedTimeGroups: state.blockedTimeGroups.filter((g) => g.id !== id),
          editingBlockedTimeGroupId:
            state.editingBlockedTimeGroupId === id
              ? null
              : state.editingBlockedTimeGroupId,
        })),

      updateBlockedTimeGroup: (id, updates) =>
        set((state) => ({
          blockedTimeGroups: state.blockedTimeGroups.map((g) =>
            g.id === id ? { ...g, ...updates } : g
          ),
        })),

      addBlockToGroup: (groupId, block) =>
        set((state) => ({
          blockedTimeGroups: state.blockedTimeGroups.map((g) =>
            g.id === groupId ? { ...g, blocks: [...g.blocks, block] } : g
          ),
        })),

      removeBlockFromGroup: (groupId, blockIndex) =>
        set((state) => {
          const updated = state.blockedTimeGroups.map((g) => {
            if (g.id !== groupId) return g
            const blocks = [...g.blocks.slice(0, blockIndex), ...g.blocks.slice(blockIndex + 1)]
            return { ...g, blocks }
          })
          return { blockedTimeGroups: updated }
        }),

      setEditingBlockedTimeGroupId: (id) =>
        set({ editingBlockedTimeGroupId: id }),

      setTheme: (theme) => set({ theme }),

      setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),

      setGenerateResult: (result, params) =>
        set({
          generateResult: result,
          generatedWithParams: params ?? null,
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

      isGenerateResultStale: () => {
        const state = get()
        if (!state.generateResult || !state.generatedWithParams) {
          return false
        }
        const current: GenerationParams = {
          term: state.term,
          minCourses: state.minCourses ?? state.slots.length,
          maxCourses: state.maxCourses ?? 8,
          slotsFingerprint: computeSlotsFingerprint(state.slots),
          blockedTimesFingerprint: computeBlockedTimesFingerprint(
            state.blockedTimeGroups
          ),
        }
        return (
          current.term !== state.generatedWithParams.term ||
          current.minCourses !== state.generatedWithParams.minCourses ||
          current.maxCourses !== state.generatedWithParams.maxCourses ||
          current.slotsFingerprint !==
            state.generatedWithParams.slotsFingerprint ||
          current.blockedTimesFingerprint !==
            state.generatedWithParams.blockedTimesFingerprint
        )
      },

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

      requestRegenerate: () => set({ regenerateRequested: true }),

      clearRegenerateRequest: () => set({ regenerateRequested: false }),
    }),
    {
      name: "schedule-optimizer",
      version: 1,
      partialize: (state) => ({
        // Only persist these fields
        tab: state.tab,
        term: state.term,
        selectedSubject: state.selectedSubject,
        minCourses: state.minCourses,
        maxCourses: state.maxCourses,
        slots: state.slots,
        blockedTimeGroups: state.blockedTimeGroups,
        theme: state.theme,
        sidebarCollapsed: state.sidebarCollapsed,
        generateResult: state.generateResult,
        generatedWithParams: state.generatedWithParams,
        currentScheduleIndex: state.currentScheduleIndex,
      }),
      migrate: (persisted, version) => {
        const state = persisted as Record<string, unknown>
        if (version === 0) {
          // v0 → v1: Add hatched + opacity to blocked time groups
          const groups = (state.blockedTimeGroups ?? []) as Record<string, unknown>[]
          state.blockedTimeGroups = groups.map((g) => ({
            ...g,
            hatched: g.hatched ?? (g.color === null),
            opacity: g.opacity ?? 20,
          }))
        }
        return state as unknown as AppState
      },
    }
  )
)
