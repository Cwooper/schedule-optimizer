import { useMemo, useCallback, useRef, useState, useEffect } from "react"
import type { HydratedSection, MeetingTime } from "@/lib/api"
import { cn, decodeHtmlEntities } from "@/lib/utils"
import type { BlockedTime } from "@/stores/app-store"

export interface ScheduleGridProps {
  courses: HydratedSection[]
  onCourseClick?: (crn: string) => void
  blockedTimes?: BlockedTime[]
  editingBlockedTimes?: boolean
  onAddBlockedTime?: (bt: BlockedTime) => void
  onRemoveBlockedTime?: (index: number) => void
}

// Grid constants
const DAYS = ["Mon", "Tue", "Wed", "Thu", "Fri"]
const TIME_COL = "2.5rem"
const TIME_COL_PX = 40
const DEFAULT_START_MIN = 8 * 60  // 8:00am
const DEFAULT_END_MIN = 16 * 60   // 4:00pm
const PADDING_MIN = 10
const MIN_HOURS = 8

// Curated 8-color palette — more saturated in light mode for visibility
const COURSE_COLORS = [
  "bg-blue-500/40 dark:bg-blue-500/35 border-blue-500 text-blue-950 dark:text-blue-100",
  "bg-emerald-500/40 dark:bg-emerald-500/35 border-emerald-500 text-emerald-950 dark:text-emerald-100",
  "bg-rose-500/40 dark:bg-rose-500/35 border-rose-500 text-rose-950 dark:text-rose-100",
  "bg-orange-500/40 dark:bg-orange-500/35 border-orange-500 text-orange-950 dark:text-orange-100",
  "bg-violet-500/40 dark:bg-violet-500/35 border-violet-500 text-violet-950 dark:text-violet-100",
  "bg-teal-500/40 dark:bg-teal-500/35 border-teal-500 text-teal-950 dark:text-teal-100",
  "bg-amber-500/40 dark:bg-amber-500/35 border-amber-500 text-amber-950 dark:text-amber-100",
  "bg-fuchsia-500/40 dark:bg-fuchsia-500/35 border-fuchsia-500 text-fuchsia-950 dark:text-fuchsia-100",
]

function hashString(str: string): number {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i)
    hash = (hash << 5) - hash + char
    hash = hash & hash
  }
  return Math.abs(hash)
}

function parseTime(time: string): number {
  if (!time) return 0
  const clean = time.replace(":", "").padStart(4, "0")
  const hours = parseInt(clean.slice(0, 2), 10) || 0
  const mins = parseInt(clean.slice(2, 4), 10) || 0
  return hours * 60 + mins
}

function formatHour(hour: number): string {
  if (hour === 0 || hour === 12) return "12"
  return hour > 12 ? `${hour - 12}` : `${hour}`
}

function formatAmPm(hour: number): string {
  return hour >= 12 ? "pm" : "am"
}

interface CourseBlock {
  course: HydratedSection
  meeting: MeetingTime
  dayIndex: number
  startMin: number
  endMin: number
  colorClass: string
}

function buildColorMap(courses: HydratedSection[]): Map<string, string> {
  const colorMap = new Map<string, string>()
  for (const course of courses) {
    const key = `${course.subject}:${course.courseNumber}`
    if (!colorMap.has(key)) {
      const colorIndex = hashString(key) % COURSE_COLORS.length
      colorMap.set(key, COURSE_COLORS[colorIndex])
    }
  }
  return colorMap
}

function meetingToBlocks(
  course: HydratedSection,
  meeting: MeetingTime,
  colorClass: string
): CourseBlock[] {
  const startMin = parseTime(meeting.startTime)
  const endMin = parseTime(meeting.endTime)
  const blocks: CourseBlock[] = []

  for (let dayIdx = 1; dayIdx <= 5; dayIdx++) {
    if (meeting.days[dayIdx]) {
      blocks.push({
        course,
        meeting,
        dayIndex: dayIdx - 1,
        startMin,
        endMin,
        colorClass,
      })
    }
  }
  return blocks
}

/** Floor minutes down to the nearest 30-min boundary */
function floorTo30(min: number): number {
  return Math.floor(min / 30) * 30
}

/** Ceil minutes up to the nearest 30-min boundary */
function ceilTo30(min: number): number {
  return Math.ceil(min / 30) * 30
}

