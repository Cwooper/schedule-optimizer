import { CalendarPlus, CircleHelp, Download } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Label } from "@/components/ui/label"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { useCalendar, useTerms } from "@/hooks/use-api"
import type { HydratedSection, RegistrationPhaseName } from "@/lib/api"
import { buildIcs, hasIcsEvents, icsFilename, type CalendarExportOptions } from "@/lib/ics"
import { selectScheduleTerm, useAppStore } from "@/stores/app-store"

const PHASES: RegistrationPhaseName[] = ["I", "II", "III"]

function downloadFile(filename: string, content: string) {
  const url = URL.createObjectURL(new Blob([content], { type: "text/calendar;charset=utf-8" }))
  const a = document.createElement("a")
  a.href = url
  a.download = filename
  a.click()
  // The download reads the URL asynchronously with no completion event, and
  // revoking at once cancels it in Safari.
  setTimeout(() => URL.revokeObjectURL(url), 60_000)
}

function OptionRow({
  id,
  label,
  checked,
  disabled,
  onChange,
  children,
}: {
  id: string
  label: string
  checked: boolean | "indeterminate"
  disabled?: boolean
  onChange: (checked: boolean) => void
  children?: React.ReactNode
}) {
  return (
    <div className="flex h-5 items-center gap-2">
      <Checkbox
        id={id}
        checked={checked}
        disabled={disabled}
        onCheckedChange={(value) => onChange(value === true)}
      />
      <Label htmlFor={id} className="font-normal peer-disabled:opacity-50">
        {label}
      </Label>
      {children}
    </div>
  )
}

/** Exports the current schedule and the term's WWU dates as an .ics file. */
export function AddToCalendar({ sections }: { sections: HydratedSection[] }) {
  const term = useAppStore(selectScheduleTerm)
  const options = useAppStore((s) => s.calendarExport)
  const setOptions = useAppStore((s) => s.setCalendarExport)
  const { data: termsData } = useTerms()
  const { data: calendar } = useCalendar(term)
  const termName = termsData?.terms.find((t) => t.code === term)?.name ?? term

  const hasSchedule = sections.length > 0
  const effective: CalendarExportOptions = {
    ...options,
    classes: options.classes && hasSchedule,
    finals: options.finals && hasSchedule,
  }
  const canDownload =
    calendar != null && hasIcsEvents({ calendar, sections, options: effective })

  const phasesOn = PHASES.filter((p) => options.registration[p]).length
  let registrationChecked: boolean | "indeterminate" = phasesOn > 0
  if (phasesOn > 0 && phasesOn < PHASES.length) registrationChecked = "indeterminate"
  const setAllPhases = (on: boolean) =>
    setOptions({ registration: { I: on, II: on, III: on } })

  const handleDownload = () => {
    if (!calendar) return
    downloadFile(
      icsFilename(termName),
      buildIcs({ calendar, sections, options: effective, now: new Date() })
    )
  }

  return (
    <Popover>
      <Tooltip>
        <TooltipTrigger asChild>
          <PopoverTrigger asChild>
            <Button variant="ghost" size="icon" disabled={!term}>
              <CalendarPlus className="size-4" />
              <span className="sr-only">Add to Calendar</span>
            </Button>
          </PopoverTrigger>
        </TooltipTrigger>
        <TooltipContent side="bottom">Add to Calendar</TooltipContent>
      </Tooltip>
      <PopoverContent align="end" className="w-72">
        <div className="mb-3 flex h-5 items-center justify-between gap-2">
          <span className="truncate text-sm font-medium">Add to Calendar - {termName}</span>
          <Tooltip>
            <TooltipTrigger asChild>
              <span className="inline-flex cursor-help">
                <CircleHelp className="text-muted-foreground size-3.5" />
              </span>
            </TooltipTrigger>
            <TooltipContent>Import it into your calendar of choice.</TooltipContent>
          </Tooltip>
        </div>
        <div className="space-y-2.5">
          <OptionRow
            id="ics-classes"
            label="Classes"
            checked={effective.classes}
            disabled={!hasSchedule}
            onChange={(classes) => setOptions({ classes })}
          />
          <OptionRow
            id="ics-finals"
            label="Final exams"
            checked={effective.finals}
            disabled={!hasSchedule}
            onChange={(finals) => setOptions({ finals })}
          />
          <OptionRow
            id="ics-holidays"
            label="Holidays"
            checked={options.holidays}
            onChange={(holidays) => setOptions({ holidays })}
          />
          <OptionRow
            id="ics-registration"
            label="Registration"
            checked={registrationChecked}
            onChange={() => setAllPhases(phasesOn < PHASES.length)}
          >
            <div className="ml-auto flex items-center gap-2.5">
              {PHASES.map((phase) => (
                <div key={phase} className="flex items-center gap-1">
                  <Checkbox
                    id={`ics-phase-${phase}`}
                    checked={options.registration[phase]}
                    onCheckedChange={(value) =>
                      setOptions({ registration: { ...options.registration, [phase]: value === true } })
                    }
                  />
                  <Label htmlFor={`ics-phase-${phase}`} className="text-xs font-normal">
                    {phase}
                  </Label>
                </div>
              ))}
            </div>
          </OptionRow>
          <OptionRow
            id="ics-deadlines"
            label="Deadlines"
            checked={options.deadlines}
            onChange={(deadlines) => setOptions({ deadlines })}
          >
            <span className="text-muted-foreground text-xs">add, drop, withdraw</span>
          </OptionRow>
        </div>
        <Button size="sm" className="mt-4 w-full" disabled={!canDownload} onClick={handleDownload}>
          <Download className="size-4" />
          Download .ics
        </Button>
      </PopoverContent>
    </Popover>
  )
}
