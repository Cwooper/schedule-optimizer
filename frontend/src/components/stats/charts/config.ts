import type { ChartConfig } from "@/components/ui/chart"

/** Series colors for ChartContainer, from the theme. */
export const STATS_CHART_CONFIG = {
  fg: { color: "var(--foreground)" },
  soft: { color: "var(--chart-soft)" },
} satisfies ChartConfig

/** Recharts axis text: 11px, like the rest of the stats axes. */
export const AXIS_TICK = { fontSize: 11 }

/**
 * The datum a Recharts tooltip renders: the hovered one, else a stand-in.
 * Recharts places a tooltip, flipping it to stay inside the chart, from its
 * wrapper's last measured size, so the hidden idle wrapper holds a tooltip
 * of the real shape; an empty or smaller one would put the first tooltip in
 * the wrong spot for a frame.
 */
export function tipDatum<T>(
  active: boolean | undefined,
  payload: readonly { payload?: unknown }[] | undefined,
  standIn: T | undefined
): T | undefined {
  return (
    ((active ? payload?.[0]?.payload : undefined) as T | undefined) ?? standIn
  )
}
