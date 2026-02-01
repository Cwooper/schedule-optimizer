import { useMemo } from "react"
import type { ScheduleSection, MeetingTime } from "@/lib/api"
import { cn } from "@/lib/utils"

interface ScheduleGridProps {
  courses: ScheduleSection[]
}

// Grid constants
const DAYS = ["Mon", "Tue", "Wed", "Thu", "Fri"]
const START_HOUR = 7 // 7am
const END_HOUR = 22 // 10pm
const HOURS = Array.from({ length: END_HOUR - START_HOUR }, (_, i) => START_HOUR + i)

// Color palette for courses - use hash of course key for stable assignment
const COURSE_COLORS = [
  "bg-blue-500/20 border-blue-500 text-blue-900 dark:text-blue-100",
  "bg-emerald-500/20 border-emerald-500 text-emerald-900 dark:text-emerald-100",
  "bg-amber-500/20 border-amber-500 text-amber-900 dark:text-amber-100",
  "bg-purple-500/20 border-purple-500 text-purple-900 dark:text-purple-100",
  "bg-rose-500/20 border-rose-500 text-rose-900 dark:text-rose-100",
  "bg-cyan-500/20 border-cyan-500 text-cyan-900 dark:text-cyan-100",
  "bg-orange-500/20 border-orange-500 text-orange-900 dark:text-orange-100",
  "bg-indigo-500/20 border-indigo-500 text-indigo-900 dark:text-indigo-100",
  "bg-teal-500/20 border-teal-500 text-teal-900 dark:text-teal-100",
  "bg-pink-500/20 border-pink-500 text-pink-900 dark:text-pink-100",
  "bg-lime-500/20 border-lime-500 text-lime-900 dark:text-lime-100",
  "bg-sky-500/20 border-sky-500 text-sky-900 dark:text-sky-100",
  "bg-violet-500/20 border-violet-500 text-violet-900 dark:text-violet-100",
  "bg-fuchsia-500/20 border-fuchsia-500 text-fuchsia-900 dark:text-fuchsia-100",
  "bg-red-500/20 border-red-500 text-red-900 dark:text-red-100",
  "bg-yellow-500/20 border-yellow-500 text-yellow-900 dark:text-yellow-100",
]

// Simple string hash for stable color assignment
function hashString(str: string): number {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i)
    hash = (hash << 5) - hash + char
    hash = hash & hash // Convert to 32-bit integer
  }
  return Math.abs(hash)
}

function parseTime(time: string): number {
  if (!time) return 0
  // "0900" or "09:00" -> minutes from midnight
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
  course: ScheduleSection
  meeting: MeetingTime
  dayIndex: number
  startMin: number
  endMin: number
  colorClass: string
}

// Build color map for courses using hash for stable color assignment
function buildColorMap(courses: ScheduleSection[]): Map<string, string> {
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

// Convert meeting days to individual blocks
function meetingToBlocks(
  course: ScheduleSection,
  meeting: MeetingTime,
  colorClass: string
): CourseBlock[] {
  const startMin = parseTime(meeting.startTime)
  const endMin = parseTime(meeting.endTime)
  const blocks: CourseBlock[] = []

  // days is [Sun, Mon, Tue, Wed, Thu, Fri, Sat]
  for (let dayIdx = 1; dayIdx <= 5; dayIdx++) {
    if (meeting.days[dayIdx]) {
      blocks.push({
        course,
        meeting,
        dayIndex: dayIdx - 1, // Convert to 0-4 (Mon-Fri)
        startMin,
        endMin,
        colorClass,
      })
    }
  }
  return blocks
}

export function ScheduleGrid({ courses }: ScheduleGridProps) {
  // Build course blocks with positions and colors
  const blocks = useMemo(() => {
    const colorMap = buildColorMap(courses)

    return courses.flatMap((course) => {
      const colorClass = colorMap.get(`${course.subject}:${course.courseNumber}`)!
      return course.meetingTimes.flatMap((meeting) =>
        meetingToBlocks(course, meeting, colorClass)
      )
    })
  }, [courses])

  // Calculate grid dimensions
  const gridStartMin = START_HOUR * 60
  const gridEndMin = END_HOUR * 60
  const gridHeight = gridEndMin - gridStartMin

  return (
    <div className="flex h-full flex-col overflow-hidden">
      {/* Header row with day names */}
      <div className="border-b bg-muted/30">
        <div className="grid grid-cols-[3rem_repeat(5,1fr)]">
          <div className="p-2" /> {/* Time column header */}
          {DAYS.map((day) => (
            <div
              key={day}
              className="border-l p-2 text-center text-sm font-medium"
            >
              {day}
            </div>
          ))}
        </div>
      </div>

      {/* Scrollable grid body */}
      <div className="relative flex-1 overflow-auto">
        <div className="grid grid-cols-[3rem_repeat(5,1fr)]" style={{ minHeight: `${HOURS.length * 3}rem` }}>
          {/* Time labels column */}
          <div className="relative">
            {HOURS.map((hour) => (
              <div
                key={hour}
                className="absolute right-1 -translate-y-1/2 text-xs text-muted-foreground"
                style={{ top: `${((hour - START_HOUR) / (END_HOUR - START_HOUR)) * 100}%` }}
              >
                {formatHour(hour)}{formatAmPm(hour)}
              </div>
            ))}
          </div>

          {/* Day columns with hour lines */}
          {DAYS.map((_, dayIdx) => (
            <div key={dayIdx} className="relative border-l">
              {/* Hour lines */}
              {HOURS.map((hour) => (
                <div
                  key={hour}
                  className="absolute left-0 right-0 border-t border-dashed border-muted"
                  style={{ top: `${((hour - START_HOUR) / (END_HOUR - START_HOUR)) * 100}%` }}
                />
              ))}
            </div>
          ))}

          {/* Course blocks overlay */}
          {blocks.map((block, idx) => {
            const top = ((block.startMin - gridStartMin) / gridHeight) * 100
            const height = ((block.endMin - block.startMin) / gridHeight) * 100
            // Each day column is 1/5 of (100% - 3rem)
            const dayWidth = "(100% - 3rem) / 5"
            const leftOffset = `calc(3rem + (${dayWidth}) * ${block.dayIndex} + 2px)`
            const blockWidth = `calc(${dayWidth} - 4px)`

            return (
              <div
                key={idx}
                className={cn(
                  "absolute rounded border-l-2 px-1 py-0.5 text-xs overflow-hidden",
                  block.colorClass
                )}
                style={{
                  top: `${top}%`,
                  height: `${height}%`,
                  left: leftOffset,
                  width: blockWidth,
                }}
              >
                <div className="font-medium truncate">
                  {block.course.subject} {block.course.courseNumber}
                </div>
                <div className="truncate opacity-75">
                  {block.meeting.building} {block.meeting.room}
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
