import { useCallback, useEffect, useRef, useState } from "react"
import type { BlockedTimeBlock } from "@/stores/app-store"
import { GRID, parseTime, minsToTimeStr, clampToAvoidOverlap, blocksToRanges, otherBlockRanges, type TimeRange } from "@/lib/schedule-utils"
import { genId } from "@/lib/utils"

// ── Types ──────────────────────────────────────────────────────────────

interface DragBase {
  dayIndex: number
  startMin: number
  currentMin: number
}

interface CreateDrag extends DragBase {
  mode: "create"
}

interface ResizeDrag extends DragBase {
  mode: "resize"
  resizeEdge: "top" | "bottom"
  blockId: string
  anchorMin: number
}

interface MoveDrag extends DragBase {
  mode: "move"
  blockId: string
  originalDay: number
  blockDuration: number
  offsetFromTop: number
  currentDay: number
}

export type DragState = CreateDrag | ResizeDrag | MoveDrag

export type HoveredEdge = {
  blockId: string
  edge: "top" | "bottom"
} | null

interface UseDragToPaintOptions {
  enabled: boolean
  gridStartMin: number
  gridEndMin: number
  groupBlocks?: BlockedTimeBlock[]
  onAddBlock: ((block: BlockedTimeBlock) => void) | undefined
  onUpdateBlock?: (blockId: string, block: BlockedTimeBlock) => void
  onRemoveBlock?: (blockId: string) => void
}

// ── Helpers ────────────────────────────────────────────────────────────

/**
 * Compute the snapped/clamped start position for a move drag.
 * Used by both the live preview and the final commit to ensure they match.
 */
export function computeMovePosition(
  currentMin: number,
  offsetFromTop: number,
  duration: number,
  gridStartMin: number,
  gridEndMin: number
): number {
  const rawStart = currentMin - offsetFromTop
  const snapped = Math.round(rawStart / GRID.SNAP_MINUTES) * GRID.SNAP_MINUTES
  return Math.max(gridStartMin, Math.min(gridEndMin - duration, snapped))
}

/** Compute clamped minute for resize drag movement */
function computeResizeMin(drag: ResizeDrag, pointerMin: number, blocks: BlockedTimeBlock[]): number {
  const existing = otherBlockRanges(blocks, drag.blockId)
  const proposedStart = Math.min(drag.anchorMin, pointerMin)
  const proposedEnd = Math.max(drag.anchorMin, pointerMin)
  const clamped = clampToAvoidOverlap(
    { day: drag.dayIndex, startMin: proposedStart, endMin: proposedEnd },
    existing, "clamp", undefined, drag.anchorMin
  )
  if (clamped) {
    return drag.resizeEdge === "bottom" ? clamped.endMin : clamped.startMin
  }
  return drag.anchorMin
}

/** Compute clamped minute for create drag movement */
function computeCreateMin(drag: CreateDrag, pointerMin: number, blocks: BlockedTimeBlock[]): number {
  const existing = blocksToRanges(blocks)
  const rawStart = Math.min(drag.startMin, pointerMin)
  const rawEnd = Math.max(drag.startMin, pointerMin)
  const clamped = clampToAvoidOverlap(
    { day: drag.dayIndex, startMin: rawStart, endMin: rawEnd },
    existing, "clamp", undefined, drag.startMin
  )
  if (clamped) {
    return pointerMin >= drag.startMin ? clamped.endMin : clamped.startMin
  }
  return drag.startMin
}

/** Commit a move drag — returns the updated block or null if invalid */
function commitMove(
  drag: MoveDrag, block: BlockedTimeBlock, allBlocks: BlockedTimeBlock[],
  gridStartMin: number, gridEndMin: number
): BlockedTimeBlock | null {
  const clampedStart = computeMovePosition(
    drag.currentMin, drag.offsetFromTop, drag.blockDuration, gridStartMin, gridEndMin
  )
  const existing = otherBlockRanges(allBlocks, drag.blockId)
  const proposed: TimeRange = {
    day: drag.currentDay, startMin: clampedStart, endMin: clampedStart + drag.blockDuration,
  }
  const snapped = clampToAvoidOverlap(proposed, existing, "snap", drag.blockDuration)
  if (!snapped) return null
  return {
    ...block,
    day: snapped.day,
    startTime: minsToTimeStr(snapped.startMin),
    endTime: minsToTimeStr(snapped.endMin),
  }
}

/** Commit a create drag — returns the new block or null if too small */
function commitCreate(drag: CreateDrag, allBlocks: BlockedTimeBlock[]): BlockedTimeBlock | null {
  const existing = blocksToRanges(allBlocks)
  const minMin = Math.min(drag.startMin, drag.currentMin)
  const maxMin = Math.max(drag.startMin, drag.currentMin)
  const clamped = clampToAvoidOverlap(
    { day: drag.dayIndex, startMin: minMin, endMin: maxMin }, existing, "clamp"
  )
  if (!clamped || clamped.endMin - clamped.startMin < GRID.MIN_DRAG_MINUTES) return null
  return {
    id: genId(),
    day: drag.dayIndex,
    startTime: minsToTimeStr(clamped.startMin),
    endTime: minsToTimeStr(clamped.endMin),
  }
}

