import { memo, useState } from "react"
import { ChevronDown, ChevronRight, Check, Copy, Plus, Users, User, BookOpen, TrendingUp, CircleCheck, CircleHelp, type LucideIcon } from "lucide-react"
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import type { Final, HydratedSection, MeetingTime } from "@/lib/api"
import { cn } from "@/lib/utils"
import { describeGrades, type GradeKind } from "@/lib/grades"
import { formatNumericDate, formatWeekday } from "@/lib/dates"
import { FINAL_GUESS_REASONS } from "@/lib/finals"

interface SectionCardProps {
  section: HydratedSection
  expanded?: boolean
  onToggleExpand?: () => void
  highlighted?: boolean
  isLoadingDetails?: boolean
  /** Called on hover or focus to prefetch section details */
  onPrefetch?: () => void
  /** Called when "+" button is clicked to add this section */
  onAdd?: () => void
  /** Whether this section is already added to schedule */
  isAdded?: boolean
  /** Optional term label shown when sections span multiple terms */
  termLabel?: string
}

// Single-letter day abbreviations (R = Thursday to distinguish from Tuesday)
const DAYS_LETTER = ["U", "M", "T", "W", "R", "F", "S"]

function formatDays(days: boolean[]): string {
  return days
    .map((active, i) => (active ? DAYS_LETTER[i] : null))
    .filter(Boolean)
    .join("")
}

function formatTime(time: string): string {
  if (!time) return ""
  const clean = time.replace(":", "").padStart(4, "0")
  const hours = parseInt(clean.slice(0, 2), 10)
  const mins = clean.slice(2, 4)
  const suffix = hours >= 12 ? "pm" : "am"
  let displayHour = hours
  if (hours > 12) {
    displayHour = hours - 12
  } else if (hours === 0) {
    displayHour = 12
  }
  return `${displayHour}:${mins}${suffix}`
}

function isAsync(meeting: MeetingTime): boolean {
  return meeting.room === "ASNC"
}

function formatMeetingTime(meeting: MeetingTime): string {
  const { startTime, endTime } = meeting
  if (meeting.days.slice(1, 6).some(Boolean) && startTime && endTime) {
    return `${formatDays(meeting.days)} ${formatTime(startTime)}-${formatTime(endTime)}`
  }
  if (isAsync(meeting)) return "Asynchronous, no set times"
  return "Times arranged with instructor"
}

function formatLocation(meeting: MeetingTime): string | null {
  const building = meeting.building || null
  // Don't show ASNC/SYNC as room names — they're schedule modifiers, not locations
  const room =
    meeting.room && meeting.room !== "ASNC" && meeting.room !== "SYNC"
      ? meeting.room
      : null
  if (!building && !room) return null
  if (!building) return room
  if (!room) return building
  return `${building} ${room}`
}

function formatEnrollment(section: HydratedSection): string {
  // Banner reports seatsAvailable as open (remaining) seats; show open/capacity
  // to match what students see in Banner (issue #53).
  const { seatsAvailable, maxEnrollment, waitCount } = section
  if (waitCount > 0) {
    return `${seatsAvailable}/${maxEnrollment} seats open (${waitCount} waitlist)`
  }
  return `${seatsAvailable}/${maxEnrollment} seats open`
}

function FinalRow({ final }: { final: Final }) {
  return (
    <div className="flex items-center gap-1.5 font-mono text-sm">
      <span>
        <span className="text-muted-foreground">Final</span> {formatWeekday(final.date)}{" "}
        {formatNumericDate(final.date)} {formatTime(final.start)}-{formatTime(final.end)}
      </span>
      {!final.certain && final.reason && (
        <Tooltip>
          <TooltipTrigger asChild>
            <span className="inline-flex cursor-help" aria-label="Estimated final">
              <CircleHelp className="size-3 text-muted-foreground" />
            </span>
          </TooltipTrigger>
          <TooltipContent>{FINAL_GUESS_REASONS[final.reason]}</TooltipContent>
        </Tooltip>
      )}
    </div>
  )
}

const GRADE_ICONS: Record<GradeKind, LucideIcon> = {
  professor: User,
  course: BookOpen,
  pass: CircleCheck,
  none: TrendingUp,
}

