import { create } from "zustand"
import { persist } from "zustand/middleware"
import { FACTORS } from "@/lib/factors"
import { decodeHtmlEntities, genId } from "@/lib/utils"
import { mergeAdjacentBlocks } from "@/lib/schedule-utils"
import { DEFAULT_CALENDAR_EXPORT, type CalendarExportOptions } from "@/lib/ics"

// --- Types ---

export type Tab = "schedule" | "search" | "statistics"
export type Theme = "light" | "dark" | "system"
export type SearchScope = "term" | "year" | "all"

export interface SearchFilters {
  scope: SearchScope
  term: string // used when scope is "term"
  year: number | null // used when scope is "year"
  subject: string
  courseNumber: string
  title: string
  instructor: string
  openSeats: boolean
  minCredits: number | null
  maxCredits: number | null
}

export const DEFAULT_SEARCH_FILTERS: SearchFilters = {
  scope: "term",
  term: "",
  year: null,
  subject: "",
  courseNumber: "",
  title: "",
  instructor: "",
  openSeats: false,
  minCredits: null,
  maxCredits: null,
}

/** Convert SearchFilters to SearchRequest for API call */
export function filtersToSearchRequest(
  filters: SearchFilters
): import("@/lib/api").SearchRequest {
  const req: import("@/lib/api").SearchRequest = {}

  // Scope determines term/year
  if (filters.scope === "term" && filters.term) {
    req.term = filters.term
  } else if (filters.scope === "year" && filters.year !== null) {
    req.year = filters.year
  }
  // scope === "all" sends neither term nor year

  // Text filters - only include if non-empty
  if (filters.subject) req.subject = filters.subject
  if (filters.courseNumber) req.courseNumber = filters.courseNumber
  if (filters.title) req.title = filters.title
  if (filters.instructor) req.instructor = filters.instructor

  // Boolean/number filters
  if (filters.openSeats) req.openSeats = true
  if (filters.minCredits !== null) req.minCredits = filters.minCredits
  if (filters.maxCredits !== null) req.maxCredits = filters.maxCredits

  return req
}

