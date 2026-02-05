import { useRef, useState, useLayoutEffect, useCallback, useMemo } from "react"
import { motion } from "framer-motion"
import {
  RotateCcw,
  ChevronRight,
  ChevronDown,
  ChevronsUpDown,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
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
import { cn, decodeHtmlEntities } from "@/lib/utils"

const scopeValues = ["term", "year", "all"] as const

const scopeLabels: Record<SearchScope, string> = {
  term: "Term",
  year: "Year",
  all: "All",
}

export function SearchView() {
  const {
    searchFilters,
    setSearchFilters,
    clearSearchFilters,
    searchResult,
    setSearchResult,
    expandedSearchCourses,
    toggleSearchCourseExpanded,
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

  const handleSearch = useCallback(() => {
    setHasSearched(true)
    refetch().then(({ data }) => {
      if (data) setSearchResult(data)
    })
  }, [refetch, setSearchResult])

  const handleClear = useCallback(() => {
    clearSearchFilters()
    setHasSearched(false)
  }, [clearSearchFilters])

  return (
    <div className="flex h-full flex-col">
      <div className="border-b p-4">
        <SearchFiltersForm
          filters={searchFilters}
          onFiltersChange={setSearchFilters}
          terms={termsData?.terms ?? []}
          currentTerm={termsData?.current}
          subjects={subjectsData?.subjects ?? []}
          academicYears={academicYears}
          onSearch={handleSearch}
          onClear={handleClear}
          isSearching={isFetching}
        />
      </div>

      <div className="scrollbar-styled flex-1 overflow-auto p-4">
        <SearchResultsArea
          hasSearched={hasSearched}
          isFetching={isFetching}
          searchResult={searchResult}
          expandedCourses={expandedSearchCourses}
          onToggleExpand={toggleSearchCourseExpanded}
        />
      </div>
    </div>
  )
}

// ── Search Filters Form ─────────────────────────────────────────────────

interface SearchFiltersFormProps {
  filters: ReturnType<typeof useAppStore>["searchFilters"]
  onFiltersChange: (filters: Partial<typeof filters>) => void
  terms: { code: string; name: string }[]
  currentTerm?: string
  subjects: { code: string; name: string }[]
  academicYears: number[]
  onSearch: () => void
  onClear: () => void
  isSearching: boolean
}

function SearchFiltersForm({
  filters,
  onFiltersChange,
  terms,
  currentTerm,
  subjects,
  academicYears,
  onSearch,
  onClear,
  isSearching,
}: SearchFiltersFormProps) {
  // Scope selector animated indicator
  const scopeRefs = useRef<Record<SearchScope, HTMLButtonElement | null>>({
    term: null,
    year: null,
    all: null,
  })
  const [indicatorStyle, setIndicatorStyle] = useState({ left: 0, width: 0 })

  const updateIndicator = useCallback(() => {
    const el = scopeRefs.current[filters.scope]
    if (el) {
      setIndicatorStyle({ left: el.offsetLeft, width: el.offsetWidth })
    }
  }, [filters.scope])

  useLayoutEffect(() => {
    updateIndicator()
  }, [updateIndicator])

  useLayoutEffect(() => {
    const onResize = () => updateIndicator()
    window.addEventListener("resize", onResize)
    return () => window.removeEventListener("resize", onResize)
  }, [updateIndicator])

  const handleScopeChange = useCallback(
    (value: string) => {
      if (value && scopeValues.includes(value as SearchScope)) {
        onFiltersChange({ scope: value as SearchScope })
      }
    },
    [onFiltersChange]
  )

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Enter") onSearch()
    },
    [onSearch]
  )

  const [subjectOpen, setSubjectOpen] = useState(false)

  return (
    <div className="grid grid-cols-2 items-end gap-3 sm:grid-cols-4 lg:grid-cols-10">
      {/* Scope - 2 cols on mobile, 2 on tablet, 2 on desktop */}
      <div className="col-span-2 space-y-1.5">
        <Label>Scope</Label>
        <ToggleGroup
          type="single"
          value={filters.scope}
          onValueChange={handleScopeChange}
          className="bg-muted/50 relative w-full rounded-md p-0.5"
        >
          <motion.div
            className="bg-background absolute inset-y-0.5 rounded-sm shadow-sm"
            initial={false}
            animate={{ left: indicatorStyle.left, width: indicatorStyle.width }}
            transition={{ type: "spring", stiffness: 400, damping: 30 }}
          />
          {scopeValues.map((scope) => (
            <ToggleGroupItem
              key={scope}
              ref={(el) => {
                scopeRefs.current[scope] = el
              }}
              value={scope}
              className="relative z-10 flex-1 px-3 py-1.5 text-sm data-[state=on]:bg-transparent"
            >
              {scopeLabels[scope]}
            </ToggleGroupItem>
          ))}
        </ToggleGroup>
      </div>

      {/* Term/Year dropdown - 2 cols on mobile, 2 on tablet, 2 on desktop */}
      <div
        className={cn(
          "col-span-2 space-y-1.5",
          filters.scope === "all" && "invisible"
        )}
      >
        <Label>{filters.scope === "term" ? "Term" : "Year"}</Label>
        {filters.scope === "term" ? (
          <Select
            value={filters.term || currentTerm || ""}
            onValueChange={(v) => onFiltersChange({ term: v })}
          >
            <SelectTrigger>
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
        ) : (
          <Select
            value={filters.year?.toString() ?? ""}
            onValueChange={(v) => onFiltersChange({ year: parseInt(v, 10) })}
          >
            <SelectTrigger>
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
        )}
      </div>

      {/* Subject - 2 cols on mobile, 2 on tablet, 2 on desktop */}
      <div className="col-span-2 space-y-1.5">
        <Label>Subject</Label>
        <Popover open={subjectOpen} onOpenChange={setSubjectOpen}>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              role="combobox"
              aria-expanded={subjectOpen}
              aria-label={
                filters.subject
                  ? `Subject: ${filters.subject}`
                  : "Select subject"
              }
              className="w-full justify-between"
            >
              <span className="truncate">{filters.subject || "Any"}</span>
              <ChevronsUpDown className="size-4 shrink-0 opacity-50" />
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-64 p-0" align="start">
            <Command>
              <CommandInput placeholder="Search subjects..." />
              <CommandList>
                <CommandEmpty>No subject found.</CommandEmpty>
                <CommandGroup>
                  <CommandItem
                    value=""
                    onSelect={() => {
                      onFiltersChange({ subject: "" })
                      setSubjectOpen(false)
                    }}
                    className={cn(!filters.subject && "bg-accent")}
                  >
                    <span className="text-muted-foreground">Any subject</span>
                  </CommandItem>
                  {subjects.map((subject) => (
                    <CommandItem
                      key={subject.code}
                      value={`${subject.code} ${subject.name}`}
                      onSelect={() => {
                        onFiltersChange({ subject: subject.code })
                        setSubjectOpen(false)
                      }}
                      className={cn(
                        filters.subject === subject.code && "bg-accent"
                      )}
                    >
                      <span className="w-12 shrink-0 font-medium whitespace-nowrap">
                        {subject.code}
                      </span>
                      <span className="text-muted-foreground truncate">
                        {decodeHtmlEntities(subject.name)}
                      </span>
                    </CommandItem>
                  ))}
                </CommandGroup>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>
      </div>

      {/* Course # - 1 col on mobile, 1 on tablet, 1 on desktop */}
      <div className="col-span-1 space-y-1.5">
        <Label htmlFor="course-number">Course #</Label>
        <Input
          id="course-number"
          placeholder="241"
          value={filters.courseNumber}
          onChange={(e) => onFiltersChange({ courseNumber: e.target.value })}
          onKeyDown={handleKeyDown}
        />
      </div>

      {/* Credits - 1 col on mobile, 1 on tablet, 2 on desktop */}
      <div className="col-span-1 space-y-1.5 lg:col-span-2">
        <Label>Credits</Label>
        <div className="flex items-center gap-1">
          <Input
            type="number"
            placeholder="Min"
            value={filters.minCredits ?? ""}
            onChange={(e) =>
              onFiltersChange({
                minCredits: e.target.value
                  ? parseInt(e.target.value, 10)
                  : null,
              })
            }
            onKeyDown={handleKeyDown}
            className="w-full"
            min={0}
            max={20}
          />
          <span className="text-muted-foreground shrink-0">-</span>
          <Input
            type="number"
            placeholder="Max"
            value={filters.maxCredits ?? ""}
            onChange={(e) =>
              onFiltersChange({
                maxCredits: e.target.value
                  ? parseInt(e.target.value, 10)
                  : null,
              })
            }
            onKeyDown={handleKeyDown}
            className="w-full"
            min={0}
            max={20}
          />
        </div>
      </div>

      {/* Open toggle - hidden on mobile (moved to row 2), shown on tablet+, 1 col */}
      <div className="col-span-1 hidden space-y-1.5 sm:block">
        <Label htmlFor="open-seats">Open</Label>
        <div className="flex h-9 items-center">
          <Switch
            id="open-seats"
            checked={filters.openSeats}
            onCheckedChange={(checked) =>
              onFiltersChange({ openSeats: checked })
            }
          />
        </div>
      </div>

      {/* Title - 2 cols on mobile (full row), 2 on tablet, 3 on desktop */}
      <div className="col-span-2 space-y-1.5 sm:col-span-2 lg:col-span-3">
        <Label htmlFor="title">Title</Label>
        <Input
          id="title"
          placeholder="e.g. Data Structures"
          value={filters.title}
          onChange={(e) => onFiltersChange({ title: e.target.value })}
          onKeyDown={handleKeyDown}
        />
      </div>

      {/* Instructor - 2 cols on mobile, 2 on tablet, 3 on desktop */}
      <div className="col-span-2 space-y-1.5 sm:col-span-2 lg:col-span-3">
        <Label htmlFor="instructor">Instructor</Label>
        <Input
          id="instructor"
          placeholder="e.g. Smith"
          value={filters.instructor}
          onChange={(e) => onFiltersChange({ instructor: e.target.value })}
          onKeyDown={handleKeyDown}
        />
      </div>

      {/* Open toggle - mobile only */}
      <div className="col-span-1 space-y-1.5 sm:hidden">
        <Label htmlFor="open-seats-mobile">Open</Label>
        <div className="flex h-9 items-center">
          <Switch
            id="open-seats-mobile"
            checked={filters.openSeats}
            onCheckedChange={(checked) =>
              onFiltersChange({ openSeats: checked })
            }
          />
        </div>
      </div>

      {/* Actions - remaining space */}
      <div className="col-span-1 flex items-end gap-2 sm:col-span-2 lg:col-span-4">
        <Button onClick={onSearch} disabled={isSearching} className="flex-1">
          {isSearching ? "Searching..." : "Search"}
        </Button>
        <Button
          variant="outline"
          size="icon"
          onClick={onClear}
          title="Clear all filters"
        >
          <RotateCcw className="size-4" />
        </Button>
      </div>
    </div>
  )
}

