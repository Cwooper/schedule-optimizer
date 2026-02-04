import { useState } from "react"
import {
  ChevronDown,
  ChevronRight,
  Paintbrush,
  Plus,
  Trash2,
  X,
} from "lucide-react"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Textarea } from "@/components/ui/textarea"
import { cn } from "@/lib/utils"
import { DAYS, BLOCKED_PRESET_COLORS, parseTime, minsToTimeStr, GRID, otherBlockRanges, clampToAvoidOverlap } from "@/lib/schedule-utils"
import { useAppStore, type BlockedTimeBlock, type BlockedTimeGroup } from "@/stores/app-store"
import { CustomStylePopover } from "./CustomStylePopover"

function formatTime(time: string): string {
  const clean = time.replace(":", "").padStart(4, "0")
  const h = parseInt(clean.slice(0, 2), 10)
  const m = parseInt(clean.slice(2, 4), 10)
  const suffix = h >= 12 ? "pm" : "am"
  let hour = h
  if (h === 0) hour = 12
  else if (h > 12) hour = h - 12
  return m === 0 ? `${hour}${suffix}` : `${hour}:${String(m).padStart(2, "0")}${suffix}`
}

/** Convert stored time "0900" or "1430" to input[type=time] format "09:00" or "14:30" */
function toInputTime(time: string): string {
  const clean = time.replace(":", "").padStart(4, "0")
  return `${clean.slice(0, 2)}:${clean.slice(2, 4)}`
}

/** Convert input[type=time] format "09:00" to stored format "0900" */
function fromInputTime(time: string): string {
  return time.replace(":", "")
}


/** Clamp a proposed block edit to avoid overlapping other blocks in the group. Returns null if too small. */
function clampBlockEdit(
  proposed: BlockedTimeBlock,
  allBlocks: BlockedTimeBlock[]
): BlockedTimeBlock | null {
  const existing = otherBlockRanges(allBlocks, proposed.id)
  const startMin = parseTime(proposed.startTime)
  const endMin = parseTime(proposed.endTime)
  const clamped = clampToAvoidOverlap(
    { day: proposed.day, startMin, endMin },
    existing,
    "clamp"
  )
  if (!clamped || clamped.endMin - clamped.startMin < GRID.MIN_DRAG_MINUTES) return null
  return {
    ...proposed,
    day: clamped.day,
    startTime: minsToTimeStr(clamped.startMin),
    endTime: minsToTimeStr(clamped.endMin),
  }
}

function summarizeBlocks(group: BlockedTimeGroup): string {
  if (group.blocks.length === 0) return "No time blocks"
  const days = [...new Set(group.blocks.map((b) => b.day))].sort()
  const dayStr = days.map((d) => DAYS[d]).join("")
  if (group.blocks.length === 1) {
    const b = group.blocks[0]
    return `${DAYS[b.day]} ${formatTime(b.startTime)}–${formatTime(b.endTime)}`
  }
  return `${dayStr} · ${group.blocks.length} blocks`
}

interface BlockedTimesDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onStartPainting: (groupId: string) => void
  initialExpandedId?: string | null
}