export interface BlockedTimeBlock {
  id: string
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

/** Mirrors the backend's DefaultPreferences. */
export const DEFAULT_PREFERENCES: Preferences = {
  start: 2,
  end: 2,
  gap: 2,
  gpa: 3,
  seats: 0,
  idealStart: "1000",
  idealEnd: "1400",
}

/** Parameters that affect schedule generation - used for stale detection */
export interface GenerationParams {
  term: string
  minCourses: number
  maxCourses: number
  slotsFingerprint: string
  blockedTimesFingerprint: string
  preferencesFingerprint: string
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

export function computePreferencesFingerprint(p: Preferences): string {
  return [p.start, p.end, p.gap, p.gpa, p.seats, p.idealStart, p.idealEnd].join("-")
}

/**
 * Whether a result was generated with any weight on; with every weight off the
 * backend scores all schedules 0. Results saved before weights existed count
 * as ranked.
 */
export function wasRanked(params: GenerationParams | null): boolean {
  const fingerprint = params?.preferencesFingerprint
  if (!fingerprint) return true
  return fingerprint.split("-").slice(0, FACTORS.length).some((level) => level !== "0")
}

type GenerationInputs = Pick<
  AppState,
  "term" | "minCourses" | "maxCourses" | "slots" | "blockedTimeGroups" | "preferences"
>

/**
 * Fingerprints all slots, not just the ones valid this term, so the
 * fingerprint tracks user intent: a course that becomes valid later still
 * marks the result stale.
 */
export function computeGenerationParams(state: GenerationInputs): GenerationParams {
  return {
    term: state.term,
    minCourses: state.minCourses ?? 0,
    maxCourses: state.maxCourses ?? 0,
    slotsFingerprint: computeSlotsFingerprint(state.slots),
    blockedTimesFingerprint: computeBlockedTimesFingerprint(state.blockedTimeGroups),
    preferencesFingerprint: computePreferencesFingerprint(state.preferences),
  }
}

/** The params that differ from the ones a result was generated with. */
export function changedGenerationParams(
  current: GenerationParams,
  generated: GenerationParams
): (keyof GenerationParams)[] {
  return (Object.keys(current) as (keyof GenerationParams)[]).filter(
    (key) => current[key] !== generated[key]
  )
}

/** The latest generation request, until its response settles. */
export interface PendingGeneration {
  seq: number
  params: GenerationParams
  slotsVersion: number
  showWarnings: boolean
}

type WeightsCloseInputs = GenerationInputs &
  Pick<AppState, "pendingGeneration" | "generatedWithParams">

/**
 * Whether closing the weights panel should rerun generation. It compares
 * against the latest requested generation, in flight or shown, so closing
 * during a pending run is not skipped and an older request cannot win.
 */
export function weightsCloseRerun(
  state: WeightsCloseInputs,
  preferencesAtOpen: string
): { showWarnings: boolean } | null {
  const current = computeGenerationParams(state)
  // Opening and closing without an edit must not rerun a result whose saved
  // params predate preferences.
  if (current.preferencesFingerprint === preferencesAtOpen) return null
  const pending = state.pendingGeneration
  const latest = pending?.params ?? state.generatedWithParams
  if (!latest) return null
  const changed = changedGenerationParams(current, latest)
  if (!changed.includes("preferencesFingerprint")) return null
  // A weights-only change re-ranks the same request, whose warnings the user
  // has already seen, unless the request it supersedes was still due to show them.
  return { showWarnings: changed.length > 1 || pending?.showWarnings === true }
}

export function buildGenerateRequest(
  state: GenerationInputs,
  validSlots: CourseSlot[]
): GenerateRequest {
  const blockedTimes = state.blockedTimeGroups
    .filter((g) => g.enabled)
    .flatMap((g) => g.blocks)
    .map(({ day, startTime, endTime }) => ({ day, startTime, endTime }))
  return {
    term: state.term,
    courseSpecs: validSlots.map((slot) => ({
      subject: slot.subject,
      courseNumber: slot.courseNumber,
      required: slot.required,
      allowedCrns: slot.sections?.map((s) => s.crn),
    })),
    blockedTimes: blockedTimes.length > 0 ? blockedTimes : undefined,
    minCourses: state.minCourses ?? 0,
    maxCourses: state.maxCourses ?? 0,
    preferences: state.preferences,
  }
}

// Re-export from api.ts for convenience
export type {
  HydratedSection,
  ScheduleRef,
  GenerateResponse,
  SearchResponse,
  SearchRequest,
} from "@/lib/api"
import type {
  GenerateRequest,
  GenerateResponse,
  Importance,
  Preferences,
  SearchResponse,
} from "@/lib/api"

// --- Store State ---

interface CourseDialogState {
  open: boolean
  selectedCrn?: string
  selectedCourseKey?: string
  source?: "schedule" | "search"
}

export type SlotAddResult = "added" | "updated" | "duplicate"

/** The Statistics tab's view: a course, or one of its sections when term and crn are set. */
export interface StatsSelection {
  subject: string
  number: string
  term: string
  crn: string
}

const EMPTY_STATS_SELECTION: StatsSelection = { subject: "", number: "", term: "", crn: "" }

interface AppState {
  // Analytics
  sessionId: string

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
  calendarExport: CalendarExportOptions
  preferences: Preferences

  // Generated schedules
  generateResult: GenerateResponse | null
  generatedWithParams: GenerationParams | null
  currentScheduleIndex: number
  // Incremented on slot changes to detect stale mutation results
  slotsVersion: number
  generationSeq: number
  pendingGeneration: PendingGeneration | null
  // Flag to request regeneration from outside ScheduleBuilder
  regenerateRequested: boolean

  // Course info dialog state (not persisted)
  courseDialog: CourseDialogState

  // Search state
  searchFilters: SearchFilters
  searchResult: SearchResponse | null
  expandedSearchCourses: Set<string> // courseKeys that are expanded in results

  statsSelection: StatsSelection