// ── Search Results Area ─────────────────────────────────────────────────

interface SearchResultsAreaProps {
  hasSearched: boolean
  isFetching: boolean
  searchResult: ReturnType<typeof useAppStore>["searchResult"]
  expandedCourses: Set<string>
  onToggleExpand: (courseKey: string) => void
}

function SearchResultsArea({
  hasSearched,
  isFetching,
  searchResult,
  expandedCourses,
  onToggleExpand,
}: SearchResultsAreaProps) {
  if (!hasSearched && !searchResult) {
    return (
      <div className="flex h-full items-center justify-center">
        <p className="text-muted-foreground text-sm">
          Enter search criteria and click Search
        </p>
      </div>
    )
  }

  if (isFetching) {
    return (
      <div className="flex h-full items-center justify-center">
        <p className="text-muted-foreground text-sm">Searching...</p>
      </div>
    )
  }

  if (!searchResult || searchResult.results.length === 0) {
    return (
      <div className="flex h-full items-center justify-center">
        <p className="text-muted-foreground text-sm">No courses found</p>
      </div>
    )
  }

  return (
    <SearchResults
      result={searchResult}
      expandedCourses={expandedCourses}
      onToggleExpand={onToggleExpand}
    />
  )
}

// ── Search Results ──────────────────────────────────────────────────────

