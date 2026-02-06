import { useState, useCallback, useMemo, useEffect } from "react"
import { RotateCcw, SlidersHorizontal } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
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
import { SubjectCombobox } from "@/components/SubjectCombobox"
import {
  CourseListItem,
  type CourseListItemData,
} from "@/components/CourseListItem"
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
import { cn, genId } from "@/lib/utils"
import { useDebouncedState } from "@/hooks/use-debounced-state"

const SCOPE_OPTIONS: { value: SearchScope; label: string }[] = [
  { value: "term", label: "Term" },
  { value: "year", label: "Year" },
  { value: "all", label: "All" },
]

export function SearchView() {
  // Zustand selectors — only re-render when these specific slices change
  const searchFilters = useAppStore((s) => s.searchFilters)
  const setSearchFilters = useAppStore((s) => s.setSearchFilters)
  const clearSearchFilters = useAppStore((s) => s.clearSearchFilters)
  const searchResult = useAppStore((s) => s.searchResult)
  const setSearchResult = useAppStore((s) => s.setSearchResult)
  const addSlot = useAppStore((s) => s.addSlot)
  const slots = useAppStore((s) => s.slots)
  const openCourseDialog = useAppStore((s) => s.openCourseDialog)
  const closeCourseDialog = useAppStore((s) => s.closeCourseDialog)

  const { data: termsData } = useTerms()
  const termForSubjects = searchFilters.term || termsData?.current || ""
  const { data: subjectsData } = useSubjects(termForSubjects)

  const academicYears = useMemo(() => {
    if (!termsData?.terms) return []
    return getAcademicYearsFromTerms(termsData.terms.map((t) => t.code))
  }, [termsData])

  // Debounced text inputs — local state for instant typing, syncs to store after 200ms
  const setCourseNumber = useCallback(
    (v: string) => setSearchFilters({ courseNumber: v }),
    [setSearchFilters]
  )
  const setTitle = useCallback(
    (v: string) => setSearchFilters({ title: v }),
    [setSearchFilters]
  )
  const setInstructor = useCallback(
    (v: string) => setSearchFilters({ instructor: v }),
    [setSearchFilters]
  )

  const [localCourseNumber, setLocalCourseNumber, flushCourseNumber] =
    useDebouncedState(searchFilters.courseNumber, setCourseNumber)
  const [localTitle, setLocalTitle, flushTitle] =
    useDebouncedState(searchFilters.title, setTitle)
  const [localInstructor, setLocalInstructor, flushInstructor] =
    useDebouncedState(searchFilters.instructor, setInstructor)

  // Only pass request to useSearch when user clicks Search (not on every filter change)
  const [submittedRequest, setSubmittedRequest] = useState<ReturnType<
    typeof filtersToSearchRequest
  > | null>(null)
  const {
    data: searchData,
    isFetching,
    error: searchError,
  } = useSearch(submittedRequest ?? {})
  const [hasSearched, setHasSearched] = useState(false)

  // Sync search results to store when data arrives
  useEffect(() => {
    if (searchData) {
      setSearchResult(searchData)
    }
  }, [searchData, setSearchResult])

  // Show error toast when search fails
  useEffect(() => {
    if (searchError) {
      toast.error("Search failed. Please try again.")
    }
  }, [searchError])

  // Stale detection uses local values for instant feedback
  const currentFiltersWithLocal = useMemo(
    () => ({
      ...searchFilters,
      courseNumber: localCourseNumber,
      title: localTitle,
      instructor: localInstructor,
    }),
    [searchFilters, localCourseNumber, localTitle, localInstructor]
  )

  const hasSearchResult = !!searchResult
  const isSearchStale = useMemo(() => {
    if (!hasSearchResult || !submittedRequest) return false
    const currentRequest = filtersToSearchRequest(currentFiltersWithLocal)
    return JSON.stringify(currentRequest) !== JSON.stringify(submittedRequest)
  }, [currentFiltersWithLocal, submittedRequest, hasSearchResult])

  const handleSearch = useCallback(() => {
    // Flush debounced values to store before building the request
    flushCourseNumber()
    flushTitle()
    flushInstructor()
    setHasSearched(true)
    setSubmittedRequest(filtersToSearchRequest(currentFiltersWithLocal))
  }, [currentFiltersWithLocal, flushCourseNumber, flushTitle, flushInstructor])

  const handleClear = useCallback(() => {
    clearSearchFilters()
    setSubmittedRequest(null)
    setHasSearched(false)
  }, [clearSearchFilters])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Enter") handleSearch()
    },
    [handleSearch]
  )

  const handleCourseClick = useCallback(
    (courseKey: string) => openCourseDialog({ courseKey }),
    [openCourseDialog]
  )

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
        sections: null,
      })
      toast.success(
        `Added ${course.subject} ${course.courseNumber} to schedule`
      )
      closeCourseDialog()
    },
    [searchResult, addSlot, closeCourseDialog]
  )

  const terms = termsData?.terms ?? []
  const currentTerm = termsData?.current
  const subjects = subjectsData?.subjects ?? []

  const isCourseAdded = useCallback(
    (subject: string, courseNumber: string) => {
      return slots.some(
        (s) => s.subject === subject && s.courseNumber === courseNumber
      )
    },
    [slots]
  )

  // Badge uses local values for instant feedback
  const advancedFilterCount = [
    localTitle,
    localInstructor,
    searchFilters.minCredits != null,
    searchFilters.maxCredits != null,
    searchFilters.openSeats,
  ].filter(Boolean).length

  return (
    <div className="flex h-full flex-col">
      {/* Single scrollable container — filters scroll with results on mobile */}
      <div className="scrollbar-styled flex-1 overflow-y-auto">
        <div className="p-3">
          <div className="flex flex-col gap-2 xl:flex-row xl:items-end">
            {/* Filter inputs — equal-width columns via grid, takes 2/3 width on xl */}
            <div className="grid grid-cols-2 gap-2 sm:grid-cols-4 xl:flex-2">
            {/* Scope */}
            <Select
              value={searchFilters.scope}
              onValueChange={(v) =>
                setSearchFilters({ scope: v as SearchScope })
              }
            >
              <SelectTrigger className="w-full" aria-label="Search scope">
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

            {/* Term/Year/All */}
            <ScopeSelector
              scope={searchFilters.scope}
              term={searchFilters.term || currentTerm || ""}
              year={searchFilters.year}
              terms={terms}
              academicYears={academicYears}
              onTermChange={(v) => setSearchFilters({ term: v })}
              onYearChange={(v) => setSearchFilters({ year: parseInt(v, 10) })}
            />

            {/* Subject */}
            <SubjectCombobox
              subjects={subjects}
              value={searchFilters.subject}
              onChange={(v) => setSearchFilters({ subject: v })}
              placeholder="Subject"
              showAnyOption
            />

            {/* Course # */}
            <Input
              placeholder="Course number"
              aria-label="Course number"
              value={localCourseNumber}
              onChange={(e) => setLocalCourseNumber(e.target.value)}
              onKeyDown={handleKeyDown}
            />
          </div>

          {/* Buttons */}
          <div className="flex items-end gap-2 xl:flex-1">
            <Button
              onClick={handleSearch}
              disabled={isFetching}
              className={cn(
                "flex-1",
                isSearchStale &&
                  "ring-offset-background ring-2 ring-amber-500 ring-offset-2"
              )}
            >
              {isFetching ? "Searching..." : "Search"}
            </Button>

            {/* Advanced filters popover */}
            <Popover>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  size="icon"
                  title="More filters"
                  aria-label="More filters"
                  className="relative shrink-0"
                >
                  <SlidersHorizontal className="size-4" />
                  {advancedFilterCount > 0 && (
                    <span className="bg-primary text-primary-foreground absolute -top-1.5 -right-1.5 flex size-4 items-center justify-center rounded-full text-[10px] font-medium">
                      {advancedFilterCount}
                    </span>
                  )}
                </Button>
              </PopoverTrigger>
              <PopoverContent align="end" className="w-72 space-y-3">
                <div className="space-y-1.5">
                  <Label>Title</Label>
                  <Input
                    placeholder="e.g. Creative Seminar"
                    value={localTitle}
                    onChange={(e) => setLocalTitle(e.target.value)}
                    onKeyDown={handleKeyDown}
                  />
                </div>
                <div className="space-y-1.5">
                  <Label>Instructor</Label>
                  <Input
                    placeholder="e.g. Smith"
                    value={localInstructor}
                    onChange={(e) => setLocalInstructor(e.target.value)}
                    onKeyDown={handleKeyDown}
                  />
                </div>
                <div className="space-y-1.5">
                  <Label>Credits</Label>
                  <div className="flex items-center gap-1">
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
                      min={0}
                      max={36}
                    />
                  </div>
                </div>
                <div className="flex items-center justify-between">
                  <Label>Open Seats Only</Label>
                  <Switch
                    checked={searchFilters.openSeats}
                    onCheckedChange={(checked) =>
                      setSearchFilters({ openSeats: checked })
                    }
                  />
                </div>
              </PopoverContent>
            </Popover>

            <Button
              variant="outline"
              size="icon"
              onClick={handleClear}
              title="Clear all filters"
              aria-label="Clear all filters"
              className="shrink-0"
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
      </div>

      {/* Stats footer — stays visible at bottom */}
      {searchResult && searchResult.results.length > 0 && (
        <div className="text-muted-foreground border-t px-4 py-2 text-xs">
          Found {searchResult.total} courses ({searchResult.stats.totalSections}{" "}
          sections) in {searchResult.stats.timeMs.toFixed(1)}ms
          {searchResult.warning && (
            <span className="ml-2 text-amber-600 dark:text-amber-400">
              — {searchResult.warning}
            </span>
          )}
        </div>
      )}
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
  )
}

/** Term/Year/All selector — extracted to avoid nested ternary */
function ScopeSelector({
  scope,
  term,
  year,
  terms,
  academicYears,
  onTermChange,
  onYearChange,
}: {
  scope: SearchScope
  term: string
  year: number | null
  terms: { code: string; name: string }[]
  academicYears: number[]
  onTermChange: (value: string) => void
  onYearChange: (value: string) => void
}) {
  if (scope === "term") {
    return (
      <Select value={term} onValueChange={onTermChange}>
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
    )
  }

  if (scope === "year") {
    return (
      <Select value={year?.toString() ?? ""} onValueChange={onYearChange}>
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
    )
  }

  return (
    <div className="text-muted-foreground flex h-9 w-full items-center rounded-md border px-3 text-sm">
      All Time
    </div>
  )
}