  // Actions
  setTab: (tab: Tab) => void
  setTerm: (term: string) => void
  setSelectedSubject: (subject: string) => void
  setCourseBounds: (min: number | null, max: number | null) => void
  addSlot: (slot: CourseSlot) => void
  addCourseToSlot: (subject: string, courseNumber: string, title: string) => SlotAddResult
  addSectionToSlot: (crn: string, term: string, subject: string, courseNumber: string, title: string, instructor: string | null) => SlotAddResult
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
  updateBlockInGroupById: (groupId: string, blockId: string, block: BlockedTimeBlock) => void
  removeBlockFromGroupById: (groupId: string, blockId: string) => void
  setEditingBlockedTimeGroupId: (id: string | null) => void
  setTheme: (theme: Theme) => void
  setSidebarCollapsed: (collapsed: boolean) => void
  setCalendarExport: (options: Partial<CalendarExportOptions>) => void
  setPreferences: (updates: Partial<Preferences>) => void
  resetPreferences: () => void
  setGenerateResult: (
    result: GenerateResponse | null,
    params?: GenerationParams
  ) => void
  setCurrentScheduleIndex: (index: number) => void
  beginGeneration: (params: GenerationParams, showWarnings: boolean) => number
  /**
   * Settles a request: only the latest one applies its result (and only if
   * slots are unchanged since it began). Returns that request, or null when a
   * newer one superseded it.
   */
  settleGeneration: (seq: number, result?: GenerateResponse) => PendingGeneration | null
  isGenerateResultStale: () => boolean
  openCourseDialog: (opts: { crn?: string; courseKey?: string; source: "schedule" | "search" }) => void
  closeCourseDialog: () => void
  requestRegenerate: () => void
  clearRegenerateRequest: () => void

  // Search actions
  setSearchFilters: (filters: Partial<SearchFilters>) => void
  clearSearchFilters: () => void
  setSearchResult: (result: SearchResponse | null) => void
  toggleSearchCourseExpanded: (courseKey: string) => void
  clearExpandedSearchCourses: () => void

