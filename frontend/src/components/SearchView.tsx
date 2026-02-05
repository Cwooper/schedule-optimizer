import { useState, useCallback, useMemo } from "react"
import { RotateCcw } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { SubjectCombobox } from "@/components/SubjectCombobox"
import { CourseListItem, type CourseListItemData } from "@/components/CourseListItem"
import { CourseInfoDialog } from "@/components/schedule/CourseInfoDialog"
import {
  useAppStore,
  filtersToSearchRequest,
  type SearchScope,
} from "@/stores/app-store"
import { useTerms, useSubjects, useSearch } from "@/hooks/use-api"
import type { SearchResponse } from "@/lib/api"
import {
  getAcademicYearsFromTerms,
  formatAcademicYear,
} from "@/lib/schedule-utils"
import { genId } from "@/lib/utils"

const SCOPE_OPTIONS: { value: SearchScope; label: string }[] = [
  { value: "term", label: "Term" },
  { value: "year", label: "Year" },
  { value: "all", label: "All" },
]

export function SearchView() {
  const {
    searchFilters,
    setSearchFilters,
    clearSearchFilters,
    searchResult,
    setSearchResult,
    addSlot,
    slots,
    term: scheduleTerm,
  } = useAppStore()

  const { data: termsData } = useTerms()
  const termForSubjects = searchFilters.term || termsData?.current || ""
  const { data: subjectsData } = useSubjects(termForSubjects)

  const academicYears = useMemo(() => {
    if (!termsData?.terms) return []
    return getAcademicYearsFromTerms(termsData.terms.map((t) => t.code))
  }, [termsData])

  const searchRequest = useMemo(
    () => filtersToSearchRequest(searchFilters),
    [searchFilters]
  )

  const { isFetching, refetch } = useSearch(searchRequest)
  const [hasSearched, setHasSearched] = useState(false)

  // Dialog state
  const [dialogOpen, setDialogOpen] = useState(false)
  const [selectedCourseKey, setSelectedCourseKey] = useState<string | undefined>()

  const handleSearch = useCallback(() => {
    setHasSearched(true)
    refetch()
      .then(({ data }) => {
        if (data) setSearchResult(data)
      })
      .catch(() => {
        toast.error("Search failed. Please try again.")
      })
  }, [refetch, setSearchResult])

  const handleClear = useCallback(() => {
    clearSearchFilters()
    setHasSearched(false)
  }, [clearSearchFilters])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Enter") handleSearch()
    },
    [handleSearch]
  )

  // Open course detail dialog
  const handleCourseClick = useCallback((courseKey: string) => {
    setSelectedCourseKey(courseKey)
    setDialogOpen(true)
  }, [])

  // Add course to schedule (all sections)
  const handleAddCourse = useCallback(
    (courseKey: string) => {
      if (!searchResult) return
      const course = searchResult.courses[courseKey]
      if (!course) return

      addSlot({
        id: genId(),
        subject: course.subject,
        courseNumber: course.courseNumber,
        displayName: `${course.subject} ${course.courseNumber}`,
        title: course.title,
        required: true,
        sections: null, // All sections allowed
      })
      toast.success(`Added ${course.subject} ${course.courseNumber} to schedule`)
      setDialogOpen(false)
    },
    [searchResult, addSlot]
  )

  // Add specific section to schedule
  const handleAddSection = useCallback(
    (crn: string, term: string) => {
      if (!searchResult) return
      const sectionKey = `${term}:${crn}`
      const section = searchResult.sections[sectionKey]
      if (!section) return

      const course = searchResult.courses[section.courseKey]
      if (!course) return

      addSlot({
        id: genId(),
        subject: course.subject,
        courseNumber: course.courseNumber,
        displayName: `${course.subject} ${course.courseNumber}`,
        title: course.title,
        required: false, // Not forcing the course, just filtering to this CRN
        sections: [
          {
            crn,
            term,
            instructor: section.instructor ?? null,
            required: false, // Just a filter, not pinned
          },
        ],
      })
      toast.success(`Added ${course.subject} ${course.courseNumber} (CRN: ${crn}) to schedule`)
    },
    [searchResult, addSlot]
  )

  const terms = termsData?.terms ?? []
  const currentTerm = termsData?.current
  const subjects = subjectsData?.subjects ?? []

  // Check if a course is already in the schedule
  const isCourseAdded = useCallback(
    (subject: string, courseNumber: string) => {
      return slots.some(
        (s) => s.subject === subject && s.courseNumber === courseNumber
      )
    },
    [slots]
  )

  // Check if a specific section (CRN) is already in the schedule
  const isSectionAdded = useCallback(
    (crn: string) => {
      return slots.some(
        (s) => s.sections?.some((sec) => sec.crn === crn)
      )
    },
    [slots]
  )

  // Determine term for dialog (use search filter term, or schedule term, or current)
  const dialogTerm = searchFilters.term || scheduleTerm || currentTerm || ""

  return (
    <div className="scrollbar-styled h-full overflow-auto">
      {/* Filters */}
      <div className="border-b p-4">
        <div className="grid grid-cols-3 items-end gap-3 sm:grid-cols-4 lg:grid-cols-10 xl:grid-cols-12">
          {/* Scope */}
          <div className="col-span-3 space-y-1.5 sm:col-span-2 xl:col-span-2">
            <Label>Scope</Label>
            <Select
              value={searchFilters.scope}
              onValueChange={(v) =>
                setSearchFilters({ scope: v as SearchScope })
              }
            >
              <SelectTrigger className="w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {SCOPE_OPTIONS.map((opt) => (
                  <SelectItem key={opt.value} value={opt.value}>
                    {opt.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Term/Year */}
          <div className="col-span-3 space-y-1.5 sm:col-span-2 lg:col-span-3 xl:col-span-2">
            <Label>
              {searchFilters.scope === "term"
                ? "Term"
                : searchFilters.scope === "year"
                  ? "Year"
                  : "Range"}
            </Label>
            {searchFilters.scope === "term" ? (
              <Select
                value={searchFilters.term || currentTerm || ""}
                onValueChange={(v) => setSearchFilters({ term: v })}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Select term" />
                </SelectTrigger>
                <SelectContent>
                  {terms.map((t) => (
                    <SelectItem key={t.code} value={t.code}>
                      {t.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            ) : searchFilters.scope === "year" ? (
              <Select
                value={searchFilters.year?.toString() ?? ""}
                onValueChange={(v) =>
                  setSearchFilters({ year: parseInt(v, 10) })
                }
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Select year" />
                </SelectTrigger>
                <SelectContent>
                  {academicYears.map((y) => (
                    <SelectItem key={y} value={y.toString()}>
                      {formatAcademicYear(y)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            ) : (
              <div className="text-muted-foreground flex h-9 items-center rounded-md border px-3 text-sm">
                All Time
              </div>
            )}
          </div>

          {/* Subject */}
          <div className="col-span-2 space-y-1.5 sm:col-span-2 lg:col-span-3 xl:col-span-3">
            <Label>Subject</Label>
            <SubjectCombobox
              subjects={subjects}
              value={searchFilters.subject}
              onChange={(v) => setSearchFilters({ subject: v })}
              placeholder="Any"
              showAnyOption
            />
          </div>

          {/* Course # */}
          <div className="col-span-1 space-y-1.5 sm:col-span-2 xl:col-span-2">
            <Label>Course #</Label>
            <Input
              placeholder="241"
              value={searchFilters.courseNumber}
              onChange={(e) =>
                setSearchFilters({ courseNumber: e.target.value })
              }
              onKeyDown={handleKeyDown}
            />
          </div>

          {/* Credits + Open */}
          <div className="col-span-3 space-y-1.5 sm:col-span-4 lg:col-span-4 xl:col-span-3">
            <div className="flex items-center">
              <Label className="flex-1">Credits</Label>
              <div className="flex justify-center px-2">
                <Label>Open</Label>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <div className="flex flex-1 items-center gap-1">
                <Input
                  type="number"
                  placeholder="Min"
                  value={searchFilters.minCredits ?? ""}
                  onChange={(e) =>
                    setSearchFilters({
                      minCredits: e.target.value
                        ? parseInt(e.target.value, 10)
                        : null,
                    })
                  }
                  onKeyDown={handleKeyDown}
                  className="w-full"
                  min={0}
                  max={36}
                />
                <span className="text-muted-foreground">-</span>
                <Input
                  type="number"
                  placeholder="Max"
                  value={searchFilters.maxCredits ?? ""}
                  onChange={(e) =>
                    setSearchFilters({
                      maxCredits: e.target.value
                        ? parseInt(e.target.value, 10)
                        : null,
                    })
                  }
                  onKeyDown={handleKeyDown}
                  className="w-full"
                  min={0}
                  max={36}
                />
              </div>
              <div className="flex justify-center px-2">
                <Switch
                  checked={searchFilters.openSeats}
                  onCheckedChange={(checked) =>
                    setSearchFilters({ openSeats: checked })
                  }
                />
              </div>
            </div>
          </div>

          {/* Title */}
          <div className="col-span-3 space-y-1.5 sm:col-span-2 lg:col-span-3 xl:col-span-4">
            <Label>Title</Label>
            <Input
              placeholder="Creative Seminar"
              value={searchFilters.title}
              onChange={(e) => setSearchFilters({ title: e.target.value })}
              onKeyDown={handleKeyDown}
            />
          </div>

          {/* Instructor */}
          <div className="col-span-3 space-y-1.5 sm:col-span-2 lg:col-span-3 xl:col-span-4">
            <Label>Instructor</Label>
            <Input
              placeholder="e.g. Smith"
              value={searchFilters.instructor}
              onChange={(e) => setSearchFilters({ instructor: e.target.value })}
              onKeyDown={handleKeyDown}
            />
          </div>

          {/* Actions */}
          <div className="col-span-3 flex items-end gap-2 sm:col-span-4 lg:col-span-10 xl:col-span-4">
            <Button
              onClick={handleSearch}
              disabled={isFetching}
              className="flex-1"
            >
              {isFetching ? "Searching..." : "Search"}
            </Button>
            <Button
              variant="outline"
              size="icon"
              onClick={handleClear}
              title="Clear all filters"
            >
              <RotateCcw className="size-4" />
            </Button>
          </div>
        </div>
      </div>

      {/* Results */}
      <div className="p-4">
        <SearchResults
          hasSearched={hasSearched}
          isFetching={isFetching}
          searchResult={searchResult}
          onCourseClick={handleCourseClick}
          onAddCourse={handleAddCourse}
          isCourseAdded={isCourseAdded}
        />
      </div>

      {/* Course Detail Dialog */}
      <CourseInfoDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        courses={searchResult?.courses}
        sections={searchResult?.sections}
        selectedCourseKey={selectedCourseKey}
        term={dialogTerm}
        onAddSection={handleAddSection}
        isSectionAdded={isSectionAdded}
      />
    </div>
  )
}

interface SearchResultsProps {
  hasSearched: boolean
  isFetching: boolean
  searchResult: SearchResponse | null
  onCourseClick: (courseKey: string) => void
  onAddCourse: (courseKey: string) => void
  /** Check if a course is already in the schedule */
  isCourseAdded: (subject: string, courseNumber: string) => boolean
}

/** Convert search result to CourseListItem data format */
function toListItemData(
  courseKey: string,
  searchResult: SearchResponse
): CourseListItemData {
  const course = searchResult.courses[courseKey]
  const sections = Object.entries(searchResult.sections)
    .filter(([, s]) => s.courseKey === courseKey)
    .map(([, s]) => ({
      crn: s.crn,
      term: s.term,
      instructor: s.instructor,
      seatsAvailable: s.seatsAvailable,
      isOpen: s.isOpen,
    }))

  return {
    subject: course.subject,
    courseNumber: course.courseNumber,
    title: course.title,
    credits: course.credits,
    creditsHigh: course.creditsHigh,
    sections,
  }
}

function SearchResults({
  hasSearched,
  isFetching,
  searchResult,
  onCourseClick,
  onAddCourse,
  isCourseAdded,
}: SearchResultsProps) {
  if (!hasSearched && !searchResult) {
    return (
      <p className="text-muted-foreground py-8 text-center text-sm">
        Enter search criteria and click Search
      </p>
    )
  }

  if (isFetching && !searchResult) {
    return (
      <p className="text-muted-foreground py-8 text-center text-sm">
        Searching...
      </p>
    )
  }

  if (!searchResult || searchResult.results.length === 0) {
    return (
      <p className="text-muted-foreground py-8 text-center text-sm">
        No courses found
      </p>
    )
  }

  return (
    <div className="space-y-1">
      <p className="text-muted-foreground mb-3 text-xs">
        Found {searchResult.total} courses ({searchResult.stats.totalSections}{" "}
        sections) in {searchResult.stats.timeMs.toFixed(1)}ms
        {searchResult.warning && (
          <span className="ml-2 text-amber-600 dark:text-amber-400">
            — {searchResult.warning}
          </span>
        )}
      </p>

      <div className="space-y-2">
        {searchResult.results.map((ref) => {
          const course = searchResult.courses[ref.courseKey]
          if (!course) return null
          return (
            <CourseListItem
              key={ref.courseKey}
              course={toListItemData(ref.courseKey, searchResult)}
              onClick={() => onCourseClick(ref.courseKey)}
              onAdd={() => onAddCourse(ref.courseKey)}
              isAdded={isCourseAdded(course.subject, course.courseNumber)}
            />
          )
        })}
      </div>
    </div>
  )
}
