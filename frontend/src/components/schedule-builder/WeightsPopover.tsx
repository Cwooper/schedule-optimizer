import {
  Armchair,
  ArrowDownUp,
  Award,
  RotateCcw,
  Sunrise,
  Sunset,
  Weight,
  type LucideIcon,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { minsToTimeStr } from "@/lib/schedule-utils"
import type { Importance, Preferences } from "@/lib/api"
import { useAppStore } from "@/stores/app-store"

type Factor = "start" | "end" | "gap" | "gpa" | "seats"

const LEVELS = ["Off", "Low", "Med", "High"]
const STRICTNESS = ["Off", "Loose", "Firm", "Strict"]

const FACTORS: {
  key: Factor
  label: string
  icon: LucideIcon
  levels: string[]
  time?: { key: "idealStart" | "idealEnd"; label: string }
}[] = [
  {
    key: "start",
    label: "Start",
    icon: Sunrise,
    levels: STRICTNESS,
    time: { key: "idealStart", label: "Prefer after" },
  },
  {
    key: "end",
    label: "End",
    icon: Sunset,
    levels: STRICTNESS,
    time: { key: "idealEnd", label: "Prefer done by" },
  },
  { key: "gap", label: "Gaps", icon: ArrowDownUp, levels: LEVELS },
  { key: "gpa", label: "GPA", icon: Award, levels: LEVELS },
  { key: "seats", label: "Seats", icon: Armchair, levels: LEVELS },
]

// 7:00 AM to 7:00 PM in 30-minute steps.
const TIME_OPTIONS = Array.from({ length: 25 }, (_, i) => {
  const mins = 7 * 60 + i * 30
  const hour = Math.floor(mins / 60)
  const hour12 = hour % 12 || 12
  const label = `${hour12}:${String(mins % 60).padStart(2, "0")} ${hour < 12 ? "AM" : "PM"}`
  return { value: minsToTimeStr(mins), label }
})

/** Lives inside the Popover that ScheduleBuilder anchors to its footer. */
export function WeightsTrigger() {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <PopoverTrigger asChild>
          <Button variant="secondary" size="icon">
            <Weight className="size-4" />
            <span className="sr-only">Weights</span>
          </Button>
        </PopoverTrigger>
      </TooltipTrigger>
      <TooltipContent>Weights</TooltipContent>
    </Tooltip>
  )
}

export function WeightsPanel({
  onPointerDownOutside,
}: Pick<React.ComponentProps<typeof PopoverContent>, "onPointerDownOutside">) {
  const preferences = useAppStore((s) => s.preferences)
  const setPreferences = useAppStore((s) => s.setPreferences)
  const resetPreferences = useAppStore((s) => s.resetPreferences)

  return (
    <PopoverContent
      side="top"
      align="start"
      sideOffset={-8}
      className="mx-2 w-[calc(var(--radix-popover-trigger-width)-1rem)]"
      onPointerDownOutside={onPointerDownOutside}
    >
      <div className="mb-3 flex items-center justify-between">
        <span className="text-sm font-medium">Weights</span>
        <Button
          variant="ghost"
          size="sm"
          className="text-muted-foreground -mr-2 h-7 px-2"
          onClick={resetPreferences}
        >
          <RotateCcw className="size-3.5" />
          Reset
        </Button>
      </div>
      <div className="space-y-3">
        {FACTORS.map((factor) => (
          <FactorRows
            key={factor.key}
            factor={factor}
            preferences={preferences}
            onChange={setPreferences}
          />
        ))}
      </div>
    </PopoverContent>
  )
}

function FactorRows({
  factor,
  preferences,
  onChange,
}: {
  factor: (typeof FACTORS)[number]
  preferences: Preferences
  onChange: (updates: Partial<Preferences>) => void
}) {
  const Icon = factor.icon
  const level = preferences[factor.key]
  const time = factor.time
  return (
    <>
      <div className="flex h-8 items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <Icon className="text-muted-foreground size-4" />
          <span className="text-sm">{factor.label}</span>
        </div>
        <ToggleGroup
          type="single"
          variant="outline"
          size="sm"
          spacing={0}
          value={String(level)}
          // Radix reports "" when the active item is clicked again; keep it.
          onValueChange={(v) =>
            v && onChange({ [factor.key]: Number(v) as Importance })
          }
          aria-label={factor.label}
        >
          {factor.levels.map((name, i) => (
            // Fixed width so every group lines up in columns despite label lengths.
            <ToggleGroupItem
              key={name}
              value={String(i)}
              className="w-12 px-0 text-xs"
            >
              {name}
            </ToggleGroupItem>
          ))}
        </ToggleGroup>
      </div>
      {time && (
        <div className="text-muted-foreground flex h-8 items-center justify-end gap-2 text-xs">
          {time.label}
          <Select
            value={preferences[time.key]}
            onValueChange={(v) => onChange({ [time.key]: v })}
            disabled={level === 0}
          >
            <SelectTrigger size="sm" className="w-28" aria-label={time.label}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {TIME_OPTIONS.map((opt) => (
                <SelectItem key={opt.value} value={opt.value}>
                  {opt.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )}
    </>
  )
}