/** Compute the visible time range based on course blocks and blocked times.
 *  - 10 min padding then floor/ceil to nearest 30-min mark
 *  - Default range: 8:00am–4:00pm
 *  - Minimum span of MIN_HOURS */
function computeTimeRange(
  blocks: CourseBlock[],
  blockedTimes?: BlockedTime[]
): { startMin: number; endMin: number } {
  let earliestMin = Infinity
  let latestMin = -Infinity

  for (const block of blocks) {
    earliestMin = Math.min(earliestMin, block.startMin)
    latestMin = Math.max(latestMin, block.endMin)
  }

  if (blockedTimes) {
    for (const bt of blockedTimes) {
      earliestMin = Math.min(earliestMin, parseTime(bt.startTime))
      latestMin = Math.max(latestMin, parseTime(bt.endTime))
    }
  }

  // If no data, use clean defaults with no padding
  if (!isFinite(earliestMin)) {
    return { startMin: DEFAULT_START_MIN, endMin: DEFAULT_END_MIN }
  }

  // Add 10-min padding, then snap to 30-min boundaries
  let startMin = floorTo30(earliestMin - PADDING_MIN)
  let endMin = ceilTo30(latestMin + PADDING_MIN)

  // Enforce defaults as bounds
  startMin = Math.max(0, Math.min(startMin, DEFAULT_START_MIN))
  endMin = Math.min(24 * 60, Math.max(endMin, DEFAULT_END_MIN))

  // Enforce minimum range
  if (endMin - startMin < MIN_HOURS * 60) {
    endMin = Math.min(24 * 60, startMin + MIN_HOURS * 60)
  }

  return { startMin, endMin }
}

