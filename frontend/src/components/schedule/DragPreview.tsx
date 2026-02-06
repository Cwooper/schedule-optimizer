import type { DragState } from "@/hooks/use-drag-to-paint"
import type { BlockedTimeGroup } from "@/stores/app-store"
import { GRID, opacityToHex } from "@/lib/schedule-utils"

interface DragPreviewProps {
  dragState: DragState
  editingGroup: BlockedTimeGroup | undefined
  gridStartMin: number
  gridEndMin: number
  gridHeight: number
}

export function DragPreview({
  dragState,
  editingGroup,
  gridStartMin,
  gridEndMin,
  gridHeight,
}: DragPreviewProps) {
  const groupColor = editingGroup?.color
  const groupOpacity = editingGroup?.opacity ?? 20
  const previewBgHex = opacityToHex(groupOpacity)
  const previewBorderHex = opacityToHex(Math.min(groupOpacity + 40, 100))
  const dayWidth = `(100% - ${GRID.TIME_COL}) / ${GRID.DAY_COUNT}`
  const borderColor = groupColor ? `${groupColor}${previewBorderHex}` : "rgba(239, 68, 68, 0.6)"
  const bgColor = groupColor ? `${groupColor}${previewBgHex}` : "rgba(239, 68, 68, 0.2)"

  if (dragState.mode === "move") {
    const duration = dragState.blockDuration
    const moveStart = dragState.currentMin - dragState.offsetFromTop
    const snappedStart = Math.round(moveStart / GRID.SNAP_MINUTES) * GRID.SNAP_MINUTES
    const clampedStart = Math.max(gridStartMin, Math.min(gridEndMin - duration, snappedStart))
    const moveDay = dragState.currentDay

    const previewTop = ((clampedStart - gridStartMin) / gridHeight) * 100
    const previewHeight = (duration / gridHeight) * 100
    const previewLeft = `calc(${GRID.TIME_COL} + (${dayWidth}) * ${moveDay})`

    const origStart = dragState.startMin
    const ghostTop = ((origStart - gridStartMin) / gridHeight) * 100
    const ghostHeight = (duration / gridHeight) * 100
    const ghostLeft = `calc(${GRID.TIME_COL} + (${dayWidth}) * ${dragState.originalDay})`

    return (
      <>
        {/* Ghost at original position */}
        <div
          className="pointer-events-none absolute z-10 rounded border-2 border-dashed opacity-30"
          style={{
            top: `${ghostTop}%`,
            height: `${ghostHeight}%`,
            left: ghostLeft,
            width: `calc(${dayWidth})`,
            borderColor,
            backgroundColor: bgColor,
          }}
        />
        {/* Preview at new position */}
        <div
          className="pointer-events-none absolute z-20 rounded border-2"
          style={{
            top: `${previewTop}%`,
            height: `${previewHeight}%`,
            left: previewLeft,
            width: `calc(${dayWidth})`,
            borderColor,
            backgroundColor: bgColor,
          }}
        />
      </>
    )
  }

  // Create / resize preview
  const isResize = dragState.mode === "resize"
  const previewStart = isResize
    ? Math.min(dragState.anchorMin, dragState.currentMin)
    : Math.min(dragState.startMin, dragState.currentMin)
  const previewEnd = isResize
    ? Math.max(dragState.anchorMin, dragState.currentMin)
    : Math.max(dragState.startMin, dragState.currentMin)
  const top = ((previewStart - gridStartMin) / gridHeight) * 100
  const height = ((previewEnd - previewStart) / gridHeight) * 100
  const leftOffset = `calc(${GRID.TIME_COL} + (${dayWidth}) * ${dragState.dayIndex})`

  return (
    <div
      className="pointer-events-none absolute z-20 rounded border-2"
      style={{
        top: `${top}%`,
        height: `${height}%`,
        left: leftOffset,
        width: `calc(${dayWidth})`,
        borderColor,
        backgroundColor: bgColor,
      }}
    />
  )
}
