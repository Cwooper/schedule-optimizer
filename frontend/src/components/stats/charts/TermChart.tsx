import {
  CartesianGrid,
  Line,
  LineChart,
  ReferenceLine,
  XAxis,
  YAxis,
} from "recharts"
import { ChartContainer, ChartTooltip } from "@/components/ui/chart"
import type { StatsMetric, StatsTerm } from "@/lib/api"
import { axisDomain, axisTicks, formatMetric, termName } from "@/lib/stats"
import { ChartTipBox, TipBody } from "./ChartTip"
import { AXIS_TICK, STATS_CHART_CONFIG } from "./config"

/** Terms consecutively, so x spacing is true time and gaps show. */
function ordinal(term: string): number {
  return Number(term.slice(0, 4)) * 4 + Number(term.slice(4)) / 10 - 1
}

interface Datum extends StatsTerm {
  x: number
  value: number
}

/** The metric per term, with the course's overall value dashed across. */
export function TermChart({
  terms,
  mean,
  metric,
}: {
  terms: StatsTerm[]
  mean: number | null
  metric: StatsMetric
}) {
  const data: Datum[] = terms
    .filter((t): t is StatsTerm & { value: number } => t.value != null)
    .map((t) => ({ ...t, x: ordinal(t.term) }))
  const domain = axisDomain(metric, [
    ...data.map((d) => d.value),
    ...(mean == null ? [] : [mean]),
  ])
  const first = data[0]?.x ?? 0
  const last = data.at(-1)?.x ?? 0
  // A tick per calendar year at its Winter term; the axis thins them.
  const years: number[] = []
  for (let yr = Math.ceil(first / 4); yr * 4 <= last; yr++) years.push(yr * 4)
  const label = metric === "gpa" ? "GPA" : "Pass"

  return (
    <ChartContainer
      config={STATS_CHART_CONFIG}
      className="aspect-auto h-[200px] w-full"
    >
      <LineChart data={data} margin={{ top: 8, right: 14, bottom: 0, left: 0 }}>
        <CartesianGrid vertical={false} />
        <XAxis
          dataKey="x"
          type="number"
          domain={[first, last]}
          ticks={years}
          tickFormatter={(x: number) => String(x / 4)}
          interval="equidistantPreserveStart"
          minTickGap={16}
          tickLine={false}
          axisLine={false}
          tick={AXIS_TICK}
          height={22}
          padding={{ left: 6 }}
        />
        <YAxis
          domain={[domain.lo, domain.hi]}
          ticks={axisTicks(domain)}
          tickFormatter={(v: number) =>
            metric === "gpa" ? v.toFixed(1) : `${Math.round(v * 100)}%`
          }
          tickLine={false}
          axisLine={false}
          tick={AXIS_TICK}
          // "100%" needs more room than "4.0".
          width={metric === "gpa" ? 30 : 36}
        />
        {mean != null && (
          <ReferenceLine
            y={mean}
            stroke="var(--muted-foreground)"
            strokeDasharray="3 4"
          />
        )}
        <ChartTooltip
          cursor={{ stroke: "var(--border)" }}
          isAnimationActive={false}
          content={({ active, payload }) => {
            const d = payload?.[0]?.payload as Datum | undefined
            if (!active || !d) return null
            return (
              <ChartTipBox>
                <TipBody
                  title={termName(d.term)}
                  rows={[
                    [label, formatMetric(metric, d.value)],
                    ["Sections", d.sections],
                    ["Students", d.students.toLocaleString()],
                  ]}
                />
              </ChartTipBox>
            )
          }}
        />
        <Line
          dataKey="value"
          stroke="var(--color-fg)"
          strokeWidth={2}
          strokeLinejoin="round"
          dot={{
            r: 3,
            fill: "var(--color-fg)",
            stroke: "var(--card)",
            strokeWidth: 1.5,
          }}
          activeDot={{
            r: 5,
            fill: "var(--color-fg)",
            stroke: "var(--card)",
            strokeWidth: 2,
          }}
          isAnimationActive={false}
        />
      </LineChart>
    </ChartContainer>
  )
}
