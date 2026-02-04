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
import { DAYS, BLOCKED_PRESET_COLORS } from "@/lib/schedule-utils"
import { useAppStore, type BlockedTimeGroup } from "@/stores/app-store"
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
    removeBlockFromGroup,
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
                onRemoveBlock={(blockIdx) =>
                  removeBlockFromGroup(group.id, blockIdx)
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
  onRemoveBlock: (blockIdx: number) => void
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
  onRemoveBlock,
  onDelete,
  onPaint,
}: GroupRowProps) {
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
              {group.blocks.map((block, idx) => (
                <div
                  key={idx}
                  className="flex items-center justify-between rounded px-2 py-1 text-xs hover:bg-muted/50"
                >
                  <span>
                    {DAYS[block.day]} {formatTime(block.startTime)}
                    {"–"}
                    {formatTime(block.endTime)}
                  </span>
                  <button
                    type="button"
                    className="text-muted-foreground hover:text-destructive"
                    onClick={() => onRemoveBlock(idx)}
                  >
                    <X className="size-3.5" />
                  </button>
                </div>
              ))}
            </div>
          )}

          {/* Actions */}
          <div className="pt-1">
            <Button
              variant="outline"
              size="sm"
              className="w-full text-xs"
              onClick={onPaint}
            >
              <Paintbrush className="size-3.5" />
              Paint on Grid
            </Button>
          </div>
        </div>
      )}
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
