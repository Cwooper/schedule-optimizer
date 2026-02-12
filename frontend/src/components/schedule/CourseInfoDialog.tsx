import { useMemo, useState, useCallback, useRef } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { useVirtualizer } from "@tanstack/react-virtual"
import { BookOpen, Users, Loader2, TrendingUp, CircleHelp } from "lucide-react"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { SectionCard } from "./SectionCard"
import { useCourse, useCRN } from "@/hooks/use-api"
import { getCRN } from "@/lib/api"
import type {
  HydratedSection,
  GenerateCourseInfo,
  CRNResponse,
} from "@/lib/api"
import { hydrateSection, sortSectionsByAvailability, type SectionInfoLike } from "@/lib/schedule-utils"
import { decodeHtmlEntities, formatTermCode } from "@/lib/utils"

interface CourseInfoDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  // Course data from generate/search response (optional - will fetch if not provided)
  courses?: Record<string, GenerateCourseInfo>
  sections?: Record<string, SectionInfoLike>
  // Selection - either CRN or courseKey
  selectedCrn?: string
  selectedCourseKey?: string // "SUBJECT:NUMBER" format
  // Term for fetching data when generate result isn't available
  term?: string
  // Add section action - when provided, shows "+" buttons on sections
  onAddSection?: (crn: string, term: string, course: { subject: string; courseNumber: string; title: string }, instructor: string | null) => void
  // Check if a section is already added
  isSectionAdded?: (crn: string, term: string) => boolean
}

interface DialogContentInnerProps {
  course: GenerateCourseInfo
  courseSections: HydratedSection[]
  highlightCrn: string | null
  term: string
  onAddSection?: (crn: string, term: string, course: { subject: string; courseNumber: string; title: string }, instructor: string | null) => void
  isSectionAdded?: (crn: string, term: string) => boolean
}

