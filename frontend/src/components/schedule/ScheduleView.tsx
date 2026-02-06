import { useCallback, useMemo, useRef, useState } from "react"
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
  Check,
  Loader2,
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
import { BlockedTimesDialog } from "./BlockedTimesDialog"
import { useAppStore, type BlockedTimeBlock } from "@/stores/app-store"
import { useTerms } from "@/hooks/use-api"
import { useExportPng } from "@/hooks/use-export-png"
import { hydrateSchedule } from "@/lib/schedule-utils"
import { cn } from "@/lib/utils"

// eslint-disable-next-line sonarjs/cognitive-complexity
export function ScheduleView() {
  const {
    generateResult,
    currentScheduleIndex,
    setCurrentScheduleIndex,
    openCourseDialog,
    requestRegenerate,
    slots,
    blockedTimeGroups,
    editingBlockedTimeGroupId,
    setEditingBlockedTimeGroupId,
    addBlockToGroup,
    removeBlockedTimeGroup,
    updateBlockInGroupById,
    removeBlockFromGroupById,
  } = useAppStore()
  const term = useAppStore((s) => s.term)
  const { data: termsData } = useTerms()
  const isGenerateResultStale = useAppStore((s) => s.isGenerateResultStale())
  const canRegenerate = slots.length > 0
  const showStale = isGenerateResultStale && canRegenerate
  const [spinCount, setSpinCount] = useState(0)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [focusedGroupId, setFocusedGroupId] = useState<string | null>(null)
  const gridRef = useRef<HTMLDivElement>(null)
  const termName = termsData?.terms.find((t) => t.code === term)?.name
  const { isExporting, handleDownloadPng } = useExportPng(gridRef, termName)

  const isEditing = editingBlockedTimeGroupId != null
  const editingGroup = blockedTimeGroups.find(
    (g) => g.id === editingBlockedTimeGroupId
  )

  const regenerateTooltip = () => {
    if (!canRegenerate) return "Add courses to regenerate"
    if (showStale) return "Course list changed. Click to regenerate."
    return "Regenerate schedules"
  }

  const hasSchedules =
    generateResult != null && generateResult.schedules.length > 0

  const currentSchedule = useMemo(() => {
    if (!generateResult || generateResult.schedules.length === 0) {
      return null
    }
    const { schedules, courses, sections } = generateResult
    const safeIndex = Math.min(currentScheduleIndex, schedules.length - 1)
    return hydrateSchedule(schedules[safeIndex], courses, sections)
  }, [generateResult, currentScheduleIndex])

  const totalSchedules = hasSchedules ? generateResult.schedules.length : 0
  const safeIndex = hasSchedules
    ? Math.min(currentScheduleIndex, totalSchedules - 1)
    : 0

  const handleStartPainting = useCallback(
    (groupId: string) => {
      setEditingBlockedTimeGroupId(groupId)
    },
    [setEditingBlockedTimeGroupId]
  )

  const handleDonePainting = useCallback(() => {
    if (editingBlockedTimeGroupId) {
      // Clean up empty groups created by "Add Blocked Time" with no painting
      const group = blockedTimeGroups.find(
        (g) => g.id === editingBlockedTimeGroupId
      )
      if (group && group.blocks.length === 0) {
        removeBlockedTimeGroup(editingBlockedTimeGroupId)
      }
    }
    setEditingBlockedTimeGroupId(null)
    setDialogOpen(true)
  }, [
    editingBlockedTimeGroupId,
    blockedTimeGroups,
    removeBlockedTimeGroup,
    setEditingBlockedTimeGroupId,
  ])

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

  const handleBlockedTimeClick = (groupId: string) => {
    setFocusedGroupId(groupId)
    setDialogOpen(true)
  }

  const handleAddBlock = useCallback(
    (block: BlockedTimeBlock) => {
      if (editingBlockedTimeGroupId) {
        addBlockToGroup(editingBlockedTimeGroupId, block)
      }
    },
    [editingBlockedTimeGroupId, addBlockToGroup]
  )

  const handleUpdateBlock = useCallback(
    (blockId: string, block: BlockedTimeBlock) => {
      if (editingBlockedTimeGroupId) {
        updateBlockInGroupById(editingBlockedTimeGroupId, blockId, block)
      }
    },
    [editingBlockedTimeGroupId, updateBlockInGroupById]
  )

  const handleRemoveBlock = useCallback(
    (blockId: string) => {
      if (editingBlockedTimeGroupId) {
        removeBlockFromGroupById(editingBlockedTimeGroupId, blockId)
      }
    },
    [editingBlockedTimeGroupId, removeBlockFromGroupById]
  )

  return (
    <div className="flex h-full flex-col">
      {/* Paint mode banner — replaces navigation header */}
      {isEditing ? (
        <div className="flex items-center justify-between border-b bg-muted/40 px-4 py-2">
          <span className="text-sm font-medium">
            Painting: {editingGroup?.title || "Blocked Time"}
          </span>
          <Button size="icon" variant="ghost" onClick={handleDonePainting}>
            <Check className="size-4" />
            <span className="sr-only">Done painting</span>
          </Button>
        </div>
      ) : (
        /* Navigation header */
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

          {hasSchedules ? (
            <>
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
            </>
          ) : (
            <span className="text-muted-foreground flex h-9 items-center text-sm">
              {generateResult ? "No valid schedules found" : "No schedules generated"}
            </span>
          )}

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
              <DropdownMenuItem onClick={() => setDialogOpen(true)}>
                <Clock className="size-4" />
                Blocked Times
              </DropdownMenuItem>
              <DropdownMenuItem
                disabled={!hasSchedules || isExporting}
                onClick={handleDownloadPng}
              >
                {isExporting ? (
                  <Loader2 className="size-4 animate-spin" />
                ) : (
                  <Download className="size-4" />
                )}
                {isExporting ? "Exporting..." : "Download PNG"}
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
      )}

      {/* Schedule grid — always visible, empty when no schedules */}
      <div className="scrollbar-styled min-h-96 flex-1 overflow-auto">
        <ScheduleGrid
          captureRef={gridRef}
          courses={currentSchedule?.courses ?? []}
          onCourseClick={handleCourseClick}
          onBlockedTimeClick={handleBlockedTimeClick}
          blockedTimeGroups={blockedTimeGroups}
          editingGroupId={editingBlockedTimeGroupId}
          onAddBlock={handleAddBlock}
          onUpdateBlock={handleUpdateBlock}
          onRemoveBlock={handleRemoveBlock}
        />
      </div>

      {/* Stats footer — only when schedules exist */}
      {hasSchedules && (
        <div className="text-muted-foreground border-t px-4 py-2 text-xs">
          Generated {generateResult.stats.totalGenerated.toLocaleString()}{" "}
          schedules in {generateResult.stats.timeMs.toFixed(1)}ms
        </div>
      )}

      {/* Blocked Times Dialog */}
      <BlockedTimesDialog
        open={dialogOpen}
        onOpenChange={(open) => {
          setDialogOpen(open)
          if (!open) setFocusedGroupId(null)
        }}
        onStartPainting={handleStartPainting}
        initialExpandedId={focusedGroupId}
      />
    </div>
  )
}
