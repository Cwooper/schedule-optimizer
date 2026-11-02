import { Fragment } from "react"
import { Calendar } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { useCalendar } from "@/hooks/use-api"
import type { Term, TermCalendar } from "@/lib/api"
import { toISODate } from "@/lib/dates"
import { keyDateRows, termStatus, type KeyDateState } from "@/lib/term-status"
import { cn } from "@/lib/utils"

interface TermStatusProps {
  term: Term | undefined
  terms: Term[]
  onSelectTerm: (code: string) => void
}

const triggerClass =
  "text-muted-foreground hover:text-foreground inline-flex min-w-0 items-baseline gap-1 text-xs leading-none transition-colors"

/** The sidebar's one-line term status, opening the term's key dates. */
export function TermStatus({ term, terms, onSelectTerm }: TermStatusProps) {
  const { data: calendar } = useCalendar(term?.code ?? "")
  if (!term) return null
  const status = termStatus(term, calendar, terms, toISODate(new Date()))
  if (!status) return null

  const content = (
    <>
      {/* Centered beside the text, so the text alone sets the baseline the
          Term label aligns to. */}
      <Calendar className="size-3.5 shrink-0 self-center" />
      <span className="truncate">{status.text}</span>
    </>
  )

  const { switchTo } = status
  if (switchTo) {
    return (
      <button type="button" className={triggerClass} onClick={() => onSelectTerm(switchTo)}>
        {content}
      </button>
    )
  }
  return (
    <Popover>
      <PopoverTrigger asChild>
        <button type="button" className={triggerClass}>
          {content}
        </button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-72">
        {calendar && <KeyDates termName={term.name} calendar={calendar} />}
      </PopoverContent>
    </Popover>
  )
}

const ROW_STATE_CLASS: Record<KeyDateState, string> = {
  past: "text-muted-foreground/50",
  current: "text-foreground font-medium",
  upcoming: "text-muted-foreground",
}

function KeyDates({ termName, calendar }: { termName: string; calendar: TermCalendar }) {
  const rows = keyDateRows(calendar, toISODate(new Date()))
  return (
    <>
      <div className="mb-2 flex h-5 items-center justify-between gap-2">
        <span className="text-sm font-medium">{termName} key dates</span>
        {calendar.estimated && <Badge variant="secondary">Estimated</Badge>}
      </div>
      <div className="grid grid-cols-[5.5rem_1fr] text-xs leading-5 tabular-nums">
        {rows.map((row, i) => (
          <Fragment key={i}>
            <span className={ROW_STATE_CLASS[row.state]}>{row.label}</span>
            <span className={cn("flex flex-wrap gap-x-1.5", ROW_STATE_CLASS[row.state])}>
              {/* A fixed width lines up the names after single dates. */}
              <span className={cn(row.note && "min-w-18")}>{row.dates}</span>
              {row.note && <span>{row.note}</span>}
            </span>
          </Fragment>
        ))}
      </div>
    </>
  )
}
