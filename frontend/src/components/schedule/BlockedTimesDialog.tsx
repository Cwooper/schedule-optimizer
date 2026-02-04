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
import { DAYS } from "@/lib/schedule-utils"
import { useAppStore, type BlockedTimeGroup } from "@/stores/app-store"

const PRESET_COLORS = [
  { key: "slate", hex: "#64748b", label: "Slate" },
  { key: "rose", hex: "#f43f5e", label: "Rose" },
  { key: "amber", hex: "#f59e0b", label: "Amber" },
  { key: "emerald", hex: "#10b981", label: "Emerald" },
  { key: "sky", hex: "#0ea5e9", label: "Sky" },
  { key: "violet", hex: "#8b5cf6", label: "Violet" },
  { key: "orange", hex: "#f97316", label: "Orange" },
  { key: "pink", hex: "#ec4899", label: "Pink" },
]

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
  const [customColorId, setCustomColorId] = useState<string | null>(null)
  const [customHex, setCustomHex] = useState("")

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
    setCustomColorId(null)
  }

  const handleColorSelect = (groupId: string, color: string | null) => {
    updateBlockedTimeGroup(groupId, { color })
    setCustomColorId(null)
  }

  const handleCustomColorSubmit = (groupId: string) => {
    if (/^#[0-9a-fA-F]{6}$/.test(customHex)) {
      updateBlockedTimeGroup(groupId, { color: customHex })
      setCustomColorId(null)
      setCustomHex("")
    }
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
                onCustomColorOpen={() => {
                  setCustomColorId(group.id)
                  setCustomHex(group.color ?? "")
                }}
                customColorOpen={customColorId === group.id}
                customHex={customHex}
                onCustomHexChange={setCustomHex}
                onCustomColorSubmit={() =>
                  handleCustomColorSubmit(group.id)
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
  onCustomColorOpen: () => void
  customColorOpen: boolean
  customHex: string
  onCustomHexChange: (hex: string) => void
  onCustomColorSubmit: () => void
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
  onCustomColorOpen,
  customColorOpen,
  customHex,
  onCustomHexChange,
  onCustomColorSubmit,
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
        <ColorIndicator color={group.color} />
        <div className="min-w-0 flex-1">
          <div className="truncate font-medium">{displayName}</div>
          <div className="text-muted-foreground truncate text-xs">
            {summarizeBlocks(group)}
          </div>
        </div>
        <div
          onClick={(e) => e.stopPropagation()}
          onKeyDown={(e) => {
            if (e.key === " ") e.stopPropagation()
          }}
        >
          <Switch
            checked={group.enabled}
            onCheckedChange={onToggleEnabled}
          />
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

          {/* Color picker */}
          <div className="space-y-1">
            <Label className="text-xs">Color</Label>
            <div className="flex flex-wrap gap-1.5">
              {/* Hatched option */}
              <button
                type="button"
                title="Hatched pattern"
                className={cn(
                  "size-6 rounded-sm border-2 transition-colors",
                  group.color === null
                    ? "border-foreground"
                    : "border-border hover:border-foreground/50"
                )}
                style={{
                  background:
                    "repeating-linear-gradient(45deg, transparent, transparent 2px, rgba(128,128,128,0.3) 2px, rgba(128,128,128,0.3) 4px)",
                }}
                onClick={() => onColorSelect(null)}
              />
              {/* Preset colors */}
              {PRESET_COLORS.map((c) => (
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
              {/* Custom color toggle */}
              <button
                type="button"
                title="Custom color"
                className={cn(
                  "flex size-6 items-center justify-center rounded-sm border-2 text-[10px] font-medium transition-colors",
                  customColorOpen || (group.color && !PRESET_COLORS.some((c) => c.hex === group.color))
                    ? "border-foreground"
                    : "border-border hover:border-foreground/50"
                )}
                style={
                  group.color && !PRESET_COLORS.some((c) => c.hex === group.color)
                    ? { backgroundColor: group.color }
                    : {}
                }
                onClick={onCustomColorOpen}
              >
                {!(group.color && !PRESET_COLORS.some((c) => c.hex === group.color)) && "#"}
              </button>
            </div>
            {/* Custom hex input */}
            {customColorOpen && (
              <div className="flex items-center gap-2 pt-1">
                <Input
                  value={customHex}
                  onChange={(e) => onCustomHexChange(e.target.value)}
                  placeholder="#6366f1"
                  className="h-7 flex-1 font-mono text-xs"
                  onKeyDown={(e) => {
                    if (e.key === "Enter") onCustomColorSubmit()
                  }}
                />
                <Button
                  size="sm"
                  variant="outline"
                  className="h-7 px-2 text-xs"
                  onClick={onCustomColorSubmit}
                  disabled={!/^#[0-9a-fA-F]{6}$/.test(customHex)}
                >
                  Apply
                </Button>
              </div>
            )}
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
          <div className="flex gap-2 pt-1">
            <Button
              variant="outline"
              size="sm"
              className="flex-1 text-xs"
              onClick={onPaint}
            >
              <Paintbrush className="size-3.5" />
              Paint on Grid
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="text-destructive hover:bg-destructive/10 text-xs"
              onClick={onDelete}
            >
              <Trash2 className="size-3.5" />
              Delete
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}

function ColorIndicator({ color }: { color: string | null }) {
  if (color === null) {
    return (
      <span
        className="size-3.5 shrink-0 rounded-sm border border-border"
        style={{
          background:
            "repeating-linear-gradient(45deg, transparent, transparent 1.5px, rgba(128,128,128,0.4) 1.5px, rgba(128,128,128,0.4) 3px)",
        }}
      />
    )
  }
  return (
    <span
      className="size-3.5 shrink-0 rounded-sm border border-border"
      style={{ backgroundColor: color }}
    />
  )
}