export const SectionCard = memo(function SectionCard({
  section,
  expanded = false,
  onToggleExpand,
  highlighted = false,
  isLoadingDetails = false,
  onPrefetch,
  onAdd,
  isAdded = false,
  termLabel,
}: SectionCardProps) {
  const [copied, setCopied] = useState(false)
  const hasMeetings = section.meetingTimes.length > 0
  const isClickable = !!onToggleExpand
  const grades = describeGrades(section.gpa, section.passRate, section.gpaSource)
  const GradeIcon = GRADE_ICONS[grades.kind]

  const handleAdd = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (!isAdded) {
      onAdd?.()
    }
  }

  const handleCopy = async (e: React.MouseEvent | React.KeyboardEvent) => {
    e.stopPropagation()
    try {
      await navigator.clipboard.writeText(section.crn)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {
      // Clipboard API may fail in non-secure contexts
    }
  }

  return (
    <Collapsible open={expanded} onOpenChange={onToggleExpand}>
      <div
        className={cn(
          "rounded-lg border transition-colors",
          highlighted && "ring-2 ring-primary",
          expanded && "bg-muted/30"
        )}
        onMouseEnter={onPrefetch}
        onFocus={onPrefetch}
      >
        {/* Header - always visible */}
        <div className="flex items-center gap-3 p-3">
          <CollapsibleTrigger
            className={cn(
              "flex min-w-0 flex-1 items-center gap-3 text-left",
              isClickable && "cursor-pointer hover:bg-muted/50 rounded-md -m-1.5 p-1.5"
            )}
            disabled={!isClickable}
          >
            {/* Expand indicator */}
            <div className="shrink-0 text-muted-foreground">
              {expanded ? (
                <ChevronDown className="size-4" />
              ) : (
                <ChevronRight className="size-4" />
              )}
            </div>

            {/* CRN + copy + instructor */}
            <div className="flex min-w-0 flex-1 items-baseline gap-1.5">
              <span className="shrink-0 font-mono text-sm font-medium">
                {section.crn}
              </span>
              <div
                role="button"
                tabIndex={0}
                className="shrink-0 inline-flex size-5 translate-y-0.5 items-center justify-center rounded hover:bg-muted"
                onClick={handleCopy}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault()
                    handleCopy(e)
                  }
                }}
                title="Copy CRN"
                aria-label={copied ? "CRN copied" : "Copy CRN to clipboard"}
              >
                {copied ? (
                  <Check className="size-3 text-emerald-600" />
                ) : (
                  <Copy className="size-3 text-muted-foreground" />
                )}
              </div>
              {section.instructor && (
                <span className="hidden min-w-0 items-baseline gap-1.5 sm:flex">
                  <span className="text-muted-foreground">·</span>
                  <span className="max-w-[20ch] truncate text-sm text-muted-foreground" title={section.instructor}>
                    {section.instructor}
                  </span>
                </span>
              )}
              {termLabel && (
                <>
                  <span className="text-muted-foreground">·</span>
                  <span className="shrink-0 text-xs text-muted-foreground">
                    {termLabel}
                  </span>
                </>
              )}
            </div>

            {/* Status badge */}
            <div
              className={cn(
                "shrink-0 rounded-full px-1.5 py-0.5 text-xs font-medium sm:px-2",
                section.isOpen
                  ? "bg-emerald-500/15 text-emerald-700 dark:text-emerald-400"
                  : "bg-red-500/15 text-red-700 dark:text-red-400"
              )}
            >
              {section.isOpen ? "Open" : "Closed"}
            </div>
          </CollapsibleTrigger>

          {/* Add button - outside the trigger to avoid nested buttons */}
          {onAdd && (
            <Button
              variant="ghost"
              size="icon"
              className={cn(
                "size-7 shrink-0",
                isAdded && "text-emerald-600 dark:text-emerald-400"
              )}
              onClick={handleAdd}
              disabled={isAdded}
              title={isAdded ? "Already in schedule" : "Add this section to schedule"}
            >
              {isAdded ? <Check className="size-4" /> : <Plus className="size-4" />}
            </Button>
          )}
        </div>

        {/* Expanded details */}
        <CollapsibleContent>
          <div className="space-y-2 px-3 pb-3 pl-10">
            {/* Loading skeleton for meeting times */}
            {isLoadingDetails && !hasMeetings && (
              <div className="space-y-1">
                <div className="h-5 w-40 animate-pulse rounded bg-muted" />
                <div className="h-5 w-32 animate-pulse rounded bg-muted" />
              </div>
            )}

            {/* Meeting times - only shown when available */}
            {hasMeetings && (
              <div className="space-y-1">
                {section.meetingTimes.map((meeting, idx) => {
                  const location = formatLocation(meeting)
                  return (
                    <div
                      key={idx}
                      className="flex items-baseline justify-between gap-4 font-mono text-sm"
                    >
                      <span>{formatMeetingTime(meeting)}</span>
                      {location && (
                        <span className="text-muted-foreground">{location}</span>
                      )}
                    </div>
                  )
                })}
                {section.final && <FinalRow final={section.final} />}
              </div>
            )}

            {/* No meeting times message - shown when not loading and no meetings */}
            {!isLoadingDetails && !hasMeetings && (
              <div className="text-sm italic text-muted-foreground">
                No scheduled meeting times
              </div>
            )}

            {/* Enrollment + GPA badges */}
            <div className="flex flex-wrap items-center gap-1.5">
              <Badge
                variant="secondary"
                className={cn(
                  "gap-1 rounded-md text-xs",
                  section.seatsAvailable <= 0 && "text-amber-600 dark:text-amber-400"
                )}
              >
                <Users className="size-3" />
                {formatEnrollment(section)}
              </Badge>
              <Badge variant="secondary" className="gap-1 rounded-md text-xs">
                <GradeIcon className="size-3" />
                {grades.label}
                <Tooltip>
                  <TooltipTrigger asChild>
                    <span className="inline-flex cursor-help">
                      <CircleHelp className="size-3 text-muted-foreground" />
                    </span>
                  </TooltipTrigger>
                  <TooltipContent>{grades.tooltip}</TooltipContent>
                </Tooltip>
              </Badge>
            </div>
          </div>
        </CollapsibleContent>
      </div>
    </Collapsible>
  )
})
