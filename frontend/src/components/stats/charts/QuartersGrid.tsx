import type { StatsMetric } from "@/lib/api"
import { formatMetric, quartersGrid } from "@/lib/stats"
import { AXIS_TEXT, ChartFrame } from "./ChartFrame"

const CELL = 22
const GAP = 4
const LEFT = 26

/**
 * Which quarters the course ran in recent academic years, one row per
 * season with its overall value on the right. Upcoming terms are dashed.
 */
export function QuartersGrid({
  offered,
  next,
  currentTerm,
  seasons,
  metric,
}: {
  offered: string[]
  next?: string
  currentTerm: string
  seasons: { quarter: number; value: number | null }[]
  metric: StatsMetric
}) {
  const { years, rows } = quartersGrid(offered, next, currentTerm)
  const height = rows.length * (CELL + GAP) + 16
  const valueOf = (q: number) =>
    seasons.find((s) => s.quarter === q)?.value ?? null
  return (
    <ChartFrame height={height}>
      {(w) => {
        const right = w - 2
        return (
          <svg
            width={w}
            height={height}
            className="block"
            role="img"
            aria-label="Quarters offered"
          >
            <text
              x={right}
              y={height - 2}
              textAnchor="end"
              className={AXIS_TEXT}
            >
              {metric === "gpa" ? "GPA" : "Pass"}
            </text>
            {rows.map((r, i) => {
              const y = i * (CELL + GAP)
              const value = valueOf(r.quarter)
              return (
                <g key={r.quarter}>
                  <text x={0} y={y + 15} className={AXIS_TEXT}>
                    {r.label}
                  </text>
                  {value != null && (
                    <text
                      x={right}
                      y={y + 16}
                      textAnchor="end"
                      className="fill-foreground text-[13px] font-semibold tabular-nums"
                    >
                      {formatMetric(metric, value)}
                    </text>
                  )}
                  {r.cells.map((cell, c) => {
                    const x = LEFT + c * (CELL + GAP)
                    return cell === "upcoming" ? (
                      <rect
                        key={c}
                        x={x + 1}
                        y={y + 1}
                        width={CELL - 2}
                        height={CELL - 2}
                        rx={4}
                        fill="none"
                        className="stroke-foreground"
                        strokeWidth={1.5}
                        strokeDasharray="3 2.5"
                      />
                    ) : (
                      <rect
                        key={c}
                        x={x}
                        y={y}
                        width={CELL}
                        height={CELL}
                        rx={4}
                        className={
                          cell === "taught" ? "fill-foreground" : "fill-muted"
                        }
                        fillOpacity={cell === "taught" ? 0.8 : 1}
                      />
                    )
                  })}
                </g>
              )
            })}
            {years.map((yr, c) => (
              <text
                key={yr}
                x={LEFT + c * (CELL + GAP) + CELL / 2}
                y={height - 2}
                textAnchor="middle"
                className={AXIS_TEXT}
              >
                {String(yr).slice(2)}
              </text>
            ))}
          </svg>
        )
      }}
    </ChartFrame>
  )
}
