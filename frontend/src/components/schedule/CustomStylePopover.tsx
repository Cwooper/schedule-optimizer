import { useState } from "react"
import { RotateCcw } from "lucide-react"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { cn } from "@/lib/utils"
import { opacityToHex, BLOCKED_PRESET_COLORS } from "@/lib/schedule-utils"
import type { BlockedTimeGroup } from "@/stores/app-store"

const DEFAULT_OPACITY = 20

export function CustomStylePopover({
  group,
  onApply,
}: {
  group: BlockedTimeGroup
  onApply: (updates: Partial<BlockedTimeGroup>) => void
}) {
  const [popoverOpen, setPopoverOpen] = useState(false)
  const [hex, setHex] = useState("")
  const [opacity, setOpacity] = useState(DEFAULT_OPACITY)

  const isCustomColor = group.color !== null && !BLOCKED_PRESET_COLORS.some((c) => c.hex === group.color)
  const hasCustomOpacity = group.color !== null && group.opacity !== DEFAULT_OPACITY

  // Sync local state when popover opens
  const handleOpenChange = (open: boolean) => {
    if (open) {
      setHex(group.color ?? "")
      setOpacity(group.opacity)
    }
    setPopoverOpen(open)
  }

  // Normalize: accept hex with or without leading #
  const normalizedHex = hex.startsWith("#") ? hex : `#${hex}`
  const validHex = /^#[0-9a-fA-F]{6}$/.test(normalizedHex)
  const previewColor = validHex ? normalizedHex : group.color
  const previewHex = previewColor ? `${previewColor}${opacityToHex(opacity)}` : null

  const handleApply = () => {
    const updates: Partial<BlockedTimeGroup> = { opacity }
    if (validHex) updates.color = normalizedHex
    onApply(updates)
    setPopoverOpen(false)
  }

  return (
    <Popover open={popoverOpen} onOpenChange={handleOpenChange}>
      <PopoverTrigger asChild>
        <button
          type="button"
          title="Custom style"
          className={cn(
            "flex size-6 items-center justify-center rounded-sm border-2 text-[10px] font-medium transition-colors",
            isCustomColor || hasCustomOpacity
              ? "border-foreground"
              : "border-border hover:border-foreground/50"
          )}
          style={isCustomColor ? { backgroundColor: group.color! } : {}}
        >
          {!isCustomColor && "#"}
        </button>
      </PopoverTrigger>
      <PopoverContent
        className="w-56 space-y-3 p-3"
        align="end"
        side="bottom"
        onOpenAutoFocus={(e) => e.preventDefault()}
      >
        {/* Preview swatch */}
        <div className="flex items-center gap-2">
          <span
            className="size-8 shrink-0 rounded border border-border"
            style={previewHex ? { backgroundColor: previewHex } : {
              background:
                "repeating-linear-gradient(45deg, transparent, transparent 2px, rgba(128,128,128,0.3) 2px, rgba(128,128,128,0.3) 4px)",
            }}
          />
          <div className="min-w-0 flex-1 text-xs">
            <div className="font-medium">Preview</div>
            <div className="text-muted-foreground truncate">
              {previewColor ?? "No color"} · {opacity}%
            </div>
          </div>
        </div>

        {/* Hex input */}
        <div className="space-y-1">
          <Label className="text-[11px]">Hex Color</Label>
          <Input
            value={hex}
            onChange={(e) => setHex(e.target.value)}
            placeholder="#6366f1"
            className="h-7 font-mono text-xs"
            onKeyDown={(e) => {
              if (e.key === "Enter" && validHex) handleApply()
            }}
          />
        </div>

        {/* Opacity slider */}
        <div className="space-y-1">
          <div className="flex items-center justify-between">
            <Label className="text-[11px]">Opacity</Label>
            {opacity !== DEFAULT_OPACITY && (
              <button
                type="button"
                className="text-muted-foreground hover:text-foreground flex items-center gap-0.5 text-[10px] transition-colors"
                onClick={() => setOpacity(DEFAULT_OPACITY)}
                title="Reset to default"
              >
                <RotateCcw className="size-2.5" />
                Reset
              </button>
            )}
          </div>
          <div className="flex items-center gap-2">
            <input
              type="range"
              min={10}
              max={80}
              step={5}
              value={opacity}
              onChange={(e) => setOpacity(Number(e.target.value))}
              className="h-1 flex-1 accent-foreground"
            />
            <span className="text-muted-foreground tabular-nums text-[11px] w-7 text-right">
              {opacity}%
            </span>
          </div>
        </div>

        {/* Apply button */}
        <Button
          size="sm"
          className="w-full text-xs"
          onClick={handleApply}
        >
          Apply
        </Button>
      </PopoverContent>
    </Popover>
  )
}