function DialogContentInner({
  course,
  courseSections,
  highlightCrn,
  term,
  onAddSection,
  isSectionAdded,
}: DialogContentInnerProps) {
  const queryClient = useQueryClient()

  // Initialize expanded section to the highlighted one (using term:crn composite
  // key for cross-term safety — CRNs can be reused across terms)
  const [expandedSection, setExpandedSection] = useState<string | null>(() => {
    if (!highlightCrn) return null
    const match = courseSections.find((s) => s.crn === highlightCrn)
    return match ? `${match.term}:${match.crn}` : null
  })

  // Find the expanded section to check if it needs meeting times
  const expandedSectionData = expandedSection
    ? courseSections.find((s) => `${s.term}:${s.crn}` === expandedSection)
    : null

  // Use the section's own term for CRN fetches — search sections can span
  // multiple terms, so the global term prop may not match.
  const expandedTerm = expandedSectionData?.term || term

  // Check if we already have cached data for the expanded section
  const cachedExpandedData = expandedSectionData
    ? queryClient.getQueryData<CRNResponse>(["crn", expandedSectionData.crn, expandedTerm])
    : null

  const needsFetch =
    expandedSectionData &&
    expandedSectionData.meetingTimes.length === 0 &&
    !cachedExpandedData?.section?.meetingTimes

  // Fetch meeting times for expanded section if needed
  const { isFetching } = useCRN(
    needsFetch ? expandedSectionData.crn : "",
    needsFetch ? expandedTerm : undefined
  )

  // Prefetch section details on hover or focus
  const handlePrefetch = useCallback(
    (crn: string, sectionTerm: string) => {
      const section = courseSections.find((s) => s.crn === crn && s.term === sectionTerm)
      if (!section) return
      const cached = queryClient.getQueryData<CRNResponse>(["crn", crn, sectionTerm])
      // Only prefetch if section doesn't have meeting times and not already cached
      if (section.meetingTimes.length === 0 && !cached && section.term) {
        queryClient.prefetchQuery({
          queryKey: ["crn", crn, section.term],
          queryFn: () => getCRN(crn, section.term),
          staleTime: 5 * 60 * 1000, // 5 minutes
        })
      }
    },
    [courseSections, queryClient]
  )

  // Merge cached meeting times into sections from TanStack Query cache.
  // Computed every render (not memoized) because getQueryData is a synchronous
  // cache read that doesn't subscribe to updates — a useMemo here would miss
  // cache changes from prefetch or when the useCRN query key switches on completion.
  const sectionsWithMeetings = courseSections.map((section) => {
    if (section.meetingTimes.length > 0) return section
    const cached = queryClient.getQueryData<CRNResponse>([
      "crn",
      section.crn,
      section.term,
    ])
    if (cached?.section?.meetingTimes) {
      return { ...section, meetingTimes: cached.section.meetingTimes }
    }
    return section
  })

  const totalSections = courseSections.length
  const useVirtual = totalSections > 50

  const scrollRef = useRef<HTMLDivElement>(null)
  // eslint-disable-next-line react-hooks/incompatible-library -- useVirtualizer returns mutable refs; React Compiler skips this component, which is fine since it's already isolated in DialogContentInner
  const virtualizer = useVirtualizer({
    count: useVirtual ? sectionsWithMeetings.length : 0,
    getScrollElement: () => scrollRef.current,
    estimateSize: () => 52,
    overscan: 10,
    enabled: useVirtual,
  })

  const renderSectionCard = (section: HydratedSection) => (
    <div key={`${section.term}:${section.crn}`} className="pb-2">
      <SectionCard
        section={section}
        expanded={expandedSection === `${section.term}:${section.crn}`}
        onToggleExpand={() => {
          const key = `${section.term}:${section.crn}`
          setExpandedSection((prev) => prev === key ? null : key)
        }}
        highlighted={section.crn === highlightCrn}
        isLoadingDetails={
          expandedSection === `${section.term}:${section.crn}` &&
          isFetching &&
          section.meetingTimes.length === 0
        }
        onPrefetch={() => handlePrefetch(section.crn, section.term)}
        onAdd={onAddSection ? () => onAddSection(section.crn, section.term, { subject: course.subject, courseNumber: course.courseNumber, title: course.title }, section.instructor ?? null) : undefined}
        isAdded={isSectionAdded?.(section.crn, section.term) ?? false}
        termLabel={formatTermCode(section.term)}
      />
    </div>
  )

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          {course.subject} {course.courseNumber} - {decodeHtmlEntities(course.title)}
        </DialogTitle>
      </DialogHeader>

      {/* Stats badges */}
      <div className="flex flex-wrap gap-2">
        <Badge variant="secondary" className="gap-1.5 rounded-md">
          <BookOpen className="size-3.5" />
          {course.credits} credit{course.credits !== 1 && "s"}
        </Badge>
        <Badge variant="secondary" className="gap-1.5 rounded-md">
          <Users className="size-3.5" />
          {totalSections} section{totalSections !== 1 && "s"}
        </Badge>
        <Badge variant="secondary" className="gap-1.5 rounded-md">
          <TrendingUp className="size-3.5" />
          {course.gpa
            ? `${course.gpa.toFixed(2)} GPA`
            : course.passRate != null
              ? `${Math.round(course.passRate * 100)}% pass`
              : "N/A"}
          <Tooltip>
            <TooltipTrigger asChild>
              <span className="inline-flex cursor-help">
                <CircleHelp className="size-3.5 text-muted-foreground" />
              </span>
            </TooltipTrigger>
            <TooltipContent>
              {course.gpa
                ? "Average GPA for this course"
                : course.passRate != null
                  ? "Historical pass rate for this S/U graded course"
                  : "No historical grade data available"}
            </TooltipContent>
          </Tooltip>
        </Badge>
      </div>

      {/* Section list — plain render for ≤50 sections, virtualized for large lists */}
      <div
        ref={scrollRef}
        className="scrollbar-styled -mx-2 max-h-[50vh] overflow-y-auto px-2 py-1"
      >
        {useVirtual ? (
          <div
            className="relative w-full"
            style={{ height: virtualizer.getTotalSize() }}
          >
            {virtualizer.getVirtualItems().map((virtualItem) => {
              const section = sectionsWithMeetings[virtualItem.index]
              return (
                <div
                  key={`${section.term}:${section.crn}`}
                  data-index={virtualItem.index}
                  ref={virtualizer.measureElement}
                  className="absolute left-0 w-full"
                  style={{ transform: `translateY(${virtualItem.start}px)` }}
                >
                  {renderSectionCard(section)}
                </div>
              )
            })}
          </div>
        ) : (
          sectionsWithMeetings.map(renderSectionCard)
        )}
      </div>
    </>
  )
}

