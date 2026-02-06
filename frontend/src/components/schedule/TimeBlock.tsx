import { X } from "lucide-react"
import { cn } from "@/lib/utils"
import { GRID, HATCH_GRADIENT, opacityToHex } from "@/lib/schedule-utils"

function blockedColorStyle(
  color: string,
  editing: boolean,
  hatched: boolean,
  opacity: number,
): React.CSSProperties {
  const bgHex = opacityToHex(editing ? Math.min(opacity + 10, 100) : opacity)
  const borderHex = opacityToHex(editing ? Math.min(opacity + 30, 100) : 100)
  return {
    backgroundColor: `${color}${bgHex}`,
    borderLeft: `3px solid ${color}${borderHex}`,
    ...(editing ? { borderColor: `${color}${borderHex}` } : {}),
    ...(hatched ? { backgroundImage: HATCH_GRADIENT } : {}),
  }
}

export interface TimeBlockProps {
  dayIndex: number
  startMin: number
  endMin: number
  gridStartMin: number
  gridHeight: number

  /** Tailwind class string (courses) OR hex string (blocked) OR null (no color) */
  color: string | null
  variant: "course" | "blocked"
  hatched?: boolean
  opacity?: number

  interactive: boolean
  dimmed: boolean
  editing?: boolean

  onDelete?: (e: React.UIEvent) => void
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
  hatched,
  opacity,
  interactive,
  dimmed,
  editing,
  onDelete,
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
  const isHatchedOnly = variant === "blocked" && hatched && color === null && !editing

  const className = cn(
    "absolute z-20 rounded overflow-hidden text-left flex flex-col justify-start transition-all px-1 sm:px-1.5 pt-0.5",
    (isCourse || isBlockedWithColor) && "border-l-3",
    isCourse && color,
    dimmed && "opacity-40 pointer-events-none",
    interactive && "cursor-pointer hover:ring-2 hover:ring-primary/50 hover:brightness-95",
    editing && "border border-red-500/50 cursor-pointer hover:brightness-125",
    editing && !color && "bg-red-500/25 dark:bg-red-500/30",
    !interactive && !editing && !dimmed && "pointer-events-none",
  )

  const style: React.CSSProperties = {
    top: `${top}%`,
    height: `${height}%`,
    left: leftOffset,
    width: blockWidth,
    ...(isBlockedWithColor
      ? blockedColorStyle(color, !!editing, !!hatched, opacity ?? 20)
      : {}),
    ...(isHatchedOnly ? { backgroundImage: HATCH_GRADIENT } : {}),
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
      {editing && onDelete && (
        <span
          role="button"
          tabIndex={0}
          className="absolute top-0 right-0 z-30 flex size-5 items-center justify-center rounded-bl bg-red-500/80 text-white hover:bg-red-600 transition-colors"
          onPointerDown={(e) => e.stopPropagation()}
          onClick={(e) => {
            e.stopPropagation()
            onDelete(e)
          }}
          onKeyDown={(e) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault()
              e.stopPropagation()
              onDelete(e)
            }
          }}
          aria-label="Delete block"
        >
          <X className="size-3" />
        </span>
      )}
      {children}
    </button>
  )
}