interface SearchResultsProps {
  result: NonNullable<ReturnType<typeof useAppStore>["searchResult"]>
  expandedCourses: Set<string>
  onToggleExpand: (courseKey: string) => void
}

function SearchResults({
  result,
  expandedCourses,
  onToggleExpand,
}: SearchResultsProps) {
  return (
    <div className="space-y-1">
      <p className="text-muted-foreground mb-3 text-xs">
        Found {result.total} courses ({result.stats.totalSections} sections) in{" "}
        {result.stats.timeMs.toFixed(1)}ms
        {result.warning && (
          <span className="ml-2 text-amber-600 dark:text-amber-400">
            — {result.warning}
          </span>
        )}
      </p>

      {result.results.map((courseRef) => (
        <CourseResultRow
          key={courseRef.courseKey}
          courseRef={courseRef}
          course={result.courses[courseRef.courseKey]}
          sections={result.sections}
          isExpanded={expandedCourses.has(courseRef.courseKey)}
          onToggle={() => onToggleExpand(courseRef.courseKey)}
        />
      ))}
    </div>
  )
}

// ── Course Result Row ───────────────────────────────────────────────────

interface CourseResultRowProps {
  courseRef: { courseKey: string; sectionKeys: string[] }
  course:
    | {
        subject: string
        courseNumber: string
        title: string
        credits: number
        creditsHigh?: number
      }
    | undefined
  sections: Record<
    string,
    {
      crn: string
      instructor?: string
      isOpen: boolean
      seatsAvailable: number
      maxEnrollment: number
      meetingTimes: { days: boolean[]; startTime: string; endTime: string }[]
    }
  >
  isExpanded: boolean
  onToggle: () => void
}