export function BlockedTimesDialog({
  open,
  onOpenChange,
  onStartPainting,
  initialExpandedId,
}: BlockedTimesDialogProps) {
  const {
    blockedTimeGroups,
    addBlockedTimeGroup,
    removeBlockedTimeGroup,
    updateBlockedTimeGroup,
    updateBlockInGroupById,
    removeBlockFromGroupById,
  } = useAppStore()

  const [expandedId, setExpandedId] = useState<string | null>(null)

  // Expand the target group when dialog opens via grid click
  const [prevOpen, setPrevOpen] = useState(false)
  if (open !== prevOpen) {
    setPrevOpen(open)
    if (open && initialExpandedId) {
      setExpandedId(initialExpandedId)
    }
  }

  const handleAddAndPaint = () => {
    const id = addBlockedTimeGroup()
    onOpenChange(false)
    onStartPainting(id)
  }

  const handlePaintExisting = (groupId: string) => {
    onOpenChange(false)
    onStartPainting(groupId)
  }

  const toggleExpanded = (id: string) => {
    setExpandedId((prev) => (prev === id ? null : id))
  }

  const handleColorSelect = (groupId: string, color: string | null) => {
    const group = blockedTimeGroups.find((g) => g.id === groupId)
    if (!group) return

    if (color === null) {
      updateBlockedTimeGroup(groupId, { color: null, hatched: true })
    } else if (group.color === color) {
      updateBlockedTimeGroup(groupId, { color: null, hatched: true })
    } else {
      updateBlockedTimeGroup(groupId, { color })
    }
  }

  const handleHatchedToggle = (groupId: string, hatched: boolean) => {
    const group = blockedTimeGroups.find((g) => g.id === groupId)
    if (!group) return
    if (!hatched && group.color === null) return
    updateBlockedTimeGroup(groupId, { hatched })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Blocked Times</DialogTitle>
          <DialogDescription>
            Block out times when you&apos;re unavailable. These will be excluded
            from generated schedules.
          </DialogDescription>
        </DialogHeader>

        <div className="max-h-80 space-y-1 overflow-y-auto">
          {blockedTimeGroups.length === 0 ? (
            <p className="text-muted-foreground py-6 text-center text-sm">
              No blocked times yet. Add one to get started.
            </p>
          ) : (
            blockedTimeGroups.map((group, idx) => (
              <GroupRow
                key={group.id}
                group={group}
                defaultName={`Blocked Time ${idx + 1}`}
                expanded={expandedId === group.id}
                onToggleExpand={() => toggleExpanded(group.id)}
                onToggleEnabled={(enabled) =>
                  updateBlockedTimeGroup(group.id, { enabled })
                }
                onTitleChange={(title) =>
                  updateBlockedTimeGroup(group.id, { title })
                }
                onDescriptionChange={(description) =>
                  updateBlockedTimeGroup(group.id, { description })
                }
                onColorSelect={(color) =>
                  handleColorSelect(group.id, color)
                }
                onHatchedToggle={(hatched) =>
                  handleHatchedToggle(group.id, hatched)
                }
                onCustomApply={(updates) =>
                  updateBlockedTimeGroup(group.id, updates)
                }
                onUpdateBlock={(blockId, block) =>
                  updateBlockInGroupById(group.id, blockId, block)
                }
                onRemoveBlock={(blockId) =>
                  removeBlockFromGroupById(group.id, blockId)
                }
                onDelete={() => {
                  removeBlockedTimeGroup(group.id)
                  if (expandedId === group.id) setExpandedId(null)
                }}
                onPaint={() => handlePaintExisting(group.id)}
              />
            ))
          )}
        </div>

        <Button
          variant="outline"
          className="w-full"
          onClick={handleAddAndPaint}
        >
          <Plus className="size-4" />
          Add Blocked Time
        </Button>
      </DialogContent>
    </Dialog>
  )
}

interface GroupRowProps {
  group: BlockedTimeGroup
  defaultName: string
  expanded: boolean
  onToggleExpand: () => void
  onToggleEnabled: (enabled: boolean) => void
  onTitleChange: (title: string) => void
  onDescriptionChange: (description: string) => void
  onColorSelect: (color: string | null) => void
  onHatchedToggle: (hatched: boolean) => void
  onCustomApply: (updates: Partial<BlockedTimeGroup>) => void
  onUpdateBlock: (blockId: string, block: BlockedTimeBlock) => void
  onRemoveBlock: (blockId: string) => void
  onDelete: () => void
  onPaint: () => void
}