  setStatsSelection: (selection: Partial<StatsSelection>) => void
}

// --- Helpers ---

function updateBlockInGroup(
  groups: BlockedTimeGroup[],
  groupId: string,
  blockId: string,
  block: BlockedTimeBlock
): BlockedTimeGroup[] {
  return groups.map((g) => {
    if (g.id !== groupId) return g
    return { ...g, blocks: mergeAdjacentBlocks(g.blocks.map((b) => (b.id === blockId ? block : b))) }
  })
}

function removeBlockFromGroup(
  groups: BlockedTimeGroup[],
  groupId: string,
  blockId: string
): BlockedTimeGroup[] {
  return groups.map((g) => {
    if (g.id !== groupId) return g
    return { ...g, blocks: g.blocks.filter((b) => b.id !== blockId) }
  })
}

function decodeDeep(value: unknown, decode: (s: string) => string): unknown {
  if (typeof value === "string") return value.includes("&") ? decode(value) : value
  if (Array.isArray(value)) return value.map((v) => decodeDeep(v, decode))
  if (value && typeof value === "object") {
    return Object.fromEntries(Object.entries(value).map(([k, v]) => [k, decodeDeep(v, decode)]))
  }
  return value
}

/**
 * Before persist version 1 the API returned Banner text HTML-escaped and
 * components decoded it at render time; now the backend stores plain text.
 * Decode the course data older clients persisted so it does not render as
 * "B&amp;W".
 *
 * Before version 2 generated sections had no finals or meeting dates, so an
 * old result would show no finals and export classes without their own date
 * bounds. It is dropped; the slots it came from stay, ready to regenerate.
 */
export function migratePersistedState(
  state: unknown,
  version: number,
  decode: (s: string) => string = decodeHtmlEntities
): unknown {
  if (!state || typeof state !== "object") return state
  const out: Record<string, unknown> = { ...state }
  if (version < 1 && "slots" in out) out.slots = decodeDeep(out.slots, decode)
  if (version < 2) {
    out.generateResult = null
    out.generatedWithParams = null
    out.currentScheduleIndex = 0
  }
  return out
}

const IMPORTANCE_KEYS = FACTORS.map((f) => f.key)
const TIME_KEYS = ["idealStart", "idealEnd"] as const

/**
 * Keeps each stored preference only if the backend would accept it, so a
 * factor added later or a corrupted value falls back to its default instead
 * of failing every Generate.
 */
function storedPreferences(stored: unknown): Preferences {
  const raw = (stored && typeof stored === "object" ? stored : {}) as Record<string, unknown>
  const out = { ...DEFAULT_PREFERENCES }
  for (const key of IMPORTANCE_KEYS) {
    const v = raw[key]
    if (typeof v === "number" && Number.isInteger(v) && v >= 0 && v <= 3) {
      out[key] = v as Importance
    }
  }
  for (const key of TIME_KEYS) {
    const v = raw[key]
    if (typeof v === "string" && /^([01]\d|2[0-3])[0-5]\d$/.test(v)) out[key] = v
  }
  return out
}

export function mergePersistedState(persisted: unknown, current: AppState): AppState {
  const stored = (persisted ?? {}) as Partial<AppState>
  return {
    ...current,
    ...stored,
    preferences: storedPreferences(stored.preferences),
  }
}

/**
 * The term of the schedule on screen: the one it was generated for, which
 * stays shown (marked stale) after the sidebar switches terms.
 */
export function selectScheduleTerm(state: AppState): string {
  return (state.generateResult && state.generatedWithParams?.term) || state.term
}

export function selectInvalidBounds(state: AppState): boolean {
  return state.minCourses !== null && state.maxCourses !== null && state.maxCourses < state.minCourses
}

// --- Store ---

export const useAppStore = create<AppState>()(
  persist(
    (set, get) => ({
      // Initial state
      sessionId: genId(),
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
      calendarExport: DEFAULT_CALENDAR_EXPORT,
      preferences: DEFAULT_PREFERENCES,
      generateResult: null,
      generatedWithParams: null,
      currentScheduleIndex: 0,
      slotsVersion: 0,
      generationSeq: 0,
      pendingGeneration: null,
      regenerateRequested: false,
      courseDialog: { open: false },
      searchFilters: { ...DEFAULT_SEARCH_FILTERS },
      searchResult: null,
      expandedSearchCourses: new Set(),
      statsSelection: EMPTY_STATS_SELECTION,

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

      addCourseToSlot: (subject, courseNumber, title) => {
        const state = get()
        const existing = state.slots.find(
          (s) => s.subject === subject && s.courseNumber === courseNumber
        )
        if (existing) {
          if (existing.sections === null && (existing.title || !title)) {
            return "duplicate"
          }
          set({
            slots: state.slots.map((s) =>
              s.id === existing.id
                ? { ...s, sections: null, title: s.title || title }
                : s
            ),
            slotsVersion: state.slotsVersion + 1,
          })
          return "updated"
        }
        set({
          slots: [
            ...state.slots,
            {
              id: genId(),
              subject,
              courseNumber,
              displayName: `${subject} ${courseNumber}`,
              title,
              required: false,
              sections: null,
            },
          ],
          slotsVersion: state.slotsVersion + 1,
        })
        return "added"
      },

      addSectionToSlot: (crn, term, subject, courseNumber, title, instructor) => {
        const state = get()
        const sectionFilter: SectionFilter = {
          crn,
          term,
          instructor,
          required: true,
        }
        const existing = state.slots.find(
          (s) => s.subject === subject && s.courseNumber === courseNumber
        )
        if (existing) {
          const existingSections = existing.sections ?? []
          if (existingSections.some((s) => s.crn === crn && s.term === term)) {
            return "duplicate"
          }
          set({
            slots: state.slots.map((s) =>
              s.id === existing.id
                ? {
                    ...s,
                    sections: [...existingSections, sectionFilter],
                    title: s.title || title,
                  }
                : s
            ),
            slotsVersion: state.slotsVersion + 1,
          })
          return "updated"
        }
        set({
          slots: [
            ...state.slots,
            {
              id: genId(),
              subject,
              courseNumber,
              displayName: `${subject} ${courseNumber}`,
              title,
              required: false,
              sections: [sectionFilter],
            },
          ],
          slotsVersion: state.slotsVersion + 1,
        })
        return "added"
      },

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
            g.id === groupId
              ? { ...g, blocks: mergeAdjacentBlocks([...g.blocks, { ...block, id: block.id || genId() }]) }
              : g
          ),
        })),