/** Commit a resize drag — returns "remove" if too small, updated block, or null */
function commitResize(
  drag: ResizeDrag, block: BlockedTimeBlock
): { action: "update"; block: BlockedTimeBlock } | { action: "remove" } {
  const newStart = Math.min(drag.anchorMin, drag.currentMin)
  const newEnd = Math.max(drag.anchorMin, drag.currentMin)
  if (newEnd - newStart < GRID.MIN_DRAG_MINUTES) {
    return { action: "remove" }
  }
  return {
    action: "update",
    block: { ...block, startTime: minsToTimeStr(newStart), endTime: minsToTimeStr(newEnd) },
  }
}

/** Edge detection threshold in pixels */
const EDGE_THRESHOLD_PX = 8

// ── Hook ───────────────────────────────────────────────────────────────

export function useDragToPaint({
  enabled,
  gridStartMin,
  gridEndMin,
  groupBlocks,
  onAddBlock,
  onUpdateBlock,
  onRemoveBlock,
}: UseDragToPaintOptions) {
  const gridHeight = gridEndMin - gridStartMin
  const gridBodyRef = useRef<HTMLDivElement>(null)
  const gridContentRef = useRef<HTMLDivElement>(null)

  const [dragState, setDragState] = useState<DragState | null>(null)
  const dragRef = useRef(dragState)
  useEffect(() => {
    dragRef.current = dragState
  }, [dragState])

  const [hoveredEdge, setHoveredEdge] = useState<HoveredEdge>(null)

  // Keep a stable ref for groupBlocks so callbacks don't go stale
  const groupBlocksRef = useRef(groupBlocks)
  useEffect(() => {
    groupBlocksRef.current = groupBlocks
  }, [groupBlocks])

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
        rawY: y,
      }
    },
    [gridStartMin, gridEndMin, gridHeight]
  )

  /** Convert a minute value to a pixel Y position in the grid content */
  const minToY = useCallback(
    (min: number): number | null => {
      const contentEl = gridContentRef.current
      if (!contentEl) return null
      const contentHeight = contentEl.offsetHeight
      return ((min - gridStartMin) / gridHeight) * contentHeight
    },
    [gridStartMin, gridHeight]
  )

  /** Hit-test pointer against block edges. Returns edge info or null. */
  const hitTestEdge = useCallback(
    (clientX: number, clientY: number): HoveredEdge => {
      const pos = pointerToGridPos(clientX, clientY)
      if (!pos) return null
      const blocks = groupBlocksRef.current
      if (!blocks) return null

      for (const block of blocks) {
        if (block.day !== pos.dayIndex) continue
        const topY = minToY(parseTime(block.startTime))
        const bottomY = minToY(parseTime(block.endTime))
        if (topY === null || bottomY === null) continue

        if (Math.abs(pos.rawY - topY) <= EDGE_THRESHOLD_PX) {
          return { blockId: block.id, edge: "top" }
        }
        if (Math.abs(pos.rawY - bottomY) <= EDGE_THRESHOLD_PX) {
          return { blockId: block.id, edge: "bottom" }
        }
      }
      return null
    },
    [pointerToGridPos, minToY]
  )

  /** Hit-test pointer against block body (interior, not edges). */
  const hitTestBody = useCallback(
    (clientX: number, clientY: number): { blockId: string; offsetFromTop: number } | null => {
      const pos = pointerToGridPos(clientX, clientY)
      if (!pos) return null
      const blocks = groupBlocksRef.current
      if (!blocks) return null

      for (const block of blocks) {
        if (block.day !== pos.dayIndex) continue
        const bStart = parseTime(block.startTime)
        const topY = minToY(bStart)
        const bottomY = minToY(parseTime(block.endTime))
        if (topY === null || bottomY === null) continue

        if (pos.rawY > topY + EDGE_THRESHOLD_PX && pos.rawY < bottomY - EDGE_THRESHOLD_PX) {
          return { blockId: block.id, offsetFromTop: pos.minute - bStart }
        }
      }
      return null
    },
    [pointerToGridPos, minToY]
  )

  /** Capture pointer on the grid body element (stable across child re-renders) */
  const capturePointer = useCallback((pointerId: number) => {
    gridBodyRef.current?.setPointerCapture(pointerId)
  }, [])

  /** Try to start a resize drag from an edge hit */
  const tryStartResize = useCallback(
    (e: React.PointerEvent): boolean => {
      const edgeHit = hitTestEdge(e.clientX, e.clientY)
      if (!edgeHit) return false
      const block = groupBlocksRef.current?.find((b) => b.id === edgeHit.blockId)
      if (!block) return false
      const bStart = parseTime(block.startTime)
      const bEnd = parseTime(block.endTime)
      capturePointer(e.pointerId)
      setDragState({
        mode: "resize",
        dayIndex: block.day,
        startMin: edgeHit.edge === "top" ? bStart : bEnd,
        currentMin: edgeHit.edge === "top" ? bStart : bEnd,
        resizeEdge: edgeHit.edge,
        blockId: block.id,
        anchorMin: edgeHit.edge === "top" ? bEnd : bStart,
      })
      return true
    },
    [hitTestEdge, capturePointer]
  )

  /** Try to start a move drag from a body hit */
  const tryStartMove = useCallback(
    (e: React.PointerEvent, posMinute: number, posDayIndex: number): boolean => {
      const bodyHit = hitTestBody(e.clientX, e.clientY)
      if (!bodyHit) return false
      const block = groupBlocksRef.current?.find((b) => b.id === bodyHit.blockId)
      if (!block) return false
      const bStart = parseTime(block.startTime)
      const bEnd = parseTime(block.endTime)
      capturePointer(e.pointerId)
      setDragState({
        mode: "move",
        dayIndex: block.day,
        startMin: bStart,
        currentMin: posMinute,
        blockId: block.id,
        originalDay: block.day,
        blockDuration: bEnd - bStart,
        offsetFromTop: bodyHit.offsetFromTop,
        currentDay: posDayIndex,
      })
      return true
    },
    [hitTestBody, capturePointer]
  )

  // ── Pointer handlers ───────────────────────────────────────────────

  const handlePointerDown = useCallback(
    (e: React.PointerEvent) => {
      if (!enabled) return
      const pos = pointerToGridPos(e.clientX, e.clientY)
      if (!pos) return

      if (tryStartResize(e)) return
      if (tryStartMove(e, pos.minute, pos.dayIndex)) return

      // Create drag on empty space
      capturePointer(e.pointerId)
      setDragState({
        mode: "create",
        dayIndex: pos.dayIndex,
        startMin: pos.minute,
        currentMin: pos.minute,
      })
    },
    [enabled, pointerToGridPos, tryStartResize, tryStartMove, capturePointer]
  )

  const handlePointerMove = useCallback(
    (e: React.PointerEvent) => {
      const drag = dragRef.current

      if (!drag) {
        if (enabled) setHoveredEdge(hitTestEdge(e.clientX, e.clientY))
        return
      }

      const pos = pointerToGridPos(e.clientX, e.clientY)
      if (!pos) return
      const blocks = groupBlocksRef.current ?? []

      if (drag.mode === "move") {
        setDragState((prev) => prev && { ...prev, currentMin: pos.minute, currentDay: pos.dayIndex })
        return
      }

      if (drag.mode === "resize") {
        const newMin = computeResizeMin(drag, pos.minute, blocks)
        setDragState((prev) => prev && { ...prev, currentMin: newMin })
        return
      }

      // Create mode — lock to original day
      if (pos.dayIndex !== drag.dayIndex) return
      const clampedMin = computeCreateMin(drag, pos.minute, blocks)
      setDragState((prev) => prev && { ...prev, currentMin: clampedMin })
    },
    [enabled, pointerToGridPos, hitTestEdge]
  )

  const finishMoveDrag = useCallback((drag: MoveDrag, allBlocks: BlockedTimeBlock[]) => {
    if (!onUpdateBlock) return
    const block = allBlocks.find((b) => b.id === drag.blockId)
    if (!block) return
    const updated = commitMove(drag, block, allBlocks, gridStartMin, gridEndMin)
    if (updated) onUpdateBlock(drag.blockId, updated)
  }, [onUpdateBlock, gridStartMin, gridEndMin])

  const finishResizeDrag = useCallback((drag: ResizeDrag, allBlocks: BlockedTimeBlock[]) => {
    if (!onUpdateBlock || !onRemoveBlock) return
    const block = allBlocks.find((b) => b.id === drag.blockId)
    if (!block) return
    const result = commitResize(drag, block)
    if (result.action === "remove") onRemoveBlock(drag.blockId)
    else onUpdateBlock(drag.blockId, result.block)
  }, [onUpdateBlock, onRemoveBlock])

  const handlePointerUp = useCallback(() => {
    const drag = dragRef.current
    if (!drag) return

    const allBlocks = groupBlocksRef.current ?? []

    if (drag.mode === "move") finishMoveDrag(drag, allBlocks)
    else if (drag.mode === "resize") finishResizeDrag(drag, allBlocks)
    else if (onAddBlock) {
      const newBlock = commitCreate(drag, allBlocks)
      if (newBlock) onAddBlock(newBlock)
    }

    setDragState(null)
  }, [onAddBlock, finishMoveDrag, finishResizeDrag])

  // Cancel drag on escape
  useEffect(() => {
    if (!dragState) return
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") setDragState(null)
    }
    window.addEventListener("keydown", handleKeyDown)
    return () => window.removeEventListener("keydown", handleKeyDown)
  }, [dragState])

  // Clear hovered edge when not editing — derived from enabled flag
  const effectiveHoveredEdge = enabled ? hoveredEdge : null

  return {
    dragState,
    hoveredEdge: effectiveHoveredEdge,
    gridBodyRef,
    gridContentRef,
    pointerHandlers: {
      onPointerDown: handlePointerDown,
      onPointerMove: handlePointerMove,
      onPointerUp: handlePointerUp,
    },
  }
}
