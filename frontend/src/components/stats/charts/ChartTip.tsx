import type { ReactNode } from "react"
import { cn } from "@/lib/utils"

export type TipRow = [ReactNode, ...ReactNode[]]

/**
 * Hover detail for every stats chart and glyph: a title over rows whose
 * first cell is a label and the rest right-aligned numbers.
 */
export function TipBody({ title, rows }: { title: ReactNode; rows: TipRow[] }) {
  const cols = Math.max(1, ...rows.map((r) => r.length))
  return (
    <div className="tabular-nums">
      <div className="font-semibold">{title}</div>
      {rows.length > 0 && (
        <div
          className="mt-0.5 grid gap-x-3"
          style={{ gridTemplateColumns: `repeat(${cols}, auto)` }}
        >
          {rows.map((row, i) =>
            row.map((cell, j) => (
              <span
                key={`${i}-${j}`}
                className={cn(j === 0 ? "opacity-70" : "text-right")}
              >
                {cell}
              </span>
            ))
          )}
        </div>
      )}
    </div>
  )
}

/** Recharts tooltip box, styled like the app's Radix tooltips. */
export function ChartTipBox({ children }: { children: ReactNode }) {
  return (
    <div className="bg-foreground text-background rounded-md px-3 py-1.5 text-xs shadow-md">
      {children}
    </div>
  )
}

/**
 * What a Recharts tooltip renders while inactive. Recharts positions the
 * tooltip from its wrapper's measured size, so an empty wrapper would put
 * the first tooltip at the chart's top left for a frame. The wrapper is
 * hidden while inactive, so this stand-in is never seen.
 */
export function ChartTipPlaceholder() {
  return (
    <ChartTipBox>
      <TipBody title="-" rows={[["-", "-"]]} />
    </ChartTipBox>
  )
}
