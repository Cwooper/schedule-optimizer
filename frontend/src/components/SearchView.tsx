import { useState, useCallback, useMemo, useEffect } from "react"
import { RotateCcw, SlidersHorizontal } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
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
import { DebouncedInput } from "@/components/DebouncedInput"
import { SearchResults } from "@/components/SearchResults"
import {
  useAppStore,
  filtersToSearchRequest,
  type SearchScope,
} from "@/stores/app-store"
import { useTerms, useSubjects, useSearch } from "@/hooks/use-api"
import {
  getAcademicYearsFromTerms,
  formatAcademicYear,
} from "@/lib/schedule-utils"
import { cn } from "@/lib/utils"

const EMPTY_SUBJECTS: { code: string; name: string }[] = []
const SCOPE_OPTIONS: { value: SearchScope; label: string }[] = [
  { value: "term", label: "Term" },
  { value: "year", label: "Year" },
  { value: "all", label: "All" },
]

export function SearchView() {
  const searchFilters = useAppStore((s) => s.searchFilters)
  const setSearchFilters = useAppStore((s) => s.setSearchFilters)
  const clearSearchFilters = useAppStore((s) => s.clearSearchFilters)
  const searchResult = useAppStore((s) => s.searchResult)
  const setSearchResult = useAppStore((s) => s.setSearchResult)
  const addCourseToSlot = useAppStore((s) => s.addCourseToSlot)
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

  // Stable callbacks for DebouncedInput onSync props
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
  const setSubject = useCallback(
    (v: string) => setSearchFilters({ subject: v }),
    [setSearchFilters]
  )

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

  useEffect(() => {
    if (searchData) {
      setSearchResult(searchData)
      if (searchData.warning) {
        toast.warning(searchData.warning)
      }
    }
  }, [searchData, setSearchResult])

  useEffect(() => {
    if (searchError) {
      toast.error("Search failed. Please try again.")
    }
  }, [searchError])

  // Stale detection — uses store values (200ms delay after typing is fine)
  const hasSearchResult = !!searchResult
  const isSearchStale = useMemo(() => {
    if (!hasSearchResult || !submittedRequest) return false
    const currentRequest = filtersToSearchRequest(searchFilters)
    return JSON.stringify(currentRequest) !== JSON.stringify(submittedRequest)
  }, [searchFilters, submittedRequest, hasSearchResult])

  // Read fresh values from store directly — DebouncedInput flushes on blur/Enter
  // before this runs, and Zustand's set() is synchronous
  const handleSearch = useCallback(() => {
    const filters = useAppStore.getState().searchFilters
    setHasSearched(true)
    setSubmittedRequest(filtersToSearchRequest(filters))
  }, [])

  const handleClear = useCallback(() => {
    clearSearchFilters()
    setSubmittedRequest(null)
    setHasSearched(false)
  }, [clearSearchFilters])

  const handleCourseClick = useCallback(
    (courseKey: string) => openCourseDialog({ courseKey, source: "search" }),
    [openCourseDialog]
  )

  const handleAddCourse = useCallback(
    (courseKey: string) => {
      if (!searchResult) return
      const course = searchResult.courses[courseKey]
      if (!course) return

      const result = addCourseToSlot(course.subject, course.courseNumber, course.title)
      if (result === "added") {
        toast.success(
          `Added ${course.subject} ${course.courseNumber} to schedule`
        )
      } else if (result === "updated") {
        toast.success(
          `Updated ${course.subject} ${course.courseNumber} in schedule`
        )
      }
      closeCourseDialog()
    },
    [searchResult, addCourseToSlot, closeCourseDialog]
  )

  const terms = termsData?.terms ?? []
  const currentTerm = termsData?.current
  const subjects = subjectsData?.subjects ?? EMPTY_SUBJECTS

  const isCourseAdded = useCallback(
    (subject: string, courseNumber: string) => {
      return slots.some(
        (s) => s.subject === subject && s.courseNumber === courseNumber
      )
    },
    [slots]
  )

  const advancedFilterCount = [
    searchFilters.title,
    searchFilters.instructor,
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
            {/* Filter inputs */}
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
                onYearChange={(v) =>
                  setSearchFilters({ year: parseInt(v, 10) })
                }
              />

              {/* Subject */}
              <SubjectCombobox
                subjects={subjects}
                value={searchFilters.subject}
                onChange={setSubject}
                placeholder="Subject"
                showAnyOption
              />

              {/* Course # */}
              <DebouncedInput
                placeholder="Course number"
                aria-label="Course number"
                storeValue={searchFilters.courseNumber}
                onSync={setCourseNumber}
                onSubmit={handleSearch}
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
                    <DebouncedInput
                      placeholder="e.g. Creative Seminar"
                      storeValue={searchFilters.title}
                      onSync={setTitle}
                      onSubmit={handleSearch}
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label>Instructor</Label>
                    <DebouncedInput
                      placeholder="e.g. Smith"
                      storeValue={searchFilters.instructor}
                      onSync={setInstructor}
                      onSubmit={handleSearch}
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label>Credits</Label>
                    <div className="flex items-center gap-1">
                      <DebouncedInput
                        type="number"
                        placeholder="Min"
                        storeValue={
                          searchFilters.minCredits?.toString() ?? ""
                        }
                        onSync={(v) =>
                          setSearchFilters({
                            minCredits: v ? parseInt(v, 10) : null,
                          })
                        }
                        onSubmit={handleSearch}
                        min={0}
                        max={36}
                      />
                      <span className="text-muted-foreground">-</span>
                      <DebouncedInput
                        type="number"
                        placeholder="Max"
                        storeValue={
                          searchFilters.maxCredits?.toString() ?? ""
                        }
                        onSync={(v) =>
                          setSearchFilters({
                            maxCredits: v ? parseInt(v, 10) : null,
                          })
                        }
                        onSubmit={handleSearch}
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
          Found {searchResult.total} courses (
          {searchResult.stats.totalSections} sections) in{" "}
          {searchResult.stats.timeMs.toFixed(1)}ms
        </div>
      )}
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
