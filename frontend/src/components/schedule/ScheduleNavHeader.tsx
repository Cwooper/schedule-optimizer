import { useState } from "react"
import {
  ChevronLeft,
  ChevronRight,
  Clock,
  Download,
  GraduationCap,
  Loader2,
  Map,
  MoreVertical,
  RefreshCw,
  Share2,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import type { HydratedSection } from "@/lib/api"
import { cn } from "@/lib/utils"
import { AddToCalendar } from "./AddToCalendar"
import { ScheduleStatsPopover } from "./ScheduleStatsPopover"

/** Replaces the nav header while painting blocked times. */
export function ScheduleBanner({
  children,
  onClose,
  closeLabel,
  closeIcon: CloseIcon,
}: {
  children: React.ReactNode
  onClose: () => void
  closeLabel: string
  closeIcon: React.ComponentType<{ className?: string }>
}) {
  return (
    <div className="flex items-center justify-between border-b bg-muted/40 px-4 py-2">
      <span className="text-sm font-medium">{children}</span>
      <Button size="icon" variant="ghost" onClick={onClose}>
        <CloseIcon className="size-4" />
        <span className="sr-only">{closeLabel}</span>
      </Button>
    </div>
  )
}

interface ScheduleNavHeaderProps {
  canRegenerate: boolean
  hasInvalidBounds: boolean
  showStale: boolean
  onRegenerate: () => void
  /** Index of the shown schedule, and how many there are (0 when none). */
  index: number
  count: number
  onIndexChange: (index: number) => void
  emptyText: string
  /** The shown schedule's sections and ranking score, for its stats and calendar export. */
  sections: HydratedSection[]
  score: number
  onBlockedTimes: () => void
  onDownloadPng: () => void
  isExporting: boolean
  finalsWeek: boolean
  hasFinals: boolean
  onFinalsWeekChange: (on: boolean) => void
}

export function ScheduleNavHeader({
  canRegenerate,
  hasInvalidBounds,
  showStale,
  onRegenerate,
  index,
  count,
  onIndexChange,
  emptyText,
  sections,
  score,
  onBlockedTimes,
  onDownloadPng,
  isExporting,
  finalsWeek,
  hasFinals,
  onFinalsWeekChange,
}: ScheduleNavHeaderProps) {
  const [spinCount, setSpinCount] = useState(0)

  const regenerateTooltip = () => {
    if (hasInvalidBounds) return "Max courses must be at least min courses"
    if (!canRegenerate) return "Add courses to regenerate"
    if (showStale) return "Settings changed. Click to regenerate."
    return "Regenerate schedules"
  }

  // A grid rather than absolute side groups, so on a narrow screen the pager
  // gives way to them instead of sliding underneath; below sm it also drops
  // "Schedule" from its label to fit at 320px.
  return (
    <div className="grid grid-cols-[1fr_auto_1fr] items-center border-b p-2">
      <Tooltip>
        <TooltipTrigger asChild>
          <div className="justify-self-start">
            <Button
              variant="ghost"
              size="icon"
              className={cn(
                showStale &&
                  "text-amber-600 hover:text-amber-700 dark:text-amber-400 dark:hover:text-amber-300"
              )}
              onClick={() => {
                setSpinCount((c) => c + 1)
                onRegenerate()
              }}
              disabled={!canRegenerate}
            >
              <RefreshCw
                key={spinCount}
                className={cn("size-4", spinCount > 0 && "animate-spin-once")}
              />
              <span className="sr-only">Regenerate schedules</span>
            </Button>
          </div>
        </TooltipTrigger>
        <TooltipContent side="bottom">{regenerateTooltip()}</TooltipContent>
      </Tooltip>

      {count > 0 ? (
        <div className="flex items-center">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => onIndexChange(index - 1)}
            disabled={index === 0}
          >
            <ChevronLeft className="size-4" />
          </Button>
          <ScheduleStatsPopover index={index} count={count} sections={sections} score={score} />
          <Button
            variant="ghost"
            size="icon"
            onClick={() => onIndexChange(index + 1)}
            disabled={index === count - 1}
          >
            <ChevronRight className="size-4" />
          </Button>
        </div>
      ) : (
        <span className="text-muted-foreground min-w-0 text-sm leading-9 max-sm:truncate">{emptyText}</span>
      )}

      <div className="flex items-center justify-self-end">
        <Tooltip>
          <TooltipTrigger asChild>
            <div>
              <Button
                variant="ghost"
                size="icon"
                aria-pressed={finalsWeek}
                className="aria-pressed:bg-accent aria-pressed:text-accent-foreground"
                disabled={!hasFinals}
                onClick={() => onFinalsWeekChange(!finalsWeek)}
              >
                <GraduationCap className="size-4" />
                <span className="sr-only">Finals week</span>
              </Button>
            </div>
          </TooltipTrigger>
          <TooltipContent side="bottom">Finals week</TooltipContent>
        </Tooltip>
        <AddToCalendar sections={sections} />
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon">
              <MoreVertical className="size-4" />
              <span className="sr-only">Schedule options</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={onBlockedTimes}>
              <Clock className="size-4" />
              Blocked Times
            </DropdownMenuItem>
            <DropdownMenuItem disabled={count === 0 || isExporting} onClick={onDownloadPng}>
              {isExporting ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <Download className="size-4" />
              )}
              {isExporting ? "Exporting..." : "Download PNG"}
            </DropdownMenuItem>
            <DropdownMenuItem disabled>
              <Share2 className="size-4" />
              Share Link
            </DropdownMenuItem>
            <DropdownMenuItem disabled>
              <Map className="size-4" />
              Campus Map
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  )
}
