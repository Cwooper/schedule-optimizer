import type { GradeCounts, GradeStats, StatsDot, StatsMetric } from "@/lib/api"
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

/**
 * A row's value against the course's, signed: "+0.13", "-0.39", "+3%". The
 * difference is taken between the rounded values so it matches the numbers
 * shown beside it.
 */
export function formatDelta(
  metric: StatsMetric,
  value: number,
  course: number
): string {
  const d = Math.round(value * 100) - Math.round(course * 100)
  let sign = ""
  if (d > 0) sign = "+"
  else if (d < 0) sign = "-"
  return metric === "gpa"
    ? `${sign}${(Math.abs(d) / 100).toFixed(2)}`
    : `${sign}${Math.abs(d)}%`
}

/** Each column's share of all the grades. */
export function gradeShares(grades: GradeCounts, columns: number[]): number[] {
  const total = grades.reduce((a, b) => a + b, 0) || 1
  return columns.map((c) => grades[c] / total)
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

export type DotKind = "current" | "mine" | "other"

export interface DotPoint extends StatsDot {
  kind: DotKind
  x: number
  y: number // in sections
}

export interface DotLayout {
  points: DotPoint[] // drawn in order, so the current section is last
  domain: AxisDomain // for ticks
  xDomain: [number, number] // the domain padded by half a bin
  yMax: number
  yTicks: number[]
  r: number
}

const MAX_DOT_PITCH = 12

/**
 * A stacked dot plot of sections: each dot sits in its value's bin (a fifth
 * of an axis step, so 0.1 GPA), stacked current first, then the
 * professor's, then the rest. A bin too tall for the plot wraps into
 * several columns of dots, each row holding the same number, so stack height
 * stays proportional to the count on the y axis.
 */
export function dotLayout(
  dots: StatsDot[],
  metric: StatsMetric,
  current: { term: string; crn: string },
  plot: { width: number; height: number }
): DotLayout {
  const domain = axisDomain(
    metric,
    dots.map((d) => d.value)
  )
  const bin = domain.step / 5
  const nBins = Math.round((domain.hi - domain.lo) / bin) + 1
  const binPx = plot.width / nBins
  const kindOf = (d: StatsDot): DotKind => {
    if (d.term === current.term && d.crn === current.crn) return "current"
    return d.mine ? "mine" : "other"
  }
  const rank = { current: 0, mine: 1, other: 2 }

  const bins = new Map<number, StatsDot[]>()
  for (const d of dots) {
    const i = Math.round((d.value - domain.lo) / bin)
    const members = bins.get(i)
    if (members) members.push(d)
    else bins.set(i, [d])
  }
  const tallest = Math.max(1, ...[...bins.values()].map((b) => b.length))

  let cols = 1
  let pitch = 0
  for (let c = 1; c <= tallest; c++) {
    const p = Math.min(
      binPx / c,
      plot.height / Math.ceil(tallest / c),
      MAX_DOT_PITCH
    )
    if (p > pitch) [cols, pitch] = [c, p]
  }

  const points: DotPoint[] = []
  for (const [i, members] of bins) {
    const center = domain.lo + i * bin
    members.sort((a, b) => rank[kindOf(a)] - rank[kindOf(b)])
    members.forEach((d, k) => {
      const row = Math.floor(k / cols)
      const inRow = Math.min(cols, members.length - row * cols)
      const col = k % cols
      points.push({
        ...d,
        kind: kindOf(d),
        x: center + (col - (inRow - 1) / 2) * pitch * (bin / binPx),
        y: (row + 0.5) * cols,
      })
    })
  }
  points.sort((a, b) => rank[b.kind] - rank[a.kind])

  // Dots touch vertically: the axis holds as many rows as fit.
  const yMax = (plot.height / pitch) * cols
  return {
    points,
    domain,
    xDomain: [domain.lo - bin / 2, domain.hi + bin / 2],
    yMax,
    yTicks: countTicks(yMax),
    r: Math.max(1, pitch / 2 - 0.75),
  }
}

/** Up to five whole-number ticks from 0, stepping by 1, 2, or 5 times a power of ten. */
export function countTicks(max: number): number[] {
  const top = Math.floor(max)
  for (let mag = 1; ; mag *= 10) {
    for (const step of [mag, 2 * mag, 5 * mag]) {
      if (top / step <= 4) {
        return Array.from(
          { length: Math.floor(top / step) + 1 },
          (_, i) => i * step
        )
      }
    }
  }
}

export interface QuarterCell {
  term: string
  state: "taught" | "upcoming" | "none"
}

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
      if (!offeredSet.has(term)) return { term, state: "none" }
      return { term, state: next && term >= next ? "upcoming" : "taught" }
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
