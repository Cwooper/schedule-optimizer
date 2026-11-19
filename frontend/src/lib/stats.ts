import type { GradeCounts, GradeStats, StatsMetric } from "@/lib/api"
import { getAcademicYear } from "@/lib/schedule-utils"

/** Labels for GradeCounts indexes. */
export const GRADE_LABELS = [
  "A",
  "A-",
  "B+",
  "B",
  "B-",
  "C+",
  "C",
  "C-",
  "D+",
  "D",
  "D-",
  "F",
  "W",
  "P",
  "NP",
  "S",
  "U",
] as const

const LETTERS = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]
const W = 12
const PASS_FAIL = [13, 14]
const SAT_UNSAT = [15, 16]

/**
 * The grade columns a course's charts show: letters, P/NP, and S/U each only
 * when the course has any, then W last. Every chart of a course shares them
 * so bars line up across rows.
 */
export function distributionColumns(course?: GradeCounts): number[] {
  if (!course) return [...LETTERS, W]
  const any = (idx: number[]) => idx.some((i) => course[i] > 0)
  return [
    ...(any(LETTERS) ? LETTERS : []),
    ...(any(PASS_FAIL) ? PASS_FAIL : []),
    ...(any(SAT_UNSAT) ? SAT_UNSAT : []),
    W,
  ]
}

export const WITHDREW_COLUMN = W

const QUARTERS: Record<string, [string, string]> = {
  "10": ["Winter", "Win"],
  "20": ["Spring", "Spr"],
  "30": ["Summer", "Sum"],
  "40": ["Fall", "Fall"],
}

/** "202340" -> "Fall 2023". */
export function termName(code: string): string {
  const q = QUARTERS[code.slice(4)]
  return q ? `${q[0]} ${code.slice(0, 4)}` : code
}

/** "202710" -> "Win 27", for tags. */
export function termShort(code: string): string {
  const q = QUARTERS[code.slice(4)]
  return q ? `${q[1]} ${code.slice(2, 4)}` : code
}

export function formatPct(v: number, digits = 0): string {
  return `${(v * 100).toFixed(digits)}%`
}

/** A row's comparison value; pass rates of unreported rows stay hidden too. */
export function metricValue(
  stats: GradeStats,
  metric: StatsMetric
): number | null {
  if (metric === "gpa") return stats.gpa
  return stats.grades ? stats.pass : null
}

export function formatMetric(metric: StatsMetric, v: number): string {
  return metric === "gpa" ? v.toFixed(2) : formatPct(v)
}

/** "Smith, Jordan" -> "Smith". */
export function surname(name: string): string {
  return name.split(",")[0].trim()
}

export interface AxisDomain {
  lo: number
  hi: number
  step: number
}

/**
 * GPA axes run from 4.0 down to 2.0, or lower in half steps when a value
 * falls below. Pass axes do the same from 100% down to 80% in tenths.
 */
export function axisDomain(metric: StatsMetric, values: number[]): AxisDomain {
  const low = values.length ? Math.min(...values) : Infinity
  if (metric === "gpa") {
    return {
      lo: Math.max(0, Math.min(2, Math.floor(low * 2) / 2)),
      hi: 4,
      step: 0.5,
    }
  }
  const lo = Math.max(0, Math.min(0.8, Math.floor(low * 10) / 10))
  return { lo, hi: 1, step: 1 - lo > 0.4 ? 0.2 : 0.1 }
}

/** The ticks from lo to hi, inclusive, without float drift. */
export function axisTicks({ lo, hi, step }: AxisDomain): number[] {
  const n = Math.round((hi - lo) / step)
  return Array.from(
    { length: n + 1 },
    (_, i) => Math.round((lo + i * step) * 1000) / 1000
  )
}

/**
 * A stable vertical offset in [-1, 1] for a dot, so the strip never
 * reshuffles. FNV-1a with a murmur finalizer: consecutive CRNs differ in one
 * digit, and a plain string hash would stack their dots.
 */
export function jitter(key: string): number {
  let h = 2166136261
  for (let i = 0; i < key.length; i++) {
    h = Math.imul(h ^ key.charCodeAt(i), 16777619)
  }
  h = Math.imul(h ^ (h >>> 16), 0x85ebca6b)
  h = Math.imul(h ^ (h >>> 13), 0xc2b2ae35)
  h ^= h >>> 16
  return ((h >>> 0) / 0xffffffff) * 2 - 1
}

export type QuarterCell = "taught" | "upcoming" | "none"

export interface QuarterRow {
  quarter: number
  label: string
  cells: QuarterCell[]
}

const GRID_ROWS: [number, string][] = [
  [40, "F"],
  [10, "W"],
  [20, "S"],
  [30, "Su"],
]

/**
 * The quarters grid: a row per season, a column per academic year (Fall
 * through Summer), ending with the later of the current year and the year
 * of the next scheduled term. Offered terms from the next one on are
 * upcoming.
 */
export function quartersGrid(
  offered: string[],
  next: string | undefined,
  currentTerm: string,
  columns = 6
): { years: number[]; rows: QuarterRow[] } {
  const end = Math.max(
    getAcademicYear(currentTerm) ?? 0,
    next ? (getAcademicYear(next) ?? 0) : 0
  )
  const years = Array.from({ length: columns }, (_, i) => end - columns + 1 + i)
  const offeredSet = new Set(offered)
  const rows = GRID_ROWS.map(([quarter, label]) => ({
    quarter,
    label,
    cells: years.map((ay): QuarterCell => {
      const term = `${quarter === 40 ? ay - 1 : ay}${quarter}`
      if (!offeredSet.has(term)) return "none"
      return next && term >= next ? "upcoming" : "taught"
    }),
  }))
  return { years, rows }
}

/** The term a date falls in, by month alone, like the backend's termcode.Current. */
export function currentTermCode(now = new Date()): string {
  const m = now.getMonth() + 1
  let q = 10
  if (m >= 9) q = 40
  else if (m >= 6) q = 30
  else if (m >= 4) q = 20
  return `${now.getFullYear()}${q}`
}
