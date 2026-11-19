import type { ReactNode } from "react"
import { useElementWidth } from "@/hooks/use-element-width"

/** Gives a chart its container's width; charts draw at 1:1 so text never scales. */
export function ChartFrame({
  height,
  children,
}: {
  height: number
  children: (width: number) => ReactNode
}) {
  const [ref, width] = useElementWidth<HTMLDivElement>()
  return (
    <div ref={ref} style={{ height }}>
      {width > 0 && children(width)}
    </div>
  )
}

/** Shared SVG text style: 11px muted, like every axis label. */
export const AXIS_TEXT = "fill-muted-foreground text-[11px]"
