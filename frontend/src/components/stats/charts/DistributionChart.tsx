import type { GradeCounts } from "@/lib/api"
import { GRADE_LABELS, WITHDREW_COLUMN, axisTicks } from "@/lib/stats"
import { AXIS_TEXT, ChartFrame } from "./ChartFrame"

const HEIGHT = 200
const PAD = { l: 34, t: 8, b: 22 }

function shares(grades: GradeCounts, columns: number[]): number[] {
  const total = grades.reduce((a, b) => a + b, 0) || 1
  return columns.map((c) => grades[c] / total)
}

function gridStep(max: number): number {
  if (max <= 0.3) return 0.05
  if (max <= 0.6) return 0.1
  return 0.2
}

/**
 * Share of students per grade. A background series (the course, behind a
 * section) draws wider and soft so both read at once.
 */
export function DistributionChart({
  grades,
  background,
  columns,
}: {
  grades?: GradeCounts
  background?: GradeCounts
  columns: number[]
}) {
  const fg = grades ? shares(grades, columns) : []
  const bg = background ? shares(background, columns) : []
  const top = Math.max(...fg, ...bg, 0.01) * 1.05
  const step = gridStep(top)
  const max = Math.ceil(top / step) * step
  const ticks = axisTicks({ lo: 0, hi: max, step })
  return (
    <ChartFrame height={HEIGHT}>
      {(w) => {
        const cw = (w - PAD.l) / columns.length
        // A fixed 10px gap until columns get narrow on phones, then a share.
        const bw = Math.min(28, Math.max(cw - 10, cw * 0.72))
        const bgw = Math.min(bw + 8, cw - 2)
        const fw = bg.length ? Math.max(bw - 8, bw * 0.55) : bw
        const h = (v: number) => (v / max) * (HEIGHT - PAD.t - PAD.b)
        const base = HEIGHT - PAD.b
        return (
          <svg
            width={w}
            height={HEIGHT}
            className="block"
            role="img"
            aria-label="Grade distribution"
          >
            {ticks.map((v) => (
              <g key={v}>
                <line
                  x1={PAD.l}
                  x2={w}
                  y1={base - h(v)}
                  y2={base - h(v)}
                  className="stroke-border"
                />
                <text
                  x={PAD.l - 8}
                  y={base - h(v) + 4}
                  textAnchor="end"
                  className={AXIS_TEXT}
                >
                  {Math.round(v * 100)}%
                </text>
              </g>
            ))}
            {columns.map((c, i) => {
              const cx = PAD.l + i * cw + cw / 2
              return (
                <g key={c}>
                  {bg.length > 0 && (
                    <rect
                      x={cx - bgw / 2}
                      y={base - h(bg[i])}
                      width={bgw}
                      height={h(bg[i])}
                      rx={3}
                      className="fill-chart-soft"
                    />
                  )}
                  {fg.length > 0 && (
                    <rect
                      x={cx - fw / 2}
                      y={base - h(fg[i])}
                      width={fw}
                      height={h(fg[i])}
                      rx={3}
                      className="fill-foreground"
                      fillOpacity={c === WITHDREW_COLUMN ? 0.45 : 1}
                    >
                      <title>{`${GRADE_LABELS[c]}: ${(fg[i] * 100).toFixed(1)}%`}</title>
                    </rect>
                  )}
                  <text
                    x={cx}
                    y={HEIGHT - 6}
                    textAnchor="middle"
                    className={AXIS_TEXT}
                  >
                    {GRADE_LABELS[c]}
                  </text>
                </g>
              )
            })}
          </svg>
        )
      }}
    </ChartFrame>
  )
}
