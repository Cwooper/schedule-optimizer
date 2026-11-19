import type { AxisDomain } from "@/lib/stats"

const WIDTH = 110

/** A row's value on a short track, with a tick at the course's. */
export function GpaTrack({
  value,
  mean,
  domain,
}: {
  value: number
  mean: number | null
  domain: AxisDomain
}) {
  const x = (v: number) => {
    const t = Math.min(
      1,
      Math.max(0, (v - domain.lo) / (domain.hi - domain.lo))
    )
    return 5 + t * (WIDTH - 10)
  }
  return (
    <svg width={WIDTH} height={16} className="block" aria-hidden>
      <line
        x1={5}
        x2={WIDTH - 5}
        y1={8}
        y2={8}
        className="stroke-border"
        strokeWidth={4}
        strokeLinecap="round"
      />
      {mean != null && (
        <line
          x1={x(mean)}
          x2={x(mean)}
          y1={1}
          y2={15}
          className="stroke-muted-foreground"
          strokeWidth={1.5}
        />
      )}
      <circle cx={x(value)} cy={8} r={5} className="fill-foreground" />
    </svg>
  )
}
