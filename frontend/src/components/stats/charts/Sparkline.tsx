import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import type { GradeCounts } from "@/lib/api"
import {
  GRADE_LABELS,
  WITHDREW_COLUMN,
  formatPct,
  gradeShares,
} from "@/lib/stats"

const WIDTH = 104
const HEIGHT = 26

/** A row's grade distribution in miniature, over its course's columns. */
export function Sparkline({
  grades,
  columns,
}: {
  grades: GradeCounts
  columns: number[]
}) {
  const bw = WIDTH / columns.length
  const max = Math.max(1, ...columns.map((c) => grades[c]))
  const shares = gradeShares(grades, columns)
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span className="block w-fit">
          <svg width={WIDTH} height={HEIGHT} className="block" aria-hidden>
            {columns.map((c, i) => {
              const h = Math.max(1, (grades[c] / max) * HEIGHT)
              return (
                <rect
                  key={c}
                  x={i * bw + 0.5}
                  y={HEIGHT - h}
                  width={bw - 1.5}
                  height={h}
                  rx={1}
                  className="fill-foreground"
                  fillOpacity={c === WITHDREW_COLUMN ? 0.4 : 1}
                />
              )
            })}
          </svg>
        </span>
      </TooltipTrigger>
      <TooltipContent>
        <div className="grid grid-cols-[repeat(4,auto_auto)] gap-x-2 tabular-nums">
          {columns.map((c, i) => [
            <span key={`${c}l`} className="opacity-70">
              {GRADE_LABELS[c]}
            </span>,
            <span key={`${c}v`} className="pr-1 text-right">
              {formatPct(shares[i])}
            </span>,
          ])}
        </div>
      </TooltipContent>
    </Tooltip>
  )
}
