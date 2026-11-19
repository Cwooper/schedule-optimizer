import type { ChartConfig } from "@/components/ui/chart"

/** Series colors for ChartContainer, from the theme. */
export const STATS_CHART_CONFIG = {
  fg: { color: "var(--foreground)" },
  soft: { color: "var(--chart-soft)" },
} satisfies ChartConfig

/** Recharts axis text: 11px, like the rest of the stats axes. */
export const AXIS_TICK = { fontSize: 11 }
