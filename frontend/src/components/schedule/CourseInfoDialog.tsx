import { useMemo, useState, useCallback } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { BookOpen, Users, Loader2 } from "lucide-react"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { SectionCard } from "./SectionCard"
import { useCourse, useCRN } from "@/hooks/use-api"
import { getCRN } from "@/lib/api"
import type {
  HydratedSection,
  GenerateCourseInfo,
  GenerateSectionInfo,
  CRNResponse,
} from "@/lib/api"
import { hydrateSection, sortSectionsByAvailability } from "@/lib/schedule-utils"
import { decodeHtmlEntities } from "@/lib/utils"

interface CourseInfoDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  // Course data from generate response (optional - will fetch if not provided)
  courses?: Record<string, GenerateCourseInfo>
  sections?: Record<string, GenerateSectionInfo>
  // Selection - either CRN or courseKey
  selectedCrn?: string
  selectedCourseKey?: string // "SUBJECT:NUMBER" format
  // Term for fetching data when generate result isn't available
  term?: string
}

interface DialogContentInnerProps {
  course: GenerateCourseInfo
  courseSections: HydratedSection[]
  highlightCrn: string | null
  term: string
}

function DialogContentInner({
  course,
  courseSections,
  highlightCrn,
  term,
}: DialogContentInnerProps) {
  const queryClient = useQueryClient()

  // Initialize expanded section to the highlighted one
  const [expandedSection, setExpandedSection] = useState<string | null>(
    highlightCrn
  )

  // Find the expanded section to check if it needs meeting times
  const expandedSectionData = expandedSection
    ? courseSections.find((s) => s.crn === expandedSection)
    : null

  // Check if we already have cached data for the expanded section
  const cachedExpandedData = expandedSection
    ? queryClient.getQueryData<CRNResponse>(["crn", expandedSection, term])
    : null

  const needsFetch =
    expandedSection &&
    expandedSectionData &&
    expandedSectionData.meetingTimes.length === 0 &&
    !cachedExpandedData?.section?.meetingTimes

  // Fetch meeting times for expanded section if needed
  const { isFetching } = useCRN(
    needsFetch ? expandedSection : "",
    needsFetch ? term : undefined
  )

  // Prefetch section details on hover or focus
  const handlePrefetch = useCallback(
    (crn: string) => {
      const section = courseSections.find((s) => s.crn === crn)
      const cached = queryClient.getQueryData<CRNResponse>(["crn", crn, term])
      // Only prefetch if section doesn't have meeting times and not already cached
      if (section?.meetingTimes.length === 0 && !cached && term) {
        queryClient.prefetchQuery({
          queryKey: ["crn", crn, term],
          queryFn: () => getCRN(crn, term),
          staleTime: 5 * 60 * 1000, // 5 minutes
        })
      }
    },
    [courseSections, term, queryClient]
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
      term,
    ])
    if (cached?.section?.meetingTimes) {
      return { ...section, meetingTimes: cached.section.meetingTimes }
    }
    return section
  })

  const totalSections = courseSections.length

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
      </div>

      {/* Section list */}
      <div className="-mx-2 max-h-[50vh] space-y-2 overflow-y-auto px-2 py-1">
        {sectionsWithMeetings.map((section) => (
          <SectionCard
            key={section.crn}
            section={section}
            expanded={expandedSection === section.crn}
            onToggleExpand={() =>
              setExpandedSection((prev) =>
                prev === section.crn ? null : section.crn
              )
            }
            highlighted={section.crn === highlightCrn}
            isLoadingDetails={
              expandedSection === section.crn &&
              isFetching &&
              section.meetingTimes.length === 0
            }
            onPrefetch={() => handlePrefetch(section.crn)}
          />
        ))}
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

    sortSectionsByAvailability(courseSections)

    return {
      course,
      courseKey,
      sections: courseSections,
      highlightCrn,
    }
  }, [courses, sections, selectedCrn, selectedCourseKey])

  // Convert fetched course data to dialog format
  const dialogDataFromFetch = useMemo(() => {
    if (!fetchedCourse?.course) return null

    const course: GenerateCourseInfo = {
      subject: fetchedCourse.course.subject,
      courseNumber: fetchedCourse.course.courseNumber,
      title: fetchedCourse.course.title,
      credits: fetchedCourse.course.credits,
    }

    // Convert sections - note: no meeting times available from this endpoint
    const courseSections: HydratedSection[] = fetchedCourse.sections.map(
      (s) => ({
        crn: s.crn,
        term: term ?? "",
        subject: course.subject,
        courseNumber: course.courseNumber,
        title: course.title,
        credits: course.credits,
        instructor: s.instructor,
        meetingTimes: [], // Not available from getCourse endpoint
        enrollment: s.enrollment,
        maxEnrollment: s.maxEnrollment,
        seatsAvailable: s.seatsAvailable,
        waitCount: s.waitCount,
        isOpen: s.isOpen,
      })
    )

    sortSectionsByAvailability(courseSections)

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
        />
      </DialogContent>
    </Dialog>
  )
}
