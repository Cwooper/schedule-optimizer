import { useState } from "react"
import { ChevronsUpDown, Plus, Loader2 } from "lucide-react"
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
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { TermSelector } from "@/components/TermSelector"
import { CourseRow, CRNPreview } from "@/components/schedule-builder"
import {
  useAppStore,
  type CourseSlot,
  type SectionFilter,
} from "@/stores/app-store"
import {
  useTerms,
  useSubjects,
  useGenerateSchedules,
  useCRN,
} from "@/hooks/use-api"
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

  // CRN lookup - only fetch for valid 5-digit CRNs
  const crnTrimmed = crnInput.trim()
  const isValidCrnFormat = /^\d{5}$/.test(crnTrimmed)
  const { data: crnData, isFetching: crnFetching } = useCRN(
    isValidCrnFormat ? crnTrimmed : "",
    term
  )

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
    } else if (inputMode === "crn" && crnInput && crnData?.section) {
      const section = crnData.section
      const sectionFilter: SectionFilter = {
        crn: section.crn,
        term: section.term,
        instructor: section.instructor || null,
        required: true, // CRNs added directly are pinned by default
      }

      // Check if a slot for this course already exists
      const existingSlot = slots.find(
        (s) =>
          s.subject === section.subject &&
          s.courseNumber === section.courseNumber
      )

      if (existingSlot) {
        // Add CRN to existing slot's sections filter
        const existingSections = existingSlot.sections ?? []
        // Don't add duplicate CRNs
        if (!existingSections.some((s) => s.crn === section.crn)) {
          updateSlot(existingSlot.id, {
            sections: [...existingSections, sectionFilter],
            // Update title if we didn't have one
            title: existingSlot.title || section.title,
          })
        }
      } else {
        // Create new slot with this CRN pinned
        const id = crypto.randomUUID()
        const slot: CourseSlot = {
          id,
          subject: section.subject,
          courseNumber: section.courseNumber,
          displayName: `${section.subject} ${section.courseNumber}`,
          title: section.title,
          required: false,
          sections: [sectionFilter],
        }
        addSlot(slot)
      }
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
    minCourses !== null && maxCourses !== null && maxCourses < minCourses

  const getGenerateTooltip = () => {
    if (!term) return "Select a term first"
    if (slots.length === 0) return "Add courses to get started"
    if (hasInvalidBounds) return "Max courses must be ≥ min courses"
    return null
  }

  return (
    <div className="flex h-full flex-col">
      {/* Term Selector */}
      <div className="space-y-4 p-4">
        <TermSelector
          value={term}
          onChange={setTerm}
          terms={termsData?.terms ?? []}
          isLoading={termsLoading}
        />

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
                  const val = e.target.value
                    ? Math.max(0, Math.min(8, Number(e.target.value)))
                    : null
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
                  const val = e.target.value
                    ? Math.max(0, Math.min(8, Number(e.target.value)))
                    : null
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
                    aria-pressed={inputMode === "subject"}
                    aria-label="Add course by subject"
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
                    aria-pressed={inputMode === "crn"}
                    aria-label="Add course by CRN"
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
                          <span className="w-12 shrink-0 font-medium whitespace-nowrap">
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
          <Popover
            open={isValidCrnFormat && (crnFetching || crnData !== undefined)}
            modal={false}
          >
            <PopoverTrigger asChild>
              <div className="flex gap-2">
                <Input
                  placeholder="12345"
                  value={crnInput}
                  onChange={(e) =>
                    setCrnInput(e.target.value.replace(/\D/g, ""))
                  }
                  onKeyDown={(e) => {
                    if (e.key === "Enter" && canAdd) handleAddCourse()
                  }}
                  inputMode="numeric"
                  className="flex-1"
                  maxLength={5}
                />
                <Button
                  size="icon"
                  variant="outline"
                  onClick={handleAddCourse}
                  disabled={!canAdd || (isValidCrnFormat && !crnData?.section)}
                >
                  <Plus className="size-4" />
                </Button>
              </div>
            </PopoverTrigger>
            <PopoverContent
              className="w-72 p-0"
              align="start"
              onOpenAutoFocus={(e) => e.preventDefault()}
              onCloseAutoFocus={(e) => e.preventDefault()}
            >
              <TooltipProvider delayDuration={300}>
                <CRNPreview
                  crnData={crnData}
                  isLoading={crnFetching}
                  currentTerm={term}
                />
              </TooltipProvider>
            </PopoverContent>
          </Popover>
        )}
      </div>

      {/* Divider */}
      <div className="bg-border h-px" />

      {/* Course List */}
      <div className="flex-1 overflow-y-auto p-4">
        <Label className="mb-2 block">Courses</Label>
        <div className="min-h-45 space-y-1">
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
                  onRemoveSection={(crn) => {
                    const newSections = slot.sections?.filter(
                      (s) => s.crn !== crn
                    )
                    if (!newSections || newSections.length === 0) {
                      removeSlot(slot.id)
                    } else {
                      updateSlot(slot.id, { sections: newSections })
                    }
                  }}
                  onToggleSectionRequired={(crn) => {
                    const newSections = slot.sections?.map((s) =>
                      s.crn === crn ? { ...s, required: !s.required } : s
                    )
                    updateSlot(slot.id, { sections: newSections })
                  }}
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
                  disabled={
                    !term ||
                    slots.length === 0 ||
                    hasInvalidBounds ||
                    generateMutation.isPending
                  }
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
              <TooltipContent>{getGenerateTooltip()}</TooltipContent>
            )}
          </Tooltip>
        </TooltipProvider>
      </div>
    </div>
  )
}
