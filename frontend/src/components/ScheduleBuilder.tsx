import { useState, useMemo } from "react"
import { Loader2 } from "lucide-react"
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
  type CourseSlot,
  type SectionFilter,
} from "@/stores/app-store"
import { useTerms, useGenerateSchedules, useValidateCourses } from "@/hooks/use-api"
import type { CourseValidationResult, CourseSpec } from "@/lib/api"
import { cn } from "@/lib/utils"

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
  const generateMutation = useGenerateSchedules()

  const [expandedCourses, setExpandedCourses] = useState<Set<string>>(new Set())

  // Validate courses against current term
  const coursesToValidate = useMemo(
    () => slots.map((s) => ({ subject: s.subject, courseNumber: s.courseNumber })),
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

  const handleAddCourse = (course: {
    subject: string
    courseNumber: string
    title: string
  }) => {
    const id = crypto.randomUUID()
    const slot: CourseSlot = {
      id,
      subject: course.subject,
      courseNumber: course.courseNumber,
      displayName: `${course.subject} ${course.courseNumber}`,
      title: course.title,
      required: false,
      sections: null,
    }
    addSlot(slot)
  }

  const handleAddCrn = (section: {
    crn: string
    term: string
    subject: string
    courseNumber: string
    title: string
    instructor: string
  }) => {
    const sectionFilter: SectionFilter = {
      crn: section.crn,
      term: section.term,
      instructor: section.instructor || null,
      required: true,
    }

    const existingSlot = slots.find(
      (s) =>
        s.subject === section.subject &&
        s.courseNumber === section.courseNumber
    )

    if (existingSlot) {
      const existingSections = existingSlot.sections ?? []
      if (!existingSections.some((s) => s.crn === section.crn)) {
        updateSlot(existingSlot.id, {
          sections: [...existingSections, sectionFilter],
          title: existingSlot.title || section.title,
        })
      }
    } else {
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
  }

  const handleGenerate = () => {
    if (!term || slots.length === 0) return

    if (tab !== "schedule") {
      setTab("schedule")
    }

    const courseSpecs: CourseSpec[] = slots.map((slot) => ({
      subject: slot.subject,
      courseNumber: slot.courseNumber,
      required: slot.required,
      allowedCrns: slot.sections?.map((s) => s.crn),
    }))

    generateMutation.mutate({
      term,
      courseSpecs,
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
                  terms={termsData?.terms ?? []}
                  validation={validationMap.get(`${slot.subject}:${slot.courseNumber}`)}
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