      updateBlockInGroupById: (groupId, blockId, block) =>
        set((state) => ({
          blockedTimeGroups: updateBlockInGroup(state.blockedTimeGroups, groupId, blockId, block),
        })),

      removeBlockFromGroupById: (groupId, blockId) =>
        set((state) => ({
          blockedTimeGroups: removeBlockFromGroup(state.blockedTimeGroups, groupId, blockId),
        })),

      setEditingBlockedTimeGroupId: (id) =>
        set({ editingBlockedTimeGroupId: id }),

      setTheme: (theme) => set({ theme }),

      setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),

      setCalendarExport: (options) =>
        set((state) => ({ calendarExport: { ...state.calendarExport, ...options } })),

      setPreferences: (updates) =>
        set((state) => ({ preferences: { ...state.preferences, ...updates } })),

      resetPreferences: () => set({ preferences: DEFAULT_PREFERENCES }),

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

      beginGeneration: (params, showWarnings) => {
        const state = get()
        const seq = state.generationSeq + 1
        set({
          generationSeq: seq,
          pendingGeneration: { seq, params, slotsVersion: state.slotsVersion, showWarnings },
        })
        return seq
      },

      settleGeneration: (seq, result) => {
        const state = get()
        const pending = state.pendingGeneration
        if (pending?.seq !== seq) return null
        if (result && state.slotsVersion === pending.slotsVersion) {
          set({
            pendingGeneration: null,
            generateResult: result,
            generatedWithParams: pending.params,
            currentScheduleIndex: 0,
          })
        } else {
          set({ pendingGeneration: null })
        }
        return pending
      },

      isGenerateResultStale: () => {
        const state = get()
        if (!state.generateResult || !state.generatedWithParams) {
          return false
        }
        return (
          changedGenerationParams(computeGenerationParams(state), state.generatedWithParams)
            .length > 0
        )
      },

      openCourseDialog: (opts) =>
        set({
          courseDialog: {
            open: true,
            selectedCrn: opts.crn,
            selectedCourseKey: opts.courseKey,
            source: opts.source,
          },
        }),

      closeCourseDialog: () =>
        set((state) => ({
          courseDialog: { ...state.courseDialog, open: false },
        })),

      requestRegenerate: () => set({ regenerateRequested: true }),

      clearRegenerateRequest: () => set({ regenerateRequested: false }),

      // Search actions
      setSearchFilters: (filters) =>
        set((state) => ({
          searchFilters: { ...state.searchFilters, ...filters },
        })),

      clearSearchFilters: () =>
        set({
          searchFilters: { ...DEFAULT_SEARCH_FILTERS },
          searchResult: null,
          expandedSearchCourses: new Set(),
        }),

      setSearchResult: (result) => set({ searchResult: result }),

      toggleSearchCourseExpanded: (courseKey) =>
        set((state) => {
          const next = new Set(state.expandedSearchCourses)
          if (next.has(courseKey)) {
            next.delete(courseKey)
          } else {
            next.add(courseKey)
          }
          return { expandedSearchCourses: next }
        }),

      clearExpandedSearchCourses: () =>
        set({ expandedSearchCourses: new Set() }),

      setStatsSelection: (selection) =>
        set((state) => ({ statsSelection: { ...state.statsSelection, ...selection } })),
    }),
    {
      name: "schedule-optimizer",
      version: 2,
      migrate: (state, version) => migratePersistedState(state, version) as AppState,
      merge: mergePersistedState,
      partialize: (state) => ({
        // Only persist these fields
        sessionId: state.sessionId,
        tab: state.tab,
        term: state.term,
        selectedSubject: state.selectedSubject,
        minCourses: state.minCourses,
        maxCourses: state.maxCourses,
        slots: state.slots,
        blockedTimeGroups: state.blockedTimeGroups,
        theme: state.theme,
        sidebarCollapsed: state.sidebarCollapsed,
        calendarExport: state.calendarExport,
        preferences: state.preferences,
        generateResult: state.generateResult,
        generatedWithParams: state.generatedWithParams,
        currentScheduleIndex: state.currentScheduleIndex,
        searchFilters: state.searchFilters,
        statsSelection: state.statsSelection,
        // Note: searchResult and expandedSearchCourses are NOT persisted
      }),
    }
  )
)
