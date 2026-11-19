import { cn } from "@/lib/utils"

const STROKE = 1.5

/**
 * A rounded square outline marking a scheduled term. Its dash pattern repeats
 * a whole number of times per side and is centered on each side, so every
 * side and corner looks the same; a fixed pattern leaves a short dash
 * wherever the outline's length runs out.
 */
export function DashedSquare({
  size,
  radius,
  dashesPerSide,
  className,
}: {
  size: number
  radius: number
  dashesPerSide: number
  className?: string
}) {
  const inset = STROKE / 2
  const side = size - STROKE
  const r = radius - inset
  const straight = side - 2 * r
  const period = (straight + (Math.PI * r) / 2) / dashesPerSide
  const dash = period * 0.55
  return (
    <svg
      width={size}
      height={size}
      className={cn("block", className)}
      aria-hidden
    >
      <rect
        x={inset}
        y={inset}
        width={side}
        height={side}
        rx={r}
        fill="none"
        stroke="currentColor"
        strokeWidth={STROKE}
        strokeDasharray={`${dash} ${period - dash}`}
        strokeDashoffset={dash / 2 - straight / 2}
      />
    </svg>
  )
}