function GroupRow({
  group,
  defaultName,
  expanded,
  onToggleExpand,
  onToggleEnabled,
  onTitleChange,
  onDescriptionChange,
  onColorSelect,
  onHatchedToggle,
  onCustomApply,
  onUpdateBlock,
  onRemoveBlock,
  onDelete,
  onPaint,
}: GroupRowProps) {
  const [editingBlockId, setEditingBlockId] = useState<string | null>(null)
  const displayName = group.title || defaultName
  return (
    <div className="rounded-md border">
      {/* Collapsed header */}
      <div
        role="button"
        tabIndex={0}
        className="flex w-full cursor-pointer items-center gap-2 px-3 py-2 text-left text-sm hover:bg-muted/50"
        onClick={onToggleExpand}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault()
            onToggleExpand()
          }
        }}
      >
        {expanded ? (
          <ChevronDown className="text-muted-foreground size-4 shrink-0" />
        ) : (
          <ChevronRight className="text-muted-foreground size-4 shrink-0" />
        )}
        <ColorIndicator color={group.color} hatched={group.hatched} />
        <div className="min-w-0 flex-1">
          <div className="truncate font-medium">{displayName}</div>
          <div className="text-muted-foreground truncate text-xs">
            {summarizeBlocks(group)}
          </div>
        </div>
        <div
          className="flex items-center gap-1.5"
          onClick={(e) => e.stopPropagation()}
          onKeyDown={(e) => {
            if (e.key === " ") e.stopPropagation()
          }}
        >
          <Switch
            checked={group.enabled}
            onCheckedChange={onToggleEnabled}
          />
          <button
            type="button"
            className="text-muted-foreground hover:text-primary rounded p-0.5 transition-colors"
            onClick={onPaint}
            aria-label="Paint blocked time on grid"
          >
            <Paintbrush className="size-4" />
          </button>
          <button
            type="button"
            className="text-muted-foreground hover:text-destructive rounded p-0.5 transition-colors"
            onClick={onDelete}
            aria-label="Delete blocked time group"
          >
            <Trash2 className="size-4" />
          </button>
        </div>
      </div>

      {/* Expanded details */}
      {expanded && (
        <div className="space-y-3 border-t px-3 py-3">
          {/* Title */}
          <div className="space-y-1">
            <Label className="text-xs">Title</Label>
            <Input
              value={group.title}
              onChange={(e) => onTitleChange(e.target.value)}
              placeholder={defaultName}
              className="h-8 text-sm"
            />
          </div>

          {/* Description */}
          <div className="space-y-1">
            <Label className="text-xs">Description</Label>
            <Textarea
              value={group.description}
              onChange={(e) => onDescriptionChange(e.target.value)}
              placeholder="Optional description..."
              className="min-h-16 resize-none text-sm"
            />
          </div>

          {/* Style controls */}
          <div className="space-y-2">
            <Label className="text-xs">Style</Label>
            <div className="flex flex-wrap items-center gap-1.5">
              {/* Hatched toggle */}
              <button
                type="button"
                title="Hatched pattern"
                className={cn(
                  "size-6 rounded-sm border-2 transition-colors",
                  group.hatched
                    ? "border-foreground"
                    : "border-border hover:border-foreground/50",
                  !group.hatched && group.color === null && "opacity-50 cursor-not-allowed"
                )}
                style={{
                  background:
                    "repeating-linear-gradient(45deg, transparent, transparent 2px, rgba(128,128,128,0.3) 2px, rgba(128,128,128,0.3) 4px)",
                }}
                onClick={() => onHatchedToggle(!group.hatched)}
              />
              {/* Separator */}
              <div className="mx-0.5 h-5 w-px bg-border" />
              {/* Preset colors */}
              {BLOCKED_PRESET_COLORS.map((c) => (
                <button
                  key={c.key}
                  type="button"
                  title={c.label}
                  className={cn(
                    "size-6 rounded-sm border-2 transition-colors",
                    group.color === c.hex
                      ? "border-foreground"
                      : "border-border hover:border-foreground/50"
                  )}
                  style={{ backgroundColor: c.hex }}
                  onClick={() => onColorSelect(c.hex)}
                />
              ))}
              {/* Custom popover */}
              <CustomStylePopover group={group} onApply={onCustomApply} />
            </div>
          </div>

          {/* Block list */}
          {group.blocks.length > 0 && (
            <div className="space-y-1">
              <Label className="text-xs">Time Blocks</Label>
              {group.blocks.map((block) => {
                const isEditing = editingBlockId === block.id
                return (
                  <div key={block.id} className="rounded border border-transparent hover:border-border">
                    {/* Summary row — click to expand */}
                    <div
                      role="button"
                      tabIndex={0}
                      className="flex cursor-pointer items-center justify-between rounded px-2 py-1 text-xs hover:bg-muted/50"
                      onClick={() => setEditingBlockId(isEditing ? null : block.id)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter" || e.key === " ") {
                          e.preventDefault()
                          setEditingBlockId(isEditing ? null : block.id)
                        }
                      }}
                    >
                      <span>
                        {DAYS[block.day]} {formatTime(block.startTime)}
                        {"–"}
                        {formatTime(block.endTime)}
                      </span>
                      <button
                        type="button"
                        className="text-muted-foreground hover:text-destructive"
                        onClick={(e) => {
                          e.stopPropagation()
                          if (isEditing) setEditingBlockId(null)
                          onRemoveBlock(block.id)
                        }}
                      >
                        <X className="size-3.5" />
                      </button>
                    </div>
                    {/* Inline edit form */}
                    {isEditing && (
                      <BlockTimeEditor
                        block={block}
                        allBlocks={group.blocks}
                        onUpdateBlock={onUpdateBlock}
                      />
                    )}
                  </div>
                )
              })}
            </div>
          )}

        </div>
      )}
    </div>
  )
}

