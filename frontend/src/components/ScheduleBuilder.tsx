import { useState, useMemo, useEffect } from "react"
import { toast } from "sonner"
import { CircleHelp, Loader2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { TermSelector } from "@/components/TermSelector"
import { CourseInput, CourseRow } from "@/components/schedule-builder"
import {
  useAppStore,
  computeSlotsFingerprint,
  computeBlockedTimesFingerprint,
  type GenerationParams,
} from "@/stores/app-store"
import {
  useTerms,
  useGenerateSchedules,
  useValidateCourses,
} from "@/hooks/use-api"
import type { CourseValidationResult, CourseSpec, GenerateResponse } from "@/lib/api"
import { getGenerateWarnings } from "@/lib/generate-warnings"
import { cn } from "@/lib/utils"

export function ScheduleBuilder() {
  const tab = useAppStore((s) => s.tab)
  const setTab = useAppStore((s) => s.setTab)
  const term = useAppStore((s) => s.term)
  const setTerm = useAppStore((s) => s.setTerm)
  const selectedSubject = useAppStore((s) => s.selectedSubject)
  const setSelectedSubject = useAppStore((s) => s.setSelectedSubject)
  const minCourses = useAppStore((s) => s.minCourses)
  const maxCourses = useAppStore((s) => s.maxCourses)
  const setCourseBounds = useAppStore((s) => s.setCourseBounds)
  const slots = useAppStore((s) => s.slots)
  const addCourseToSlot = useAppStore((s) => s.addCourseToSlot)
  const addSectionToSlot = useAppStore((s) => s.addSectionToSlot)
  const removeSlot = useAppStore((s) => s.removeSlot)
  const updateSlot = useAppStore((s) => s.updateSlot)
  const setGenerateResult = useAppStore((s) => s.setGenerateResult)
  const getSlotsVersion = useAppStore((s) => s.getSlotsVersion)
  const openCourseDialog = useAppStore((s) => s.openCourseDialog)
  const blockedTimeGroups = useAppStore((s) => s.blockedTimeGroups)
  const regenerateRequested = useAppStore((s) => s.regenerateRequested)
  const clearRegenerateRequest = useAppStore((s) => s.clearRegenerateRequest)

  const { data: termsData, isLoading: termsLoading } = useTerms()
  const generateMutation = useGenerateSchedules()

  // Auto-select the current term on first load if no term is set
  useEffect(() => {
    if (!term && termsData?.current) {
      setTerm(termsData.current)
    }
  }, [term, termsData, setTerm])

  const showGenerateWarnings = (data: GenerateResponse) => {
    for (const w of getGenerateWarnings(data)) {
      if (w.type === "info") toast.info(w.message)
      else toast.warning(w.message)
    }
  }

  // Store result when generation succeeds, but only if slots haven't changed
  const handleGenerateMutate = (
    req: Parameters<typeof generateMutation.mutate>[0],
    params: GenerationParams
  ) => {
    const versionAtStart = getSlotsVersion()
    generateMutation.mutate(req, {
      onSuccess: (data) => {
        // Only set result if slots haven't changed during the request
        if (getSlotsVersion() === versionAtStart) {
          setGenerateResult(data, params)
        }
        showGenerateWarnings(data)
      },
      onError: (error) => {
        toast.error(
          error.message || "Failed to generate schedules. Please try again."
        )
      },
    })
  }

  const [expandedCourses, setExpandedCourses] = useState<Set<string>>(new Set())

  // Validate courses against current term
  const coursesToValidate = useMemo(
    () =>
      slots.map((s) => ({ subject: s.subject, courseNumber: s.courseNumber })),
    [slots]
  )
  const { data: validationData } = useValidateCourses(term, coursesToValidate)

  const validationMap = useMemo(() => {
    const map = new Map<string, CourseValidationResult>()
    validationData?.results.forEach((r) => {
      map.set(`${r.subject}:${r.courseNumber}`, r)
    })
    return map
  }, [validationData])

  // Sort slots alphabetically by subject, then by course number (numeric)
  const sortedSlots = useMemo(
    () =>
      [...slots].sort((a, b) => {
        const subjectCmp = a.subject.localeCompare(b.subject)
        if (subjectCmp !== 0) return subjectCmp
        return parseInt(a.courseNumber) - parseInt(b.courseNumber)
      }),
    [slots]
  )

  const handleAddCourse = (course: {
    subject: string
    courseNumber: string
    title: string
  }) => {
    addCourseToSlot(course.subject, course.courseNumber, course.title)
  }

  const handleAddCrn = (section: {
    crn: string
    term: string
    subject: string
    courseNumber: string
    title: string
    instructor: string
  }) => {
    addSectionToSlot(
      section.crn,
      section.term,
      section.subject,
      section.courseNumber,
      section.title,
      section.instructor || null
    )
  }

  // Filter to only courses that exist in the current term
  const validSlots = useMemo(
    () =>
      slots.filter((slot) => {
        const validation = validationMap.get(
          `${slot.subject}:${slot.courseNumber}`
        )
        return validation?.exists !== false
      }),
    [slots, validationMap]
  )

  const handleGenerate = () => {
    if (!term || validSlots.length === 0) return

    if (tab !== "schedule") {
      setTab("schedule")
    }

    const courseSpecs: CourseSpec[] = validSlots.map((slot) => ({
      subject: slot.subject,
      courseNumber: slot.courseNumber,
      required: slot.required,
      allowedCrns: slot.sections?.map((s) => s.crn),
    }))

    const effectiveMin = minCourses ?? 0
    const effectiveMax = maxCourses ?? 0

    // Capture params at time of generation for stale detection.
    // Uses `slots` (not `validSlots`) so fingerprint tracks user intent -
    // if a course becomes valid later, stale detection still works correctly.
    const params: GenerationParams = {
      term,
      minCourses: effectiveMin,
      maxCourses: effectiveMax,
      slotsFingerprint: computeSlotsFingerprint(slots),
      blockedTimesFingerprint: computeBlockedTimesFingerprint(
        blockedTimeGroups
      ),
    }

    const blockedTimes = blockedTimeGroups
      .filter((g) => g.enabled)
      .flatMap((g) => g.blocks)

    handleGenerateMutate(
      {
        term,
        courseSpecs,
        blockedTimes: blockedTimes.length > 0 ? blockedTimes : undefined,
        minCourses: effectiveMin,
        maxCourses: effectiveMax,
      },
      params
    )
  }

  // Handle regenerate requests from other components (e.g., stale warning click)
  useEffect(() => {
    if (regenerateRequested) {
      clearRegenerateRequest()
      handleGenerate()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [regenerateRequested])

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

  const hasInvalidBounds =
    minCourses !== null && maxCourses !== null && maxCourses < minCourses

  const hasNoValidCourses = slots.length > 0 && validSlots.length === 0

  const getGenerateTooltip = () => {
    if (!term) return "Select a term first"
    if (slots.length === 0) return "Add courses to get started"
    if (hasNoValidCourses) return "No courses available in this term"
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
                    <CircleHelp className="text-muted-foreground size-4 cursor-help" />
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
                    <CircleHelp className="text-muted-foreground size-4 cursor-help" />
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
      <CourseInput
        term={term}
        selectedSubject={selectedSubject}
        onSubjectChange={setSelectedSubject}
        onAddCourse={handleAddCourse}
        onAddCrn={handleAddCrn}
      />

      {/* Divider */}
      <div className="bg-border h-px" />

      {/* Course List */}
      <div className="scrollbar-styled flex-1 overflow-y-auto p-4">
        <Label className="mb-2 block">Courses</Label>
        <div className="min-h-45 space-y-1">
          {sortedSlots.length === 0 ? (
            <p className="text-muted-foreground py-4 text-center text-sm">
              Add courses above to get started
            </p>
          ) : (
            <TooltipProvider delayDuration={300}>
              {sortedSlots.map((slot) => (
                <CourseRow
                  key={slot.id}
                  slot={slot}
                  expanded={expandedCourses.has(slot.id)}
                  onToggleExpand={() => toggleExpanded(slot.id)}
                  onToggleRequired={() =>
                    updateSlot(slot.id, { required: !slot.required })
                  }
                  onRemove={() => removeSlot(slot.id)}
                  onRemoveSection={(crn, sectionTerm) => {
                    const newSections = slot.sections?.filter(
                      (s) => !(s.crn === crn && s.term === sectionTerm)
                    )
                    // Set to null (all sections) when last specific section is removed
                    updateSlot(slot.id, {
                      sections: newSections?.length ? newSections : null,
                    })
                  }}
                  onToggleSectionRequired={(crn, sectionTerm) => {
                    const newSections = slot.sections?.map((s) =>
                      s.crn === crn && s.term === sectionTerm ? { ...s, required: !s.required } : s
                    )
                    updateSlot(slot.id, { sections: newSections })
                  }}
                  onCourseClick={(courseKey) => openCourseDialog({ courseKey, source: "schedule" })}
                  onSectionClick={(crn) =>
                    openCourseDialog({
                      crn,
                      courseKey: `${slot.subject}:${slot.courseNumber}`,
                      source: "schedule",
                    })
                  }
                  currentTerm={term}
                  terms={termsData?.terms ?? []}
                  validation={validationMap.get(
                    `${slot.subject}:${slot.courseNumber}`
                  )}
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
                    hasNoValidCourses ||
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
            {(!term ||
              slots.length === 0 ||
              hasNoValidCourses ||
              hasInvalidBounds) && (
              <TooltipContent>{getGenerateTooltip()}</TooltipContent>
            )}
          </Tooltip>
        </TooltipProvider>
      </div>
    </div>
  )
}
