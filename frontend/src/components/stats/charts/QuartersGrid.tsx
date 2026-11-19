import type { StatsCourse, StatsMetric } from "@/lib/api"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import {
  formatMetric,
  type QuarterCell,
  quartersGrid,
  termName,
} from "@/lib/stats"
import { cn } from "@/lib/utils"
import { TipBody, type TipRow } from "./ChartTip"

const AXIS = "text-muted-foreground text-[11px] leading-4"

/**
 * Which quarters the course ran in recent academic years, one row per
 * season with its overall value on the right. Upcoming terms are dashed.
 */
export function QuartersGrid({
  course,
  currentTerm,
}: {
  course: Pick<StatsCourse, "offered" | "terms" | "seasons" | "next" | "metric">
  currentTerm: string
}) {
  const { metric } = course
  const { years, rows } = quartersGrid(
    course.offered.map((o) => o.term),
    course.next?.term,
    currentTerm
  )
  const valueOf = (q: number) =>
    course.seasons.find((s) => s.quarter === q)?.value ?? null
  return (
    <div
      className="grid items-center gap-1"
      style={{ gridTemplateColumns: `22px repeat(${years.length}, 22px) 1fr` }}
    >
      {rows.map((r) => {
        const value = valueOf(r.quarter)
        return [
          <span key={`${r.quarter}l`} className={AXIS}>
            {r.label}
          </span>,
          ...r.cells.map((cell) => (
            <QuarterSquare
              key={cell.term}
              cell={cell}
              course={course}
              metric={metric}
            />
          )),
          <span
            key={`${r.quarter}v`}
            className="text-right text-[13px] font-semibold tabular-nums"
          >
            {value != null && formatMetric(metric, value)}
          </span>,
        ]
      })}
      <span />
      {years.map((yr) => (
        <span key={yr} className={cn(AXIS, "text-center")}>
          {String(yr).slice(2)}
        </span>
      ))}
      <span className={cn(AXIS, "text-right")}>
        {metric === "gpa" ? "GPA" : "Pass"}
      </span>
    </div>
  )
}

function QuarterSquare({
  cell,
  course,
  metric,
}: {
  cell: QuarterCell
  course: Pick<StatsCourse, "offered" | "terms">
  metric: StatsMetric
}) {
  const graded = course.terms.find((t) => t.term === cell.term)
  const offered = course.offered.find((o) => o.term === cell.term)
  const rows: TipRow[] = []
  if (graded?.value != null) {
    rows.push([
      metric === "gpa" ? "GPA" : "Pass",
      formatMetric(metric, graded.value),
    ])
  }
  const sections = graded?.sections ?? offered?.sections
  if (sections) rows.push(["Sections", sections])
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span
          className={cn(
            "block size-[22px] rounded-[4px]",
            cell.state === "taught" && "bg-foreground/80",
            cell.state === "none" && "bg-muted"
          )}
        >
          {cell.state === "upcoming" && (
            <svg width="22" height="22" className="block" aria-hidden>
              <rect
                x="1"
                y="1"
                width="20"
                height="20"
                rx="4"
                fill="none"
                className="stroke-foreground"
                strokeWidth={1.5}
                strokeDasharray="3 2.5"
              />
            </svg>
          )}
        </span>
      </TooltipTrigger>
      <TooltipContent>
        {cell.state === "none" ? (
          <TipBody title={termName(cell.term)} rows={[["Not offered"]]} />
        ) : (
          <TipBody title={termName(cell.term)} rows={rows} />
        )}
      </TooltipContent>
    </Tooltip>
  )
}
