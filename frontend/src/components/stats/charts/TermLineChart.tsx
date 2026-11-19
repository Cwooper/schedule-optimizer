import type { StatsMetric } from "@/lib/api"
import { axisDomain, axisTicks, formatMetric, termName } from "@/lib/stats"
import { AXIS_TEXT, ChartFrame } from "./ChartFrame"

const HEIGHT = 200
const PAD = { r: 14, t: 8, b: 22 }
const MIN_LABEL_GAP = 44

/** Terms consecutively, so x spacing is true time and gaps show. */
function ordinal(term: string): number {
  return Number(term.slice(0, 4)) * 4 + Number(term.slice(4)) / 10 - 1
}

/** The metric per term, with the course's overall value dashed across. */
export function TermLineChart({
  points,
  mean,
  metric,
}: {
  points: { term: string; value: number }[]
  mean: number | null
  metric: StatsMetric
}) {
  const domain = axisDomain(metric, [
    ...points.map((p) => p.value),
    ...(mean == null ? [] : [mean]),
  ])
  const first = points.length ? ordinal(points[0].term) : 0
  const last = points.length ? ordinal(points[points.length - 1].term) : 0
  // "100%" needs more room than "4.0".
  const left = metric === "gpa" ? 30 : 36
  const tickLabel = (v: number) =>
    metric === "gpa" ? v.toFixed(1) : `${Math.round(v * 100)}%`
  return (
    <ChartFrame height={HEIGHT}>
      {(w) => {
        const span = Math.max(1, last - first)
        const x0 = left + 6
        const xw = w - left - PAD.r - 6
        const x = (term: string) =>
          points.length === 1
            ? x0 + xw / 2
            : x0 + ((ordinal(term) - first) / span) * xw
        const y = (v: number) =>
          PAD.t +
          (1 - (v - domain.lo) / (domain.hi - domain.lo)) *
            (HEIGHT - PAD.t - PAD.b)
        // Label calendar years at their Winter term, every other year or
        // sparser so labels never crowd.
        const perYear = (4 / span) * xw
        const stride = Math.max(1, Math.ceil(MIN_LABEL_GAP / perYear))
        const years: number[] = []
        for (let yr = Math.ceil((first + 1) / 4); yr * 4 <= last; yr++) {
          if (yr % stride === 0) years.push(yr)
        }
        const path = points
          .map((p, i) => `${i ? "L" : "M"}${x(p.term)} ${y(p.value)}`)
          .join("")
        return (
          <svg
            width={w}
            height={HEIGHT}
            className="block"
            role="img"
            aria-label="By term"
          >
            {axisTicks(domain).map((v) => (
              <g key={v}>
                <line
                  x1={left}
                  x2={w}
                  y1={y(v)}
                  y2={y(v)}
                  className="stroke-border"
                />
                <text
                  x={left - 8}
                  y={y(v) + 4}
                  textAnchor="end"
                  className={AXIS_TEXT}
                >
                  {tickLabel(v)}
                </text>
              </g>
            ))}
            {years.map((yr) => (
              <text
                key={yr}
                x={x(`${yr}10`)}
                y={HEIGHT - 6}
                textAnchor="middle"
                className={AXIS_TEXT}
              >
                {yr}
              </text>
            ))}
            {mean != null && (
              <line
                x1={left}
                x2={w}
                y1={y(mean)}
                y2={y(mean)}
                className="stroke-muted-foreground"
                strokeDasharray="3 4"
              />
            )}
            <path
              d={path}
              fill="none"
              className="stroke-foreground"
              strokeWidth={2}
              strokeLinejoin="round"
            />
            {points.map((p) => (
              <circle
                key={p.term}
                cx={x(p.term)}
                cy={y(p.value)}
                r={3}
                className="fill-foreground stroke-card"
                strokeWidth={1.5}
              >
                <title>{`${termName(p.term)}: ${formatMetric(metric, p.value)}`}</title>
              </circle>
            ))}
          </svg>
        )
      }}
    </ChartFrame>
  )
}
