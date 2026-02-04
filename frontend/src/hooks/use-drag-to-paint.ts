import { useCallback, useEffect, useRef, useState } from "react"
import type { BlockedTimeBlock } from "@/stores/app-store"
import { GRID } from "@/lib/schedule-utils"
import { genId } from "@/lib/utils"

export interface DragState {
  dayIndex: number
  startMin: number
  currentMin: number
}

interface UseDragToPaintOptions {
  enabled: boolean
  gridStartMin: number
  gridEndMin: number
  onAddBlock: ((block: BlockedTimeBlock) => void) | undefined
}

export function useDragToPaint({
  enabled,
  gridStartMin,
  gridEndMin,
  onAddBlock,
}: UseDragToPaintOptions) {
  const gridHeight = gridEndMin - gridStartMin
  const gridBodyRef = useRef<HTMLDivElement>(null)
  const gridContentRef = useRef<HTMLDivElement>(null)

  const [dragState, setDragState] = useState<DragState | null>(null)
  const dragRef = useRef(dragState)
  useEffect(() => {
    dragRef.current = dragState
  }, [dragState])

  const pointerToGridPos = useCallback(
    (clientX: number, clientY: number) => {
      const el = gridBodyRef.current
      const contentEl = gridContentRef.current
      if (!el || !contentEl) return null
      const rect = el.getBoundingClientRect()
      const x = clientX - rect.left - GRID.TIME_COL_PX
      const y = clientY - rect.top + el.scrollTop
      const dayColWidth = (rect.width - GRID.TIME_COL_PX) / GRID.DAY_COUNT
      const dayIndex = Math.floor(x / dayColWidth)
      if (dayIndex < 0 || dayIndex > GRID.DAY_COUNT - 1) return null
      const contentHeight = contentEl.offsetHeight
      const minuteOffset = (y / contentHeight) * gridHeight
      const minute = gridStartMin + minuteOffset
      const snapped =
        Math.round(minute / GRID.SNAP_MINUTES) * GRID.SNAP_MINUTES
      return {
        dayIndex,
        minute: Math.max(gridStartMin, Math.min(gridEndMin, snapped)),
      }
    },
    [gridStartMin, gridEndMin, gridHeight]
  )

  const handlePointerDown = useCallback(
    (e: React.PointerEvent) => {
      if (!enabled) return
      const pos = pointerToGridPos(e.clientX, e.clientY)
      if (!pos) return
      const target = e.target as HTMLElement
      target.setPointerCapture(e.pointerId)
      setDragState({
        dayIndex: pos.dayIndex,
        startMin: pos.minute,
        currentMin: pos.minute,
      })
    },
    [enabled, pointerToGridPos]
  )

  const handlePointerMove = useCallback(
    (e: React.PointerEvent) => {
      const drag = dragRef.current
      if (!drag) return
      const pos = pointerToGridPos(e.clientX, e.clientY)
      if (!pos || pos.dayIndex !== drag.dayIndex) return
      setDragState((prev) => prev && { ...prev, currentMin: pos.minute })
    },
    [pointerToGridPos]
  )

  const handlePointerUp = useCallback(() => {
    const drag = dragRef.current
    if (!drag || !onAddBlock) return
    const minMin = Math.min(drag.startMin, drag.currentMin)
    const maxMin = Math.max(drag.startMin, drag.currentMin)
    if (maxMin - minMin >= GRID.MIN_DRAG_MINUTES) {
      const startH = String(Math.floor(minMin / 60)).padStart(2, "0")
      const startM = String(minMin % 60).padStart(2, "0")
      const endH = String(Math.floor(maxMin / 60)).padStart(2, "0")
      const endM = String(maxMin % 60).padStart(2, "0")
      onAddBlock({
        id: genId(),
        day: drag.dayIndex,
        startTime: `${startH}${startM}`,
        endTime: `${endH}${endM}`,
      })
    }
    setDragState(null)
  }, [onAddBlock])

  // Cancel drag on escape
  useEffect(() => {
    if (!dragState) return
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") setDragState(null)
    }
    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [dragState])

  return {
    dragState,
    gridBodyRef,
    gridContentRef,
    pointerHandlers: {
      onPointerDown: handlePointerDown,
      onPointerMove: handlePointerMove,
      onPointerUp: handlePointerUp,
    },
  }
}
