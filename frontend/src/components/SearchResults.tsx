import { memo } from "react"
import {
  CourseListItem,
  type CourseListItemData,
} from "@/components/CourseListItem"
import type { SearchResponse } from "@/lib/api"

interface SearchResultsProps {
  hasSearched: boolean
  isFetching: boolean
  searchResult: SearchResponse | null
  onCourseClick: (courseKey: string) => void
  onAddCourse: (courseKey: string) => void
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

export const SearchResults = memo(function SearchResults({
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
})
