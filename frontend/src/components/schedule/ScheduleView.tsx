import { useMemo } from "react"
import { ChevronLeft, ChevronRight, AlertTriangle } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { ScheduleGrid } from "./ScheduleGrid"
import { useAppStore } from "@/stores/app-store"
import { hydrateSchedule } from "@/lib/schedule-utils"

export function ScheduleView() {
  const {
    generateResult,
    isGenerateResultStale,
    currentScheduleIndex,
    setCurrentScheduleIndex,
  } = useAppStore()

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
        <p className="text-sm text-muted-foreground">
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

  return (
    <div className="flex h-full flex-col">
      {/* Navigation header */}
      <div className="flex items-center justify-center border-b px-4 py-2">
        <Button
          variant="ghost"
          size="icon"
          onClick={handlePrev}
          disabled={safeIndex === 0}
        >
          <ChevronLeft className="size-4" />
        </Button>
        <span className="min-w-[8rem] text-center text-sm">
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
      </div>

      {/* Schedule grid */}
      <div className="flex-1 overflow-hidden">
        <ScheduleGrid
          courses={currentSchedule!.courses}
          cornerContent={
            isGenerateResultStale ? (
              <Tooltip>
                <TooltipTrigger asChild>
                  <div className="cursor-help rounded-full bg-amber-500/15 p-1 text-amber-600 dark:text-amber-400">
                    <AlertTriangle className="size-4" />
                  </div>
                </TooltipTrigger>
                <TooltipContent side="right" className="max-w-[200px]">
                  <p>
                    Course list changed. Click Generate to see updated schedules.
                  </p>
                </TooltipContent>
              </Tooltip>
            ) : undefined
          }
        />
      </div>

      {/* Stats footer */}
      <div className="border-t px-4 py-2 text-xs text-muted-foreground">
        Generated {stats.totalGenerated.toLocaleString()} schedules in{" "}
        {stats.timeMs.toFixed(1)}ms
      </div>
    </div>
  )
}
