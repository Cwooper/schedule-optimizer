import type { StatsDot, StatsMetric } from "@/lib/api"
import {
  axisDomain,
  axisTicks,
  formatMetric,
  jitter,
  termName,
} from "@/lib/stats"
import { AXIS_TEXT, ChartFrame } from "./ChartFrame"

const HEIGHT = 200
const PAD = { l: 8, r: 8, t: 14, b: 26 }

/**
 * Every reported section of a course on one axis; the current section is
 * ringed and its professor's are dark.
 */
export function DotStrip({
  dots,
  current,
  metric,
}: {
  dots: StatsDot[]
  current: { term: string; crn: string }
  metric: StatsMetric
}) {
  const domain = axisDomain(
    metric,
    dots.map((d) => d.value)
  )
  const isCurrent = (d: StatsDot) =>
    d.term === current.term && d.crn === current.crn
  const me = dots.find(isCurrent)
  // Others first, so the professor's dots sit on top.
  const ordered = dots
    .filter((d) => !isCurrent(d))
    .sort((a, b) => Number(a.mine ?? false) - Number(b.mine ?? false))
  const tickLabel = (v: number) =>
    metric === "gpa" ? v.toFixed(1) : `${Math.round(v * 100)}%`
  return (
    <ChartFrame height={HEIGHT}>
      {(w) => {
        const x = (v: number) =>
          PAD.l +
          ((v - domain.lo) / (domain.hi - domain.lo)) * (w - PAD.l - PAD.r)
        const mid = PAD.t + (HEIGHT - PAD.t - PAD.b) / 2
        const spread = (HEIGHT - PAD.t - PAD.b - 30) / 2
        return (
          <svg
            width={w}
            height={HEIGHT}
            className="block"
            role="img"
            aria-label="Sections compared"
          >
            {axisTicks(domain).map((v) => (
              <g key={v}>
                <line
                  x1={x(v)}
                  x2={x(v)}
                  y1={PAD.t}
                  y2={HEIGHT - PAD.b}
                  className="stroke-border"
                />
                <text
                  x={x(v)}
                  y={HEIGHT - 8}
                  textAnchor="middle"
                  className={AXIS_TEXT}
                >
                  {tickLabel(v)}
                </text>
              </g>
            ))}
            {ordered.map((d) => (
              <circle
                key={`${d.term}${d.crn}`}
                cx={x(d.value)}
                cy={mid + jitter(`${d.term}${d.crn}`) * spread}
                r={5}
                className={d.mine ? "fill-foreground" : "fill-chart-soft"}
                fillOpacity={d.mine ? 0.75 : 1}
              >
                <title>{`${termName(d.term)}, CRN ${d.crn}: ${formatMetric(metric, d.value)}`}</title>
              </circle>
            ))}
            {me && (
              <>
                <line
                  x1={x(me.value)}
                  x2={x(me.value)}
                  y1={PAD.t}
                  y2={HEIGHT - PAD.b}
                  className="stroke-foreground"
                  strokeWidth={1.5}
                />
                <circle
                  cx={x(me.value)}
                  cy={mid}
                  r={8}
                  className="fill-card stroke-foreground"
                  strokeWidth={3}
                />
              </>
            )}
          </svg>
        )
      }}
    </ChartFrame>
  )
}
