import { useId } from "react"
import { BookOpen, CalendarDays, type LucideIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import type { HydratedSection } from "@/lib/api"
import { FACTORS, type Factor } from "@/lib/factors"
import {
  computeScheduleStats,
  formatDuration,
  type DayTime,
  type ScheduleStats,
} from "@/lib/schedule-stats"
import { DAYS, DAY_LETTERS, formatClock } from "@/lib/schedule-utils"
import { useAppStore, wasRanked } from "@/stores/app-store"

/** The pager label, which opens the shown schedule's stats. */
export function ScheduleStatsPopover({
  index,
  count,
  sections,
  score,
}: {
  index: number
  count: number
  sections: HydratedSection[]
  /** 0 to 1, from the generator's ranking. */
  score: number
}) {
  const titleId = useId()
  const ranked = useAppStore((s) => wasRanked(s.generatedWithParams))
  return (
    <Popover>
      <PopoverTrigger asChild>
        {/* Sized like the plain label it replaced, so the pager doesn't move. */}
        <Button
          variant="ghost"
          className="min-w-20 px-0 font-normal sm:min-w-32"
        >
          <span>
            <span className="max-sm:hidden">Schedule </span>
            {index + 1} of {count}
          </span>
        </Button>
      </PopoverTrigger>
      <PopoverContent side="bottom" align="center" className="w-80" aria-labelledby={titleId}>
        <div className="mb-3 flex items-center justify-between">
          <span id={titleId} className="text-sm font-medium">
            Schedule {index + 1}
          </span>
          {ranked && (
            <span className="text-sm tabular-nums">
              <Muted>Score </Muted>
              {/* Clamped: results saved by the old scorer could exceed 1. */}
              {Math.round(Math.min(Math.max(score, 0), 1) * 100)}%
            </span>
          )}
        </div>
        <StatRows stats={computeScheduleStats(sections)} />
      </PopoverContent>
    </Popover>
  )
}

function StatRows({ stats }: { stats: ScheduleStats }) {
  const values: Record<Factor, React.ReactNode> = {
    start: <DayTimeValue time={stats.start} />,
    end: <DayTimeValue time={stats.end} />,
    gap: stats.gapMins > 0 ? `${formatDuration(stats.gapMins)} / wk` : <Muted>None</Muted>,
    gpa:
      stats.gpa == null ? (
        <Muted>No data</Muted>
      ) : (
        <>
          {stats.gpa.toFixed(2)}
          {stats.gpaKnown < stats.sections && (
            <Muted> ({stats.gpaKnown} of {classes(stats.sections)})</Muted>
          )}
        </>
      ),
    seats: `${stats.open} of ${stats.sections}`,
  }
  // Each value is one extreme or a mean of the week, which the bare factor
  // name doesn't say.
  const qualifiers: Partial<Record<Factor, string>> = {
    start: "earliest",
    end: "latest",
    gpa: "average",
  }
  const labels: Partial<Record<Factor, string>> = { seats: "Classes open" }
  const days = stats.days.flatMap((on, i) => (on ? [DAY_LETTERS[i + 1]] : []))

  return (
    <>
      {FACTORS.map((f) => (
        <StatRow key={f.key} icon={f.icon} label={labels[f.key] ?? f.label} qualifier={qualifiers[f.key]}>
          {values[f.key]}
        </StatRow>
      ))}
      <div className="bg-border my-2 h-px" />
      <StatRow icon={BookOpen} label="Credits">
        {stats.credits}
      </StatRow>
      <StatRow icon={CalendarDays} label="Days">
        {days.length > 0 ? days.join(" ") : <Muted>None</Muted>}
      </StatRow>
    </>
  )
}

function StatRow({
  icon: Icon,
  label,
  qualifier,
  children,
}: {
  icon: LucideIcon
  label: string
  qualifier?: string
  children: React.ReactNode
}) {
  return (
    <div className="flex h-8 items-center justify-between gap-3">
      <div className="flex items-center gap-2">
        <Icon className="text-muted-foreground size-4" />
        <span className="text-sm">
          {label}
          {qualifier && <Muted> ({qualifier})</Muted>}
        </span>
      </div>
      <span className="text-sm tabular-nums">{children}</span>
    </div>
  )
}

function DayTimeValue({ time }: { time: DayTime | null }) {
  if (!time) return <Muted>None</Muted>
  // Fixed cells keep the weekdays in one column whatever the time's width.
  return (
    <span className="inline-grid grid-cols-[1.75rem_4.25rem]">
      <span className="text-muted-foreground text-right">{DAYS[time.day]}</span>
      <span className="text-right">{formatClock(time.mins)}</span>
    </span>
  )
}

function classes(n: number): string {
  return `${n} ${n === 1 ? "class" : "classes"}`
}

function Muted({ children }: { children: React.ReactNode }) {
  return <span className="text-muted-foreground">{children}</span>
}