export function CourseInfoDialog({
  open,
  onOpenChange,
  courses,
  sections,
  selectedCrn,
  selectedCourseKey,
  term,
  onAddSection,
  isSectionAdded,
}: CourseInfoDialogProps) {
  // Parse courseKey to get subject and courseNumber for fetching
  const parsedCourseKey = useMemo(() => {
    if (selectedCourseKey) {
      const [subject, courseNumber] = selectedCourseKey.split(":")
      return { subject, courseNumber }
    }
    // If we have a CRN and sections, get the courseKey from there
    if (selectedCrn && sections?.[selectedCrn]) {
      const [subject, courseNumber] = sections[selectedCrn].courseKey.split(":")
      return { subject, courseNumber }
    }
    return null
  }, [selectedCourseKey, selectedCrn, sections])

  // Check if the specific course exists in generateResult
  const courseKeyForLookup = parsedCourseKey
    ? `${parsedCourseKey.subject}:${parsedCourseKey.courseNumber}`
    : null
  const hasCourseInGenerate = courseKeyForLookup && courses?.[courseKeyForLookup]

  // Fetch course data if we don't have it from generateResult
  const shouldFetch = open && parsedCourseKey && term && !hasCourseInGenerate
  const { data: fetchedCourse, isLoading } = useCourse(
    shouldFetch ? term : "",
    shouldFetch ? parsedCourseKey?.subject ?? "" : "",
    shouldFetch ? parsedCourseKey?.courseNumber ?? "" : ""
  )

  // Derive course info and relevant sections from generateResult
  const dialogDataFromGenerate = useMemo(() => {
    if (!courses || !sections) return null

    let courseKey: string | null = null
    const highlightCrn: string | null = selectedCrn ?? null

    // If opened by CRN, find the course key from that section
    if (selectedCrn && sections[selectedCrn]) {
      courseKey = sections[selectedCrn].courseKey
    } else if (selectedCourseKey) {
      courseKey = selectedCourseKey
    }

    if (!courseKey || !courses[courseKey]) return null

    const course = courses[courseKey]

    // Find all sections for this course
    const courseSections: HydratedSection[] = []
    for (const [crn, sectionInfo] of Object.entries(sections)) {
      if (sectionInfo.courseKey === courseKey) {
        const hydrated = hydrateSection(crn, courses, sections)
        if (hydrated) {
          courseSections.push(hydrated)
        }
      }
    }

    sortSectionsByAvailability(courseSections, term)

    return {
      course,
      courseKey,
      sections: courseSections,
      highlightCrn,
    }
  }, [courses, sections, selectedCrn, selectedCourseKey, term])

  // Convert fetched course data to dialog format
  const dialogDataFromFetch = useMemo(() => {
    if (!fetchedCourse?.course) return null

    const course: GenerateCourseInfo = {
      subject: fetchedCourse.course.subject,
      courseNumber: fetchedCourse.course.courseNumber,
      title: fetchedCourse.course.title,
      credits: fetchedCourse.course.credits,
      gpa: fetchedCourse.course.gpa,
      passRate: fetchedCourse.course.passRate,
    }

    // Meeting times not available from this endpoint — fetched per-CRN on expand
    const courseSections: HydratedSection[] = fetchedCourse.sections.map(
      (s) => ({
        crn: s.crn,
        term: term ?? "",
        subject: course.subject,
        courseNumber: course.courseNumber,
        title: course.title,
        credits: course.credits,
        instructor: s.instructor,
        meetingTimes: [],
        enrollment: s.enrollment,
        maxEnrollment: s.maxEnrollment,
        seatsAvailable: s.seatsAvailable,
        waitCount: s.waitCount,
        isOpen: s.isOpen,
        gpa: s.gpa,
        gpaSource: s.gpaSource,
        passRate: s.passRate,
      })
    )

    sortSectionsByAvailability(courseSections, term)

    return {
      course,
      courseKey: `${course.subject}:${course.courseNumber}`,
      sections: courseSections,
      highlightCrn: selectedCrn ?? null,
    }
  }, [fetchedCourse, term, selectedCrn])

  // Use generateResult data if available, otherwise use fetched data
  const dialogData = dialogDataFromGenerate ?? dialogDataFromFetch

  // Loading state
  if (open && isLoading) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent
          className="sm:max-w-lg"
          aria-describedby={undefined}
        >
          <DialogTitle className="sr-only">Loading course details</DialogTitle>
          <div className="flex items-center justify-center py-8">
            <Loader2 className="size-6 animate-spin text-muted-foreground" />
          </div>
        </DialogContent>
      </Dialog>
    )
  }

  // Show "not found" message when dialog is open but no data available
  if (!dialogData) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent
          className="sm:max-w-lg"
          aria-describedby={undefined}
        >
          <DialogTitle className="sr-only">Course not found</DialogTitle>
          <div className="py-8 text-center text-muted-foreground">
            Course not found
          </div>
        </DialogContent>
      </Dialog>
    )
  }

  const { course, sections: courseSections, highlightCrn } = dialogData

  // Use key to reset inner component state when selection changes
  const dialogKey = `${selectedCrn ?? ""}-${selectedCourseKey ?? ""}`

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="max-h-[85vh] overflow-hidden sm:max-w-lg"
        aria-describedby={undefined}
      >
        <DialogContentInner
          key={dialogKey}
          course={course}
          courseSections={courseSections}
          highlightCrn={highlightCrn}
          term={term ?? ""}
          onAddSection={onAddSection}
          isSectionAdded={isSectionAdded}
        />
      </DialogContent>
    </Dialog>
  )
}
