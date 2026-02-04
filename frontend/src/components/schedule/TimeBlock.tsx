import { cn } from "@/lib/utils"
import { GRID } from "@/lib/schedule-utils"

const HATCH_CLASS =
  "bg-[repeating-linear-gradient(45deg,transparent,transparent_4px,rgba(0,0,0,0.06)_4px,rgba(0,0,0,0.06)_8px)] dark:bg-[repeating-linear-gradient(45deg,transparent,transparent_4px,rgba(255,255,255,0.06)_4px,rgba(255,255,255,0.06)_8px)]"

export interface TimeBlockProps {
  dayIndex: number
  startMin: number
  endMin: number
  gridStartMin: number
  gridHeight: number

  /** Tailwind class string (courses) OR hex string (blocked) OR null (hatched) */
  color: string | null
  variant: "course" | "blocked"

  interactive: boolean
  dimmed: boolean
  editing?: boolean

  onClick?: (e: React.MouseEvent) => void
  tabIndex?: number
  ariaLabel?: string
  children?: React.ReactNode
}

export function TimeBlock({
  dayIndex,
  startMin,
  endMin,
  gridStartMin,
  gridHeight,
  color,
  variant,
  interactive,
  dimmed,
  editing,
  onClick,
  tabIndex,
  ariaLabel,
  children,
}: TimeBlockProps) {
  const top = ((startMin - gridStartMin) / gridHeight) * 100
  const height = ((endMin - startMin) / gridHeight) * 100
  const dayWidth = `(100% - ${GRID.TIME_COL}) / ${GRID.DAY_COUNT}`
  const leftOffset = `calc(${GRID.TIME_COL} + (${dayWidth}) * ${dayIndex} + ${GRID.BLOCK_INSET_PX}px)`
  const blockWidth = `calc(${dayWidth} - ${GRID.BLOCK_INSET_PX * 2}px)`

  const isCourse = variant === "course"
  const isBlockedWithColor = variant === "blocked" && color !== null
  const isHatched = variant === "blocked" && color === null && !editing

  const className = cn(
    "absolute z-20 rounded overflow-hidden text-left flex flex-col justify-start transition-all px-1 sm:px-1.5 pt-0.5",
    (isCourse || isBlockedWithColor) && "border-l-3",
    isCourse && color,
    dimmed && "opacity-40 pointer-events-none",
    interactive && "cursor-pointer hover:ring-2 hover:ring-primary/50 hover:brightness-95",
    editing && "border border-red-500/50 cursor-pointer hover:brightness-125",
    editing && !color && "bg-red-500/25 dark:bg-red-500/30",
    isHatched && HATCH_CLASS,
    !interactive && !editing && !dimmed && "pointer-events-none",
  )

  const style: React.CSSProperties = {
    top: `${top}%`,
    height: `${height}%`,
    left: leftOffset,
    width: blockWidth,
    ...(isBlockedWithColor && !editing
      ? { backgroundColor: `${color}30`, borderLeft: `3px solid ${color}` }
      : {}),
    ...(isBlockedWithColor && editing
      ? { backgroundColor: `${color}40`, borderColor: `${color}80` }
      : {}),
  }

  return (
    <button
      type="button"
      className={className}
      style={style}
      onClick={onClick}
      tabIndex={tabIndex}
      aria-label={ariaLabel}
    >
      {children}
    </button>
  )
}
