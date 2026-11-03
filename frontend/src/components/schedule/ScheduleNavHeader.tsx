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
  showStale: boolean
  onRegenerate: () => void
  /** Index of the shown schedule, and how many there are (0 when none). */
  index: number
  count: number
  onIndexChange: (index: number) => void
  emptyText: string
  /** The shown schedule's sections, for calendar export. */
  sections: HydratedSection[]
  onBlockedTimes: () => void
  onDownloadPng: () => void
  isExporting: boolean
  finalsWeek: boolean
  hasFinals: boolean
  onFinalsWeekChange: (on: boolean) => void
}

export function ScheduleNavHeader({
  canRegenerate,
  showStale,
  onRegenerate,
  index,
  count,
  onIndexChange,
  emptyText,
  sections,
  onBlockedTimes,
  onDownloadPng,
  isExporting,
  finalsWeek,
  hasFinals,
  onFinalsWeekChange,
}: ScheduleNavHeaderProps) {
  const [spinCount, setSpinCount] = useState(0)

  const regenerateTooltip = () => {
    if (!canRegenerate) return "Add courses to regenerate"
    if (showStale) return "Course list changed. Click to regenerate."
    return "Regenerate schedules"
  }

  return (
    <div className="relative flex items-center justify-center border-b px-4 py-2">
      <Tooltip>
        <TooltipTrigger asChild>
          <div className="absolute left-2">
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
        <>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => onIndexChange(index - 1)}
            disabled={index === 0}
          >
            <ChevronLeft className="size-4" />
          </Button>
          <span className="min-w-32 text-center text-sm">
            Schedule {index + 1} of {count}
          </span>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => onIndexChange(index + 1)}
            disabled={index === count - 1}
          >
            <ChevronRight className="size-4" />
          </Button>
        </>
      ) : (
        <span className="text-muted-foreground flex h-9 items-center text-sm">{emptyText}</span>
      )}

      <div className="absolute right-2 flex items-center">
        <Tooltip>
          <TooltipTrigger asChild>
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
