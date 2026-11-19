import {
  Bar,
  BarChart,
  type BarShapeProps,
  CartesianGrid,
  Cell,
  XAxis,
  YAxis,
} from "recharts"
import { ChartContainer, ChartTooltip } from "@/components/ui/chart"
import type { GradeCounts } from "@/lib/api"
import {
  GRADE_LABELS,
  WITHDREW_COLUMN,
  axisTicks,
  formatPct,
  gradeShares,
} from "@/lib/stats"
import { ChartTipBox, ChartTipPlaceholder, TipBody, type TipRow } from "./ChartTip"
import { AXIS_TICK, STATS_CHART_CONFIG } from "./config"

function gridStep(max: number): number {
  if (max <= 0.3) return 0.05
  if (max <= 0.6) return 0.1
  return 0.2
}

export interface DistributionSeries {
  label: string
  grades: GradeCounts
}

interface Datum {
  column: number
  grade: string
  fg?: number
  bg?: number
}

/**
 * Share of grades per column. A background series (the course, behind a
 * section) draws full width and soft, with the foreground narrower on top so
 * both read at once.
 */
export function DistributionChart({
  series,
  background,
  columns,
}: {
  series?: DistributionSeries
  background?: DistributionSeries
  columns: number[]
}) {
  const fg = series && gradeShares(series.grades, columns)
  const bg = background && gradeShares(background.grades, columns)
  const data: Datum[] = columns.map((c, i) => ({
    column: c,
    grade: GRADE_LABELS[c],
    fg: fg?.[i],
    bg: bg?.[i],
  }))
  const top = Math.max(...(fg ?? []), ...(bg ?? []), 0.01) * 1.05
  const step = gridStep(top)
  const max = Math.ceil(top / step) * step

  const tipRow = (s: DistributionSeries, column: number, share?: number) =>
    [
      s.label,
      `${s.grades[column].toLocaleString()} ${s.grades[column] === 1 ? "student" : "students"}`,
      formatPct(share ?? 0, 1),
    ] as TipRow

  return (
    <ChartContainer
      config={STATS_CHART_CONFIG}
      className="aspect-auto h-[200px] w-full"
    >
      <BarChart
        data={data}
        margin={{ top: 8, right: 0, bottom: 0, left: 0 }}
        barCategoryGap="14%"
      >
        <CartesianGrid vertical={false} />
        <XAxis
          xAxisId="fg"
          dataKey="grade"
          tickLine={false}
          axisLine={false}
          tick={AXIS_TICK}
          interval={0}
          height={22}
        />
        <XAxis xAxisId="bg" dataKey="grade" orientation="top" hide />
        <YAxis
          domain={[0, max]}
          ticks={axisTicks({ lo: 0, hi: max, step })}
          tickFormatter={(v: number) => `${Math.round(v * 100)}%`}
          tickLine={false}
          axisLine={false}
          tick={AXIS_TICK}
          width={34}
        />
        <ChartTooltip
          cursor={{
            fill: "var(--foreground)",
            fillOpacity: 0.05,
            stroke: "none",
          }}
          isAnimationActive={false}
          content={({ active, payload }) => {
            const d = payload?.[0]?.payload as Datum | undefined
            if (!active || !d) return <ChartTipPlaceholder />
            const rows: TipRow[] = []
            if (series) rows.push(tipRow(series, d.column, d.fg))
            if (background) rows.push(tipRow(background, d.column, d.bg))
            return (
              <ChartTipBox>
                <TipBody title={d.grade} rows={rows} />
              </ChartTipBox>
            )
          }}
        />
        {bg && (
          <Bar
            xAxisId="bg"
            dataKey="bg"
            fill="var(--color-soft)"
            radius={3}
            maxBarSize={36}
            isAnimationActive={false}
          />
        )}
        {fg && (
          <Bar
            xAxisId="fg"
            dataKey="fg"
            fill="var(--color-fg)"
            radius={3}
            maxBarSize={bg ? 36 : 28}
            isAnimationActive={false}
            shape={bg ? InsetBar : undefined}
          >
            {data.map((d) => (
              <Cell
                key={d.column}
                fillOpacity={d.column === WITHDREW_COLUMN ? 0.45 : 1}
              />
            ))}
          </Bar>
        )}
      </BarChart>
    </ChartContainer>
  )
}

/** The foreground bar over a background one: narrower, so both show. */
function InsetBar(props: BarShapeProps) {
  const { x = 0, y = 0, width = 0, height = 0, fill, fillOpacity } = props
  const w = Math.max(width - 8, width * 0.55)
  if (height <= 0) return <g />
  return (
    <rect
      x={x + (width - w) / 2}
      y={y}
      width={w}
      height={height}
      rx={3}
      fill={fill}
      fillOpacity={fillOpacity}
    />
  )
}
