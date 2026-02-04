import { useMemo, useState } from "react"
import {
  ChevronLeft,
  ChevronRight,
  RefreshCw,
  MoreVertical,
  Clock,
  Download,
  Calendar,
  Share2,
  Map,
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
import { ScheduleGrid } from "./ScheduleGrid"
import { useAppStore } from "@/stores/app-store"
import { hydrateSchedule } from "@/lib/schedule-utils"
import { cn } from "@/lib/utils"

export function ScheduleView() {
  const {
    generateResult,
    currentScheduleIndex,
    setCurrentScheduleIndex,
    openCourseDialog,
    requestRegenerate,
    slots,
  } = useAppStore()
  const isGenerateResultStale = useAppStore((s) => s.isGenerateResultStale())
  const canRegenerate = slots.length > 0
  const showStale = isGenerateResultStale && canRegenerate
  const [spinCount, setSpinCount] = useState(0)

  const regenerateTooltip = () => {
    if (!canRegenerate) return "Add courses to regenerate"
    if (showStale) return "Course list changed. Click to regenerate."
    return "Regenerate schedules"
  }

  // Hydrate the current schedule ref into full course data
  // Must be called before early return to satisfy Rules of Hooks
  const currentSchedule = useMemo(() => {
    if (!generateResult || generateResult.schedules.length === 0) {
      return null
    }
    const { schedules, courses, sections } = generateResult
    const safeIndex = Math.min(currentScheduleIndex, schedules.length - 1)
    return hydrateSchedule(schedules[safeIndex], courses, sections)
  }, [generateResult, currentScheduleIndex])

  if (!generateResult || generateResult.schedules.length === 0) {
    return (
      <div className="flex h-full flex-col items-center justify-center gap-2 p-8 text-center">
        <p className="text-muted-foreground">No schedules generated yet</p>
        <p className="text-muted-foreground text-sm">
          Add courses in the sidebar and click Generate
        </p>
      </div>
    )
  }

  const { schedules, stats } = generateResult
  const totalSchedules = schedules.length
  const safeIndex = Math.min(currentScheduleIndex, totalSchedules - 1)

  const handlePrev = () => {
    if (safeIndex > 0) {
      setCurrentScheduleIndex(safeIndex - 1)
    }
  }

  const handleNext = () => {
    if (safeIndex < totalSchedules - 1) {
      setCurrentScheduleIndex(safeIndex + 1)
    }
  }

  const handleCourseClick = (crn: string) => {
    openCourseDialog({ crn })
  }

  return (
    <div className="flex h-full flex-col">
      {/* Navigation header */}
      <div className="relative flex items-center justify-center border-b px-4 py-2">
        {/* Regenerate button */}
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
                  requestRegenerate()
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
          <TooltipContent side="bottom">
            {regenerateTooltip()}
          </TooltipContent>
        </Tooltip>

        <Button
          variant="ghost"
          size="icon"
          onClick={handlePrev}
          disabled={safeIndex === 0}
        >
          <ChevronLeft className="size-4" />
        </Button>
        <span className="min-w-32 text-center text-sm">
          Schedule {safeIndex + 1} of {totalSchedules}
        </span>
        <Button
          variant="ghost"
          size="icon"
          onClick={handleNext}
          disabled={safeIndex === totalSchedules - 1}
        >
          <ChevronRight className="size-4" />
        </Button>

        {/* Schedule actions dropdown */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="absolute right-2"
            >
              <MoreVertical className="size-4" />
              <span className="sr-only">Schedule options</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem disabled>
              <Clock className="size-4" />
              Blocked Times
            </DropdownMenuItem>
            <DropdownMenuItem disabled>
              <Download className="size-4" />
              Download PNG
            </DropdownMenuItem>
            <DropdownMenuItem disabled>
              <Calendar className="size-4" />
              Export ICS
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

      {/* Schedule grid */}
      <div className="flex-1 overflow-hidden">
        <ScheduleGrid
          courses={currentSchedule!.courses}
          onCourseClick={handleCourseClick}
        />
      </div>

      {/* Stats footer */}
      <div className="text-muted-foreground border-t px-4 py-2 text-xs">
        Generated {stats.totalGenerated.toLocaleString()} schedules in{" "}
        {stats.timeMs.toFixed(1)}ms
      </div>
    </div>
  )
}
