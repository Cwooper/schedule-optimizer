import { memo, useState } from "react"
import { ChevronDown, ChevronRight, Check, Copy, Plus } from "lucide-react"
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"
import { Button } from "@/components/ui/button"
import type { HydratedSection, MeetingTime } from "@/lib/api"
import { cn, decodeHtmlEntities } from "@/lib/utils"

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

function hasScheduledTime(meeting: MeetingTime): boolean {
  const hasWeekday = meeting.days.slice(1, 6).some(Boolean)
  return hasWeekday && !!meeting.startTime && !!meeting.endTime
}

function isAsync(meeting: MeetingTime): boolean {
  return meeting.room === "ASNC"
}

function formatMeetingTime(meeting: MeetingTime): string {
  if (hasScheduledTime(meeting)) {
    const days = formatDays(meeting.days)
    const time = `${formatTime(meeting.startTime)}-${formatTime(meeting.endTime)}`
    return `${days} ${time}`
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
  const { enrollment, maxEnrollment, waitCount } = section
  if (waitCount > 0) {
    return `${enrollment}/${maxEnrollment} seats (${waitCount} waitlist)`
  }
  return `${enrollment}/${maxEnrollment} seats`
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
        <CollapsibleTrigger
          className={cn(
            "flex w-full items-center gap-3 p-3 text-left",
            isClickable && "cursor-pointer hover:bg-muted/50"
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
              <>
                <span className="text-muted-foreground">·</span>
                <span className="truncate text-sm text-muted-foreground">
                  {decodeHtmlEntities(section.instructor)}
                </span>
              </>
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
              "shrink-0 rounded-full px-2 py-0.5 text-xs font-medium",
              section.isOpen
                ? "bg-emerald-500/15 text-emerald-700 dark:text-emerald-400"
                : "bg-red-500/15 text-red-700 dark:text-red-400"
            )}
          >
            {section.isOpen ? "Open" : "Closed"}
          </div>

          {/* Add button */}
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
        </CollapsibleTrigger>

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
              </div>
            )}

            {/* No meeting times message - shown when not loading and no meetings */}
            {!isLoadingDetails && !hasMeetings && (
              <div className="text-sm italic text-muted-foreground">
                No scheduled meeting times
              </div>
            )}

            {/* Enrollment */}
            <div
              className={cn(
                "text-sm",
                section.seatsAvailable <= 0
                  ? "text-amber-600 dark:text-amber-400"
                  : "text-muted-foreground"
              )}
            >
              {formatEnrollment(section)}
            </div>
          </div>
        </CollapsibleContent>
      </div>
    </Collapsible>
  )
})
