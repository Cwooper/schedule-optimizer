import type { ReactNode } from "react"
import type { GradeStats, StatsMetric } from "@/lib/api"
import {
  type AxisDomain,
  formatMetric,
  formatPct,
  metricValue,
} from "@/lib/stats"
import { TD, TH } from "./table"
import { cn } from "@/lib/utils"
import { GpaTrack } from "./charts/GpaTrack"
import { Sparkline } from "./charts/Sparkline"

export function StatTile({
  label,
  value,
  detail,
  children,
}: {
  label: string
  value: ReactNode
  detail?: ReactNode
  children?: ReactNode
}) {
  return (
    <div className="bg-card rounded-lg border px-3.5 py-3">
      <div className="text-muted-foreground text-xs leading-4">{label}</div>
      <div className="mt-0.5 text-[26px] leading-8 font-semibold tabular-nums">
        {value}
      </div>
      <div className="text-muted-foreground mt-0.5 truncate text-xs leading-4 tabular-nums">
        {detail ?? " "}
      </div>
      {children}
    </div>
  )
}

/** The muted "/ 40" after a tile value. */
export function TileOf({ children }: { children: ReactNode }) {
  return (
    <small className="text-muted-foreground text-[15px] font-medium">
      {" "}
      / {children}
    </small>
  )
}

export function Card({
  title,
  legend,
  className,
  children,
}: {
  title: ReactNode
  legend?: ReactNode
  className?: string
  children: ReactNode
}) {
  return (
    <section
      className={cn("bg-card min-w-0 rounded-lg border px-4 py-3.5", className)}
    >
      <h3 className="mb-2.5 flex items-center gap-3.5 text-[13px] leading-5 font-semibold">
        {title}
        {legend && (
          <span className="text-muted-foreground ml-auto flex gap-3 text-xs font-normal">
            {legend}
          </span>
        )}
      </h3>
      {children}
    </section>
  )
}

export function LegendItem({
  swatch,
  children,
}: {
  swatch: string
  children: ReactNode
}) {
  return (
    <span className="flex items-center gap-1.5 whitespace-nowrap">
      <i className={cn("inline-block size-2.5", swatch)} />
      {children}
    </span>
  )
}

/** In-app navigation styled as a link; there is no URL routing. */
export function LinkButton({
  onClick,
  className,
  children,
}: {
  onClick: () => void
  className?: string
  children: ReactNode
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "decoration-foreground/35 hover:decoration-foreground cursor-pointer font-medium underline underline-offset-3",
        className
      )}
    >
      {children}
    </button>
  )
}

/** Header cells matching StatCells. */
export function StatHeaders({ metric }: { metric: StatsMetric }) {
  return (
    <>
      <th className={cn(TH, "text-right")}>
        {metric === "gpa" ? "GPA" : "Pass"}
      </th>
      <th className={TH} aria-label="Compared to the course" />
      <th className={TH}>Distribution</th>
      {metric === "gpa" && <th className={cn(TH, "text-right")}>Pass</th>}
      <th className={cn(TH, "text-right")}>W</th>
    </>
  )
}

export interface RowScale {
  metric: StatsMetric
  columns: number[]
  mean: number | null
  domain: AxisDomain
}

/** The value, track, sparkline, pass, and W cells of a table row. */
export function StatCells({
  stats,
  scale,
  cellClass,
}: {
  stats: GradeStats
  scale: RowScale
  cellClass?: string
}) {
  const value = metricValue(stats, scale.metric)
  const td = cn(TD, cellClass)
  return (
    <>
      <td className={cn(td, "text-right font-semibold")}>
        {value != null && formatMetric(scale.metric, value)}
      </td>
      <td className={td}>
        {value != null && (
          <GpaTrack value={value} mean={scale.mean} domain={scale.domain} />
        )}
      </td>
      <td className={td}>
        {stats.grades && (
          <Sparkline grades={stats.grades} columns={scale.columns} />
        )}
      </td>
      {scale.metric === "gpa" && (
        <td className={cn(td, "text-right")}>
          {stats.pass != null && formatPct(stats.pass)}
        </td>
      )}
      <td className={cn(td, "text-right")}>
        {stats.withdrew != null && formatPct(stats.withdrew)}
      </td>
    </>
  )
}

export function StatsMessage({ error }: { error: Error }) {
  return (
    <p className="text-muted-foreground mt-10 text-center text-sm">
      {error.message}
    </p>
  )
}
