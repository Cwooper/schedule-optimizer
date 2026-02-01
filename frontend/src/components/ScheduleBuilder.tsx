import { useState } from "react"
import {
  ChevronDown,
  ChevronRight,
  ChevronsUpDown,
  Circle,
  CircleDot,
  Plus,
  X,
  AlertTriangle,
  Loader2,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Toggle } from "@/components/ui/toggle"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"
import { useAppStore, type CourseSlot } from "@/stores/app-store"
import { useTerms, useSubjects, useGenerateSchedules } from "@/hooks/use-api"
import { cn } from "@/lib/utils"

type InputMode = "subject" | "crn"

export function ScheduleBuilder() {
  const {
    tab,
    setTab,
    term,
    setTerm,
    selectedSubject,
    setSelectedSubject,
    minCourses,
    maxCourses,
    setCourseBounds,
    slots,
    addSlot,
    removeSlot,
    updateSlot,
  } = useAppStore()

  const { data: termsData, isLoading: termsLoading } = useTerms()
  const { data: subjectsData } = useSubjects(term)
  const generateMutation = useGenerateSchedules()

  const [inputMode, setInputMode] = useState<InputMode>("subject")
  const [subjectOpen, setSubjectOpen] = useState(false)
  const [numberInput, setNumberInput] = useState("")
  const [crnInput, setCrnInput] = useState("")
  const [expandedCourses, setExpandedCourses] = useState<Set<string>>(new Set())

  const subjects = subjectsData?.subjects ?? []

  // Decode HTML entities in subject names
  const decodeHtml = (html: string) => {
    const txt = document.createElement("textarea")
    txt.innerHTML = html
    return txt.value
  }

  const handleAddCourse = () => {
    if (inputMode === "subject" && selectedSubject && numberInput) {
      const id = crypto.randomUUID()
      const normalizedNumber = numberInput.trim().toUpperCase()
      const slot: CourseSlot = {
        id,
        subject: selectedSubject,
        courseNumber: normalizedNumber,
        displayName: `${selectedSubject} ${normalizedNumber}`,
        required: false,
        sections: null,
      }
      addSlot(slot)
      setNumberInput("")
    } else if (inputMode === "crn" && crnInput) {
      // TODO: Fetch CRN details and add as a course with that specific section
      setCrnInput("")
    }
  }

  const handleGenerate = () => {
    if (!term || slots.length === 0) return

    if (tab !== "schedule") {
      setTab("schedule")
    }

    generateMutation.mutate({
      term,
      courses: slots.map((slot) => `${slot.subject} ${slot.courseNumber}`),
      minCourses: minCourses ?? undefined,
      maxCourses: maxCourses ?? undefined,
    })
  }

  const toggleExpanded = (id: string) => {
    setExpandedCourses((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  const courseNumberPattern = /^\d{3}[A-Za-z]?$/
  const isValidCourseNumber = courseNumberPattern.test(numberInput.trim())

  const canAdd =
    inputMode === "subject"
      ? selectedSubject && isValidCourseNumber
      : crnInput.trim()

  const hasInvalidBounds =
    minCourses !== null &&
    maxCourses !== null &&
    maxCourses < minCourses

  return (
    <div className="flex h-full flex-col">
      {/* Term Selector */}
      <div className="space-y-4 p-4">
        <div className="space-y-2">
          <Label htmlFor="term">Term</Label>
          <Select value={term} onValueChange={setTerm} disabled={termsLoading}>
            <SelectTrigger id="term" className="w-full">
              <SelectValue placeholder="Select term..." />
            </SelectTrigger>
            <SelectContent>
              {termsData?.terms.map((t) => (
                <SelectItem key={t.code} value={t.code}>
                  {t.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Course Bounds */}
        <TooltipProvider delayDuration={300}>
          <div className="flex gap-3">
            <div className="flex-1 space-y-2">
              <div className="flex items-center gap-1">
                <Label htmlFor="min-courses">Min Courses</Label>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <span className="text-muted-foreground cursor-help text-xs">
                      (?)
                    </span>
                  </TooltipTrigger>
                  <TooltipContent>
                    Minimum number of courses to include in each schedule
                  </TooltipContent>
                </Tooltip>
              </div>
              <Input
                id="min-courses"
                type="number"
                inputMode="numeric"
                min={0}
                max={8}
                placeholder="—"
                value={minCourses ?? ""}
                onChange={(e) => {
                  const val = e.target.value ? Math.max(0, Math.min(8, Number(e.target.value))) : null
                  setCourseBounds(val, maxCourses)
                }}
                className={cn(hasInvalidBounds && "border-destructive")}
              />
            </div>
            <div className="flex-1 space-y-2">
              <div className="flex items-center gap-1">
                <Label htmlFor="max-courses">Max Courses</Label>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <span className="text-muted-foreground cursor-help text-xs">
                      (?)
                    </span>
                  </TooltipTrigger>
                  <TooltipContent>
                    Maximum number of courses to include in each schedule
                  </TooltipContent>
                </Tooltip>
              </div>
              <Input
                id="max-courses"
                type="number"
                inputMode="numeric"
                min={0}
                max={8}
                placeholder="—"
                value={maxCourses ?? ""}
                onChange={(e) => {
                  const val = e.target.value ? Math.max(0, Math.min(8, Number(e.target.value))) : null
                  setCourseBounds(minCourses, val)
                }}
                className={cn(hasInvalidBounds && "border-destructive")}
              />
            </div>
          </div>
        </TooltipProvider>
      </div>

      {/* Divider */}
      <div className="bg-border h-px" />

      {/* Add Course Section */}
      <div className="space-y-3 p-4">
        <div className="flex items-center justify-between">
          <Label>Add Course</Label>
          <TooltipProvider delayDuration={300}>
            <div className="flex gap-1 text-sm">
              <Tooltip>
                <TooltipTrigger asChild>
                  <button
                    className={cn(
                      "rounded px-2 py-0.5 transition-colors",
                      inputMode === "subject"
                        ? "bg-accent text-accent-foreground"
                        : "text-muted-foreground hover:text-foreground"
                    )}
                    onClick={() => setInputMode("subject")}
                  >
                    Subject
                  </button>
                </TooltipTrigger>
                <TooltipContent>
                  Search by subject code and course number (e.g., CSCI 241)
                </TooltipContent>
              </Tooltip>
              <span className="text-muted-foreground">|</span>
              <Tooltip>
                <TooltipTrigger asChild>
                  <button
                    className={cn(
                      "rounded px-2 py-0.5 transition-colors",
                      inputMode === "crn"
                        ? "bg-accent text-accent-foreground"
                        : "text-muted-foreground hover:text-foreground"
                    )}
                    onClick={() => setInputMode("crn")}
                  >
                    CRN
                  </button>
                </TooltipTrigger>
                <TooltipContent>
                  Add a specific section by its Course Reference Number
                </TooltipContent>
              </Tooltip>
            </div>
          </TooltipProvider>
        </div>

        {inputMode === "subject" ? (
          <div className="flex gap-2">
            <Popover open={subjectOpen} onOpenChange={setSubjectOpen}>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  role="combobox"
                  aria-expanded={subjectOpen}
                  className="flex-1 justify-between px-2"
                  disabled={!term}
                >
                  <span className="truncate">
                    {selectedSubject || "Subject"}
                  </span>
                  <ChevronsUpDown className="size-4 shrink-0 opacity-50" />
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-64 p-0" align="start">
                <Command>
                  <CommandInput placeholder="Search subjects..." />
                  <CommandList>
                    <CommandEmpty>No subject found.</CommandEmpty>
                    <CommandGroup>
                      {subjects.map((subject) => (
                        <CommandItem
                          key={subject.code}
                          value={`${subject.code} ${subject.name}`}
                          onSelect={() => {
                            setSelectedSubject(subject.code)
                            setSubjectOpen(false)
                          }}
                          className={cn(
                            selectedSubject === subject.code && "bg-accent"
                          )}
                        >
                          <span className="w-12 shrink-0 whitespace-nowrap font-medium">
                            {subject.code}
                          </span>
                          <span className="text-muted-foreground truncate">
                            {decodeHtml(subject.name)}
                          </span>
                        </CommandItem>
                      ))}
                    </CommandGroup>
                  </CommandList>
                </Command>
              </PopoverContent>
            </Popover>
            <Input
              placeholder="241"
              value={numberInput}
              onChange={(e) => setNumberInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && canAdd) handleAddCourse()
              }}
              className="w-16 flex-none"
              maxLength={4}
            />
            <Button
              size="icon"
              variant="outline"
              onClick={handleAddCourse}
              disabled={!canAdd}
            >
              <Plus className="size-4" />
            </Button>
          </div>
        ) : (
          <div className="flex gap-2">
            <Input
              placeholder="12345"
              value={crnInput}
              onChange={(e) => setCrnInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && canAdd) handleAddCourse()
              }}
              className="flex-1"
              maxLength={6}
            />
            <Button
              size="icon"
              variant="outline"
              onClick={handleAddCourse}
              disabled={!canAdd}
            >
              <Plus className="size-4" />
            </Button>
          </div>
        )}
      </div>

      {/* Divider */}
      <div className="bg-border h-px" />

      {/* Course List */}
      <div className="flex-1 overflow-y-auto p-4">
        <Label className="mb-2 block">Courses</Label>
        <div className="min-h-[180px] space-y-1">
          {slots.length === 0 ? (
            <p className="text-muted-foreground py-4 text-center text-sm">
              Add courses above to get started
            </p>
          ) : (
            <TooltipProvider delayDuration={300}>
              {slots.map((slot) => (
                <CourseRow
                  key={slot.id}
                  slot={slot}
                  expanded={expandedCourses.has(slot.id)}
                  onToggleExpand={() => toggleExpanded(slot.id)}
                  onToggleRequired={() =>
                    updateSlot(slot.id, { required: !slot.required })
                  }
                  onRemove={() => removeSlot(slot.id)}
                  currentTerm={term}
                />
              ))}
            </TooltipProvider>
          )}
        </div>
      </div>

      {/* Generate Button - Sticky Bottom */}
      <div className="bg-background sticky bottom-0 border-t p-4">
        <TooltipProvider delayDuration={0}>
          <Tooltip>
            <TooltipTrigger asChild>
              <div className="w-full">
                <Button
                  className="w-full"
                  onClick={handleGenerate}
                  disabled={!term || slots.length === 0 || hasInvalidBounds || generateMutation.isPending}
                >
                  {generateMutation.isPending ? (
                    <>
                      <Loader2 className="size-4 animate-spin" />
                      Generating...
                    </>
                  ) : (
                    "Generate"
                  )}
                </Button>
              </div>
            </TooltipTrigger>
            {(!term || slots.length === 0 || hasInvalidBounds) && (
              <TooltipContent>
                {!term
                  ? "Select a term first"
                  : slots.length === 0
                    ? "Add courses to get started"
                    : hasInvalidBounds
                      ? "Max courses must be ≥ min courses"
                      : null}
              </TooltipContent>
            )}
          </Tooltip>
        </TooltipProvider>
      </div>
    </div>
  )
}

interface CourseRowProps {
  slot: CourseSlot
  expanded: boolean
  onToggleExpand: () => void
  onToggleRequired: () => void
  onRemove: () => void
  currentTerm: string
}

function CourseRow({
  slot,
  expanded,
  onToggleExpand,
  onToggleRequired,
  onRemove,
  currentTerm,
}: CourseRowProps) {
  const hasSections = slot.sections && slot.sections.length > 0
  const hasTermMismatch =
    slot.sections?.some((s) => s.term !== currentTerm) ?? false

  return (
    <Collapsible open={expanded} onOpenChange={onToggleExpand}>
      <div
        className={cn(
          "hover:bg-muted/50 flex items-center gap-2 rounded px-2 py-1.5 transition-colors",
          hasTermMismatch && "text-muted-foreground"
        )}
      >
        {/* Expand/collapse trigger */}
        <CollapsibleTrigger asChild disabled={!hasSections}>
          <button
            className={cn(
              "size-5 shrink-0",
              !hasSections && "cursor-default opacity-0"
            )}
          >
            {expanded ? (
              <ChevronDown className="size-4" />
            ) : (
              <ChevronRight className="size-4" />
            )}
          </button>
        </CollapsibleTrigger>

        {/* Warning icon for term mismatch */}
        {hasTermMismatch && (
          <Tooltip>
            <TooltipTrigger asChild>
              <AlertTriangle className="text-warning size-4 shrink-0 text-amber-500" />
            </TooltipTrigger>
            <TooltipContent>
              Some sections are from a different term
            </TooltipContent>
          </Tooltip>
        )}

        {/* Course name */}
        <span className="flex-1 truncate text-sm font-medium">
          {slot.displayName}
        </span>

        {/* Required toggle */}
        <Tooltip>
          <TooltipTrigger asChild>
            <Toggle
              size="sm"
              pressed={slot.required}
              onPressedChange={onToggleRequired}
              className="size-7 p-0"
            >
              {slot.required ? (
                <CircleDot className="size-4" />
              ) : (
                <Circle className="size-4" />
              )}
            </Toggle>
          </TooltipTrigger>
          <TooltipContent>
            {slot.required ? "Required" : "Optional"}
          </TooltipContent>
        </Tooltip>

        {/* Remove button */}
        <Button
          variant="ghost"
          size="icon"
          className="size-7"
          onClick={onRemove}
          aria-label={`Remove ${slot.displayName}`}
        >
          <X className="size-4" />
        </Button>
      </div>

      {/* Expanded section list */}
      {hasSections && (
        <CollapsibleContent>
          <div className="ml-7 space-y-1 py-1">
            {slot.sections!.map((section) => (
              <div
                key={section.crn}
                className={cn(
                  "hover:bg-muted/50 flex items-center gap-2 rounded px-2 py-1 text-sm",
                  section.term !== currentTerm && "text-muted-foreground"
                )}
              >
                {section.term !== currentTerm && (
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <AlertTriangle className="size-3.5 shrink-0 text-amber-500" />
                    </TooltipTrigger>
                    <TooltipContent>
                      This section is from {section.term}
                    </TooltipContent>
                  </Tooltip>
                )}
                <span className="flex-1 truncate">
                  {section.crn}
                  {section.instructor && ` (${section.instructor})`}
                </span>
                <Toggle
                  size="sm"
                  pressed={section.required}
                  onPressedChange={() => {
                    // TODO: Update section required state
                  }}
                  className="size-6 p-0"
                >
                  {section.required ? (
                    <CircleDot className="size-3.5" />
                  ) : (
                    <Circle className="size-3.5" />
                  )}
                </Toggle>
                <Button
                  variant="ghost"
                  size="icon"
                  className="size-6"
                  aria-label={`Remove section ${section.crn}`}
                >
                  <X className="size-3.5" />
                </Button>
              </div>
            ))}
          </div>
        </CollapsibleContent>
      )}
    </Collapsible>
  )
}