function CourseResultRow({
  courseRef,
  course,
  sections,
  isExpanded,
  onToggle,
}: CourseResultRowProps) {
  if (!course) return null

  const sectionCount = courseRef.sectionKeys.length
  const creditsDisplay =
    course.creditsHigh && course.creditsHigh !== course.credits
      ? `${course.credits}-${course.creditsHigh}`
      : String(course.credits)

  return (
    <div className="border-b last:border-b-0">
      <button
        onClick={onToggle}
        aria-expanded={isExpanded}
        className="hover:bg-muted/50 flex w-full items-center gap-2 py-2 text-left transition-colors"
      >
        {isExpanded ? (
          <ChevronDown className="text-muted-foreground size-4 flex-shrink-0" />
        ) : (
          <ChevronRight className="text-muted-foreground size-4 flex-shrink-0" />
        )}
        <span className="font-medium">
          {course.subject} {course.courseNumber}
        </span>
        <span className="text-muted-foreground">—</span>
        <span className="flex-1 truncate">{course.title}</span>
        <span className="text-muted-foreground flex-shrink-0 text-sm">
          {creditsDisplay} cr
        </span>
        <span className="text-muted-foreground w-20 flex-shrink-0 text-right text-sm">
          {sectionCount} section{sectionCount !== 1 ? "s" : ""}
        </span>
      </button>

      {isExpanded && (
        <div className="space-y-1 pb-2 pl-6">
          {courseRef.sectionKeys.map((sectionKey) => (
            <SectionRow key={sectionKey} section={sections[sectionKey]} />
          ))}
        </div>
      )}
    </div>
  )
}

// ── Section Row ─────────────────────────────────────────────────────────

interface SectionRowProps {
  section:
    | {
        crn: string
        instructor?: string
        isOpen: boolean
        seatsAvailable: number
        maxEnrollment: number
        meetingTimes: { days: boolean[]; startTime: string; endTime: string }[]
      }
    | undefined
}

const DAY_ABBREVS = ["Su", "M", "T", "W", "Th", "F", "Sa"]

function SectionRow({ section }: SectionRowProps) {
  if (!section) return null

  const days = section.meetingTimes
    .map((mt) =>
      mt.days
        .map((d, i) => (d ? DAY_ABBREVS[i] : ""))
        .filter(Boolean)
        .join("")
    )
    .join(", ")
  const times = section.meetingTimes
    .map((mt) => `${mt.startTime}-${mt.endTime}`)
    .join(", ")

  return (
    <div className="text-muted-foreground flex items-center gap-2 text-sm">
      <span className="font-mono">{section.crn}</span>
      <span>·</span>
      <span>{section.instructor || "TBA"}</span>
      <span>·</span>
      <span>
        {days || "TBA"} {times || ""}
      </span>
      <span>·</span>
      <span
        className={cn(
          section.isOpen
            ? "text-emerald-600 dark:text-emerald-400"
            : "text-red-600 dark:text-red-400"
        )}
      >
        {section.seatsAvailable}/{section.maxEnrollment} seats
      </span>
    </div>
  )
}
