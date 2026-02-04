import { useMemo } from "react"
import type { HydratedSection, MeetingTime } from "@/lib/api"
import { cn, decodeHtmlEntities } from "@/lib/utils"
import type { BlockedTimeBlock, BlockedTimeGroup } from "@/stores/app-store"
import {
  DAYS,
  GRID,
  parseTime,
  formatHour,
  formatAmPm,
  computeTimeRange,
  buildColorMap,
} from "@/lib/schedule-utils"
import { useDragToPaint } from "@/hooks/use-drag-to-paint"
import { TimeBlock } from "./TimeBlock"

export interface ScheduleGridProps {
  courses: HydratedSection[]
  onCourseClick?: (crn: string) => void
  onBlockedTimeClick?: (groupId: string) => void
  blockedTimeGroups?: BlockedTimeGroup[]
  editingGroupId?: string | null
  onAddBlock?: (block: BlockedTimeBlock) => void
  onRemoveBlock?: (groupId: string, blockIndex: number) => void
}

interface CourseBlock {
  course: HydratedSection
  meeting: MeetingTime
  dayIndex: number
  startMin: number
  endMin: number
  colorClass: string
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

export function ScheduleGrid({
  courses,
  onCourseClick,
  onBlockedTimeClick,
  blockedTimeGroups,
  editingGroupId,
  onAddBlock,
  onRemoveBlock,
}: ScheduleGridProps) {
  const isEditing = editingGroupId != null

  const blocks = useMemo(() => {
    const colorMap = buildColorMap(courses)
    return courses.flatMap((course) => {
      const colorClass = colorMap.get(`${course.subject}:${course.courseNumber}`)!
      return course.meetingTimes.flatMap((meeting) =>
        meetingToBlocks(course, meeting, colorClass)
      )
    })
  }, [courses])

  // Build unified time spans for range computation
  const timeSpans = useMemo(() => {
    const spans: { startMin: number; endMin: number }[] = blocks.map((b) => ({
      startMin: b.startMin,
      endMin: b.endMin,
    }))
    if (blockedTimeGroups) {
      for (const group of blockedTimeGroups) {
        if (!group.enabled) continue
        for (const bt of group.blocks) {
          spans.push({
            startMin: parseTime(bt.startTime),
            endMin: parseTime(bt.endTime),
          })
        }
      }
    }
    return spans
  }, [blocks, blockedTimeGroups])

  const { startMin: gridStartMin, endMin: gridEndMin } = useMemo(
    () => computeTimeRange(timeSpans),
    [timeSpans]
  )

  const gridHeight = gridEndMin - gridStartMin

  const hours = useMemo(() => {
    const firstHour = Math.ceil(gridStartMin / 60)
    const lastHour = Math.floor(gridEndMin / 60)
    return Array.from({ length: lastHour - firstHour + 1 }, (_, i) => firstHour + i)
  }, [gridStartMin, gridEndMin])

  const { dragState, gridBodyRef, gridContentRef, pointerHandlers } =
    useDragToPaint({
      enabled: isEditing,
      gridStartMin,
      gridEndMin,
      onAddBlock,
    })

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
          isEditing && "cursor-crosshair"
        )}
        style={isEditing ? { touchAction: "none", overscrollBehavior: "none" } : undefined}
        {...pointerHandlers}
      >
        <div
          ref={gridContentRef}
          className="relative grid grid-cols-[2.5rem_repeat(5,1fr)]"
          style={{ minHeight: `${(gridHeight / 60) * GRID.REM_PER_HOUR}rem` }}
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
              {hours.map((hour) => (
                <div
                  key={hour}
                  className="absolute left-0 right-0 border-t border-border/60"
                  style={{ top: `${((hour * 60 - gridStartMin) / gridHeight) * 100}%` }}
                />
              ))}
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
          {blockedTimeGroups?.map((group) => {
            if (!group.enabled && group.id !== editingGroupId) return null
            const isEditingThisGroup = group.id === editingGroupId

            return group.blocks.map((bt, blockIdx) => {
              const btStartMin = parseTime(bt.startTime)
              const btEndMin = parseTime(bt.endTime)
              const durationMin = btEndMin - btStartMin
              const descFirstLine = group.description
                ? group.description.split("\n")[0]
                : ""

              return (
                <TimeBlock
                  key={`blocked-${group.id}-${blockIdx}`}
                  variant="blocked"
                  dayIndex={bt.day}
                  startMin={btStartMin}
                  endMin={btEndMin}
                  gridStartMin={gridStartMin}
                  gridHeight={gridHeight}
                  color={group.color}
                  interactive={!isEditing && !!onBlockedTimeClick && !isEditingThisGroup}
                  dimmed={isEditing && !isEditingThisGroup}
                  editing={isEditingThisGroup}
                  onClick={(e) => {
                    e.stopPropagation()
                    if (isEditingThisGroup) {
                      onRemoveBlock?.(group.id, blockIdx)
                    } else if (!isEditing && onBlockedTimeClick) {
                      onBlockedTimeClick(group.id)
                    }
                  }}
                  tabIndex={isEditingThisGroup || (!isEditing && onBlockedTimeClick) ? 0 : -1}
                  ariaLabel={
                    isEditingThisGroup
                      ? `Remove blocked time on ${DAYS[bt.day]}`
                      : undefined
                  }
                >
                  {!isEditingThisGroup && group.color && (
                    <>
                      {group.title && (
                        <div className="font-semibold leading-tight text-[10px] sm:text-xs overflow-hidden whitespace-nowrap">
                          {group.title}
                        </div>
                      )}
                      {durationMin >= 80 && descFirstLine && (
                        <div className="truncate text-[10px] leading-tight opacity-60">
                          {descFirstLine}
                        </div>
                      )}
                    </>
                  )}
                </TimeBlock>
              )
            })
          })}

          {/* Drag preview for blocked time painting */}
          {dragState && (() => {
            const minMin = Math.min(dragState.startMin, dragState.currentMin)
            const maxMin = Math.max(dragState.startMin, dragState.currentMin)
            const top = ((minMin - gridStartMin) / gridHeight) * 100
            const height = ((maxMin - minMin) / gridHeight) * 100
            const dayWidth = `(100% - ${GRID.TIME_COL}) / ${GRID.DAY_COUNT}`
            const leftOffset = `calc(${GRID.TIME_COL} + (${dayWidth}) * ${dragState.dayIndex})`
            const blockWidth = `calc(${dayWidth})`
            const editingGroup = blockedTimeGroups?.find((g) => g.id === editingGroupId)
            const groupColor = editingGroup?.color

            return (
              <div
                className="pointer-events-none absolute z-20 rounded border-2"
                style={{
                  top: `${top}%`,
                  height: `${height}%`,
                  left: leftOffset,
                  width: blockWidth,
                  borderColor: groupColor ? `${groupColor}99` : "rgba(239, 68, 68, 0.6)",
                  backgroundColor: groupColor ? `${groupColor}33` : "rgba(239, 68, 68, 0.2)",
                }}
              />
            )
          })()}

          {/* Course blocks overlay */}
          {blocks.map((block, idx) => {
            const durationMin = block.endMin - block.startMin

            return (
              <TimeBlock
                key={idx}
                variant="course"
                dayIndex={block.dayIndex}
                startMin={block.startMin}
                endMin={block.endMin}
                gridStartMin={gridStartMin}
                gridHeight={gridHeight}
                color={block.colorClass}
                interactive={!isEditing && !!onCourseClick}
                dimmed={isEditing}
                onClick={() => !isEditing && onCourseClick?.(block.course.crn)}
                tabIndex={isEditing ? -1 : 0}
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
              </TimeBlock>
            )
          })}
        </div>

      </div>
    </div>
  )
}