export function ScheduleGrid({
  courses,
  onCourseClick,
  blockedTimes,
  editingBlockedTimes,
  onAddBlockedTime,
  onRemoveBlockedTime,
}: ScheduleGridProps) {
  const blocks = useMemo(() => {
    const colorMap = buildColorMap(courses)
    return courses.flatMap((course) => {
      const colorClass = colorMap.get(`${course.subject}:${course.courseNumber}`)!
      return course.meetingTimes.flatMap((meeting) =>
        meetingToBlocks(course, meeting, colorClass)
      )
    })
  }, [courses])

  const { startMin: gridStartMin, endMin: gridEndMin } = useMemo(
    () => computeTimeRange(blocks, blockedTimes),
    [blocks, blockedTimes]
  )

  const gridHeight = gridEndMin - gridStartMin

  // Hour labels: whole hours that fall within the grid range
  const hours = useMemo(() => {
    const firstHour = Math.ceil(gridStartMin / 60)
    const lastHour = Math.floor(gridEndMin / 60)
    return Array.from({ length: lastHour - firstHour + 1 }, (_, i) => firstHour + i)
  }, [gridStartMin, gridEndMin])

  // Drag state for blocked time painting
  const gridBodyRef = useRef<HTMLDivElement>(null)
  const [dragState, setDragState] = useState<{
    dayIndex: number
    startMin: number
    currentMin: number
  } | null>(null)
  const dragRef = useRef(dragState)
  useEffect(() => {
    dragRef.current = dragState
  }, [dragState])

  const pointerToGridPos = useCallback(
    (clientX: number, clientY: number) => {
      const el = gridBodyRef.current
      if (!el) return null
      const rect = el.getBoundingClientRect()
      const timeColWidth = TIME_COL_PX
      const x = clientX - rect.left - timeColWidth
      const y = clientY - rect.top + el.scrollTop
      const dayColWidth = (rect.width - timeColWidth) / 5
      const dayIndex = Math.floor(x / dayColWidth)
      if (dayIndex < 0 || dayIndex > 4) return null
      const totalHeight = el.scrollHeight
      const minuteOffset = (y / totalHeight) * gridHeight
      const minute = gridStartMin + minuteOffset
      // Snap to 15-minute increments
      const snapped = Math.round(minute / 15) * 15
      return { dayIndex, minute: Math.max(gridStartMin, Math.min(gridEndMin, snapped)) }
    },
    [gridStartMin, gridEndMin, gridHeight]
  )

  const handlePointerDown = useCallback(
    (e: React.PointerEvent) => {
      if (!editingBlockedTimes) return
      const pos = pointerToGridPos(e.clientX, e.clientY)
      if (!pos) return
      const target = e.target as HTMLElement
      target.setPointerCapture(e.pointerId)
      setDragState({ dayIndex: pos.dayIndex, startMin: pos.minute, currentMin: pos.minute })
    },
    [editingBlockedTimes, pointerToGridPos]
  )

  const handlePointerMove = useCallback(
    (e: React.PointerEvent) => {
      const drag = dragRef.current
      if (!drag) return
      const pos = pointerToGridPos(e.clientX, e.clientY)
      if (!pos || pos.dayIndex !== drag.dayIndex) return
      setDragState((prev) => prev && { ...prev, currentMin: pos.minute })
    },
    [pointerToGridPos]
  )

  const handlePointerUp = useCallback(() => {
    const drag = dragRef.current
    if (!drag || !onAddBlockedTime) return
    const minMin = Math.min(drag.startMin, drag.currentMin)
    const maxMin = Math.max(drag.startMin, drag.currentMin)
    if (maxMin - minMin >= 15) {
      const startH = String(Math.floor(minMin / 60)).padStart(2, "0")
      const startM = String(minMin % 60).padStart(2, "0")
      const endH = String(Math.floor(maxMin / 60)).padStart(2, "0")
      const endM = String(maxMin % 60).padStart(2, "0")
      onAddBlockedTime({
        day: drag.dayIndex,
        startTime: `${startH}${startM}`,
        endTime: `${endH}${endM}`,
      })
    }
    setDragState(null)
  }, [onAddBlockedTime])

  // Cleanup drag on escape
  useEffect(() => {
    if (!dragState) return
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") setDragState(null)
    }
    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [dragState])

  return (
    <div className="flex h-full flex-col overflow-hidden">
      {/* Header row with day names */}
      <div className="border-b bg-muted/30">
        <div className="grid grid-cols-[2.5rem_repeat(5,1fr)]">
          <div className="p-1" />
          {DAYS.map((day, idx) => (
            <div
              key={day}
              className={cn(
                "border-l p-2 text-center text-sm font-medium",
                idx % 2 === 1 && "bg-muted/20"
              )}
            >
              {day}
            </div>
          ))}
        </div>
      </div>

      {/* Scrollable grid body */}
      <div
        ref={gridBodyRef}
        className={cn(
          "relative flex-1 overflow-auto",
          editingBlockedTimes && "cursor-crosshair"
        )}
        onPointerDown={handlePointerDown}
        onPointerMove={handlePointerMove}
        onPointerUp={handlePointerUp}
      >
        <div
          className="relative grid grid-cols-[2.5rem_repeat(5,1fr)]"
          style={{ minHeight: `${(gridHeight / 60) * 3}rem` }}
        >
          {/* Time labels column */}
          <div className="relative">
            {hours.map((hour) => (
              <div
                key={hour}
                className="absolute right-1.5 -translate-y-1/2 text-[11px] leading-none text-muted-foreground tabular-nums"
                style={{ top: `${((hour * 60 - gridStartMin) / gridHeight) * 100}%` }}
              >
                <span className="font-medium">{formatHour(hour)}</span>
                <span className="text-muted-foreground/70">{formatAmPm(hour)}</span>
              </div>
            ))}
          </div>

          {/* Day columns with hour lines */}
          {DAYS.map((_, dayIdx) => (
            <div key={dayIdx} className={cn("relative border-l border-b border-border/60", dayIdx % 2 === 1 && "bg-muted/20")}>
              {/* Hour lines */}
              {hours.map((hour) => (
                <div
                  key={hour}
                  className="absolute left-0 right-0 border-t border-border/60"
                  style={{ top: `${((hour * 60 - gridStartMin) / gridHeight) * 100}%` }}
                />
              ))}
              {/* Half-hour lines */}
              {hours.map((hour) => {
                const halfMin = hour * 60 + 30
                if (halfMin >= gridEndMin) return null
                return (
                  <div
                    key={`${hour}-half`}
                    className="absolute left-0 right-0 border-t border-dashed border-border/30"
                    style={{ top: `${((halfMin - gridStartMin) / gridHeight) * 100}%` }}
                  />
                )
              })}
            </div>
          ))}

          {/* Blocked time regions */}
          {blockedTimes?.map((bt, idx) => {
            const btStartMin = parseTime(bt.startTime)
            const btEndMin = parseTime(bt.endTime)
            const top = ((btStartMin - gridStartMin) / gridHeight) * 100
            const height = ((btEndMin - btStartMin) / gridHeight) * 100
            const dayWidth = `(100% - ${TIME_COL}) / 5`
            const leftOffset = `calc(${TIME_COL} + (${dayWidth}) * ${bt.day})`
            const blockWidth = `calc(${dayWidth})`

            return (
              <button
                key={`blocked-${idx}`}
                type="button"
                className={cn(
                  "absolute z-10 rounded-sm",
                  editingBlockedTimes
                    ? "bg-red-500/25 dark:bg-red-500/30 border border-red-500/50 cursor-pointer hover:bg-red-500/40"
                    : "pointer-events-none bg-[repeating-linear-gradient(45deg,transparent,transparent_4px,rgba(0,0,0,0.06)_4px,rgba(0,0,0,0.06)_8px)] dark:bg-[repeating-linear-gradient(45deg,transparent,transparent_4px,rgba(255,255,255,0.06)_4px,rgba(255,255,255,0.06)_8px)]"
                )}
                style={{
                  top: `${top}%`,
                  height: `${height}%`,
                  left: leftOffset,
                  width: blockWidth,
                }}
                onClick={
                  editingBlockedTimes
                    ? (e) => {
                        e.stopPropagation()
                        onRemoveBlockedTime?.(idx)
                      }
                    : undefined
                }
                tabIndex={editingBlockedTimes ? 0 : -1}
                aria-label={
                  editingBlockedTimes
                    ? `Remove blocked time on ${DAYS[bt.day]}`
                    : undefined
                }
              />
            )
          })}

          {/* Drag preview for blocked time painting */}
          {dragState && (() => {
            const minMin = Math.min(dragState.startMin, dragState.currentMin)
            const maxMin = Math.max(dragState.startMin, dragState.currentMin)
            const top = ((minMin - gridStartMin) / gridHeight) * 100
            const height = ((maxMin - minMin) / gridHeight) * 100
            const dayWidth = `(100% - ${TIME_COL}) / 5`
            const leftOffset = `calc(${TIME_COL} + (${dayWidth}) * ${dragState.dayIndex})`
            const blockWidth = `calc(${dayWidth})`

            return (
              <div
                className="pointer-events-none absolute z-20 rounded-sm border-2 border-red-500/60 bg-red-500/20"
                style={{
                  top: `${top}%`,
                  height: `${height}%`,
                  left: leftOffset,
                  width: blockWidth,
                }}
              />
            )
          })()}

          {/* Course blocks overlay */}
          {blocks.map((block, idx) => {
            const top = ((block.startMin - gridStartMin) / gridHeight) * 100
            const height = ((block.endMin - block.startMin) / gridHeight) * 100
            const durationMin = block.endMin - block.startMin
            const dayWidth = `(100% - ${TIME_COL}) / 5`
            const leftOffset = `calc(${TIME_COL} + (${dayWidth}) * ${block.dayIndex} + 2px)`
            const blockWidth = `calc(${dayWidth} - 4px)`

            return (
              <button
                key={idx}
                type="button"
                className={cn(
                  "absolute z-20 rounded border-l-3 px-1 sm:px-1.5 pt-0.5 text-xs overflow-hidden text-left transition-all flex flex-col justify-start",
                  block.colorClass,
                  editingBlockedTimes && "opacity-40 pointer-events-none",
                  !editingBlockedTimes && onCourseClick && "cursor-pointer hover:ring-2 hover:ring-primary/50 hover:brightness-95"
                )}
                style={{
                  top: `${top}%`,
                  height: `${height}%`,
                  left: leftOffset,
                  width: blockWidth,
                }}
                onClick={() => !editingBlockedTimes && onCourseClick?.(block.course.crn)}
                tabIndex={editingBlockedTimes ? -1 : 0}
              >
                <div className="font-semibold leading-tight text-[10px] sm:text-xs overflow-hidden whitespace-nowrap">
                  {block.course.subject} {block.course.courseNumber}
                </div>
                {durationMin >= 70 && (
                  <div className="truncate text-[10px] leading-tight opacity-80">
                    {decodeHtmlEntities(block.course.title)}
                  </div>
                )}
                {durationMin >= 80 && (
                  <div className="truncate text-[10px] leading-tight opacity-70">
                    {decodeHtmlEntities(block.course.instructor)}
                  </div>
                )}
                {durationMin >= 50 && (
                  <div className="truncate text-[10px] leading-tight opacity-60">
                    {block.meeting.building} {block.meeting.room}
                  </div>
                )}
              </button>
            )
          })}
        </div>

      </div>
    </div>
  )
}
