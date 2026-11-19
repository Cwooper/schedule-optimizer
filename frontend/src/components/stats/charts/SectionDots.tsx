import {
  Scatter,
  ScatterChart,
  type ScatterShapeProps,
  XAxis,
  YAxis,
} from "recharts"
import { ChartContainer, ChartTooltip } from "@/components/ui/chart"
import { useElementWidth } from "@/hooks/use-element-width"
import type { StatsDot, StatsMetric } from "@/lib/api"
import {
  axisTicks,
  type DotPoint,
  dotLayout,
  formatMetric,
  termName,
} from "@/lib/stats"
import { cn } from "@/lib/utils"
import { ChartTipBox, ChartTipPlaceholder, TipBody } from "./ChartTip"
import { AXIS_TICK, STATS_CHART_CONFIG } from "./config"

// Fixed geometry, so the plot's pixel size is known before Recharts lays it
// out: dotLayout sizes the dots from it.
const HEIGHT = 200
const MARGIN = { top: 8, right: 8, bottom: 0, left: 0 }
const Y_AXIS_WIDTH = 44
const X_AXIS_HEIGHT = 40

/**
 * Every reported section of a course as a stacked dot plot by value; the
 * current section is ringed and its professor's are dark. Clicking a dot
 * opens that section.
 */
export function SectionDots({
  dots,
  current,
  metric,
  onSelect,
}: {
  dots: StatsDot[]
  current: { term: string; crn: string }
  metric: StatsMetric
  onSelect: (dot: StatsDot) => void
}) {
  const [ref, width] = useElementWidth<HTMLDivElement>()
  const plot = {
    width: Math.max(1, width - MARGIN.left - MARGIN.right - Y_AXIS_WIDTH),
    height: HEIGHT - MARGIN.top - MARGIN.bottom - X_AXIS_HEIGHT,
  }
  const layout = dotLayout(dots, metric, current, plot)
  const label = metric === "gpa" ? "GPA" : "Pass"

  return (
    <div ref={ref}>
      <ChartContainer
        config={STATS_CHART_CONFIG}
        className="aspect-auto w-full"
        style={{ height: HEIGHT }}
      >
        <ScatterChart margin={MARGIN}>
          <XAxis
            dataKey="x"
            type="number"
            domain={layout.xDomain}
            ticks={axisTicks(layout.domain)}
            tickFormatter={(v: number) =>
              metric === "gpa" ? v.toFixed(1) : `${Math.round(v * 100)}%`
            }
            tickLine={false}
            axisLine={{ stroke: "var(--border)" }}
            tick={AXIS_TICK}
            height={X_AXIS_HEIGHT}
            label={{
              value: metric === "gpa" ? "Section GPA" : "Section pass rate",
              position: "insideBottom",
              offset: 3,
              fontSize: 11,
              fill: "var(--muted-foreground)",
            }}
          />
          <YAxis
            dataKey="y"
            type="number"
            domain={[0, layout.yMax]}
            ticks={layout.yTicks}
            tickLine={false}
            axisLine={false}
            tick={AXIS_TICK}
            width={Y_AXIS_WIDTH}
            label={{
              value: "Sections",
              angle: -90,
              position: "insideLeft",
              offset: 8,
              fontSize: 11,
              fill: "var(--muted-foreground)",
              style: { textAnchor: "middle" },
            }}
          />
          <ChartTooltip
            cursor={false}
            isAnimationActive={false}
            content={({ active, payload }) => {
              const d = payload?.[0]?.payload as DotPoint | undefined
              if (!active || !d) return <ChartTipPlaceholder />
              return (
                <ChartTipBox>
                  <TipBody
                    title={termName(d.term)}
                    rows={[
                      ["CRN", d.crn],
                      ["Professor", d.professor || "Unlisted"],
                      ["Students", d.students],
                      [label, formatMetric(metric, d.value)],
                    ]}
                  />
                </ChartTipBox>
              )
            }}
          />
          <Scatter
            data={layout.points}
            isAnimationActive={false}
            shape={(props: ScatterShapeProps) => (
              <DotShape {...props} r={layout.r} onSelect={onSelect} />
            )}
          />
        </ScatterChart>
      </ChartContainer>
    </div>
  )
}

function DotShape({
  cx = 0,
  cy = 0,
  payload,
  r,
  onSelect,
}: ScatterShapeProps & { r: number; onSelect: (dot: StatsDot) => void }) {
  const d = payload as DotPoint
  if (d.kind === "current") {
    const ring = Math.max(r + 1.5, 4.5)
    return (
      <circle
        cx={cx}
        cy={cy}
        r={ring}
        className="fill-card stroke-foreground"
        strokeWidth={Math.min(3, ring * 0.45)}
      />
    )
  }
  return (
    <circle
      cx={cx}
      cy={cy}
      r={r}
      onClick={() => onSelect(d)}
      className={cn(
        "hover:stroke-foreground cursor-pointer stroke-[1.5px]",
        d.kind === "mine" ? "fill-foreground" : "fill-chart-soft"
      )}
      fillOpacity={d.kind === "mine" ? 0.75 : 1}
    />
  )
}
