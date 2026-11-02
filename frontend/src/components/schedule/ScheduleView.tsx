import { useCallback, useMemo, useRef, useState } from "react"
import { Check, X } from "lucide-react"
import { ScheduleGrid } from "./ScheduleGrid"
import { AsyncCoursesList } from "./AsyncCoursesList"
import { BlockedTimesDialog } from "./BlockedTimesDialog"
import { ScheduleBanner, ScheduleNavHeader } from "./ScheduleNavHeader"
import { useAppStore, type BlockedTimeBlock } from "@/stores/app-store"
import { useCalendar, useTerms } from "@/hooks/use-api"
import { useExportPng } from "@/hooks/use-export-png"
import { hydrateSchedule } from "@/lib/schedule-utils"
import { examSections, finalsWeekDays } from "@/lib/finals"
import { formatDayRange, formatNumericDate } from "@/lib/dates"

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
  const { data: calendar } = useCalendar(term)
  const isGenerateResultStale = useAppStore((s) => s.isGenerateResultStale())
  const canRegenerate = slots.length > 0
  const showStale = isGenerateResultStale && canRegenerate
  const [dialogOpen, setDialogOpen] = useState(false)
  const [finalsView, setFinalsView] = useState(false)
  const [focusedGroupId, setFocusedGroupId] = useState<string | null>(null)
  const gridRef = useRef<HTMLDivElement>(null)
  const termName = termsData?.terms.find((t) => t.code === term)?.name
  const { isExporting, handleDownloadPng } = useExportPng(gridRef, termName)

  const isEditing = editingBlockedTimeGroupId != null
  const editingGroup = blockedTimeGroups.find(
    (g) => g.id === editingBlockedTimeGroupId
  )

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

  const scheduleSections = useMemo(() => currentSchedule?.courses ?? [], [currentSchedule])
  const exams = useMemo(() => examSections(scheduleSections), [scheduleSections])
  const showFinals = finalsView && exams.length > 0 && !isEditing
  const finalsDays = useMemo(() => finalsWeekDays(exams, calendar?.finals ?? null), [exams, calendar])
  const guessedCrns = useMemo(
    () => new Set(exams.filter((e) => !e.final?.certain).map((e) => e.crn)),
    [exams]
  )

  const handleStartPainting = useCallback(
    (groupId: string) => {
      setFinalsView(false)
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

  const handleCourseClick = (crn: string) => {
    openCourseDialog({ crn, source: "schedule" })
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
      {isEditing && (
        <ScheduleBanner onClose={handleDonePainting} closeLabel="Done painting" closeIcon={Check}>
          Painting: {editingGroup?.title || "Blocked Time"}
        </ScheduleBanner>
      )}
      {showFinals && (
        <ScheduleBanner onClose={() => setFinalsView(false)} closeLabel="Close finals week" closeIcon={X}>
          Finals week: {formatDayRange(finalsDays[0], finalsDays[4])}
        </ScheduleBanner>
      )}
      {!isEditing && !showFinals && (
        <ScheduleNavHeader
          canRegenerate={canRegenerate}
          showStale={showStale}
          onRegenerate={requestRegenerate}
          index={safeIndex}
          count={totalSchedules}
          onIndexChange={setCurrentScheduleIndex}
          emptyText={generateResult ? "No valid schedules found" : "No schedules yet"}
          sections={scheduleSections}
          onBlockedTimes={() => setDialogOpen(true)}
          onDownloadPng={handleDownloadPng}
          isExporting={isExporting}
          onFinalsWeek={() => setFinalsView(true)}
          hasFinals={exams.length > 0}
        />
      )}

      {/* Schedule grid — always visible, empty when no schedules */}
      <div className="scrollbar-styled min-h-96 flex-1 overflow-auto">
        <ScheduleGrid
          captureRef={gridRef}
          courses={showFinals ? exams : scheduleSections}
          dayDates={showFinals ? finalsDays.map(formatNumericDate) : undefined}
          guessedCrns={showFinals ? guessedCrns : undefined}
          onCourseClick={handleCourseClick}
          onBlockedTimeClick={handleBlockedTimeClick}
          blockedTimeGroups={blockedTimeGroups}
          editingGroupId={editingBlockedTimeGroupId}
          onAddBlock={handleAddBlock}
          onUpdateBlock={handleUpdateBlock}
          onRemoveBlock={handleRemoveBlock}
        />
        {generateResult && !showFinals && (
          <AsyncCoursesList
            generateResult={generateResult}
            onCourseClick={handleCourseClick}
          />
        )}
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