function BlockTimeEditor({
  block,
  allBlocks,
  onUpdateBlock,
}: {
  block: BlockedTimeBlock
  allBlocks: BlockedTimeBlock[]
  onUpdateBlock: (blockId: string, block: BlockedTimeBlock) => void
}) {
  const [draftStart, setDraftStart] = useState(toInputTime(block.startTime))
  const [draftEnd, setDraftEnd] = useState(toInputTime(block.endTime))

  const commitTimes = (start: string, end: string) => {
    const newStart = fromInputTime(start)
    const newEnd = fromInputTime(end)
    const startMins = parseTime(newStart)
    const endMins = parseTime(newEnd)
    if (endMins - startMins < GRID.MIN_DRAG_MINUTES) return
    const proposed = { ...block, startTime: newStart, endTime: newEnd }
    const clamped = clampBlockEdit(proposed, allBlocks)
    if (clamped) {
      onUpdateBlock(block.id, clamped)
      setDraftStart(toInputTime(clamped.startTime))
      setDraftEnd(toInputTime(clamped.endTime))
    } else {
      setDraftStart(toInputTime(block.startTime))
      setDraftEnd(toInputTime(block.endTime))
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      e.currentTarget.blur()
    }
  }

  return (
    <div className="flex items-center gap-1.5 px-2 pb-1.5 pt-0.5">
      <select
        value={block.day}
        onChange={(e) => {
          const proposed = { ...block, day: Number(e.target.value) }
          const clamped = clampBlockEdit(proposed, allBlocks)
          if (clamped) onUpdateBlock(block.id, clamped)
        }}
        className="h-7 rounded border border-input bg-background px-1 text-xs"
      >
        {DAYS.map((day, dayIdx) => (
          <option key={day} value={dayIdx}>{day}</option>
        ))}
      </select>
      <input
        type="time"
        value={draftStart}
        onChange={(e) => setDraftStart(e.target.value)}
        onBlur={() => commitTimes(draftStart, draftEnd)}
        onKeyDown={handleKeyDown}
        className="h-7 rounded border border-input bg-background px-1 text-xs tabular-nums"
      />
      <span className="text-muted-foreground text-xs">–</span>
      <input
        type="time"
        value={draftEnd}
        onChange={(e) => setDraftEnd(e.target.value)}
        onBlur={() => commitTimes(draftStart, draftEnd)}
        onKeyDown={handleKeyDown}
        className="h-7 rounded border border-input bg-background px-1 text-xs tabular-nums"
      />
    </div>
  )
}

function ColorIndicator({ color, hatched }: { color: string | null; hatched: boolean }) {
  const hatchGradient = "repeating-linear-gradient(45deg, transparent, transparent 1.5px, rgba(128,128,128,0.4) 1.5px, rgba(128,128,128,0.4) 3px)"

  if (color === null) {
    return (
      <span
        className="size-3.5 shrink-0 rounded-sm border border-border"
        style={{
          background: hatched ? hatchGradient : undefined,
        }}
      />
    )
  }
  return (
    <span
      className="size-3.5 shrink-0 rounded-sm border border-border"
      style={{
        backgroundColor: color,
        backgroundImage: hatched ? hatchGradient : undefined,
      }}
    />
  )
}
