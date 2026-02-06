import { useMemo } from "react"
import { Monitor } from "lucide-react"
import { CourseListItem } from "@/components/CourseListItem"
import type { GenerateResponse } from "@/lib/api"
import { hydrateAsyncs, groupSectionsByCourse } from "@/lib/schedule-utils"

interface AsyncCoursesListProps {
  generateResult: GenerateResponse
  onCourseClick: (crn: string) => void
}

export function AsyncCoursesList({
  generateResult,
  onCourseClick,
}: AsyncCoursesListProps) {
  const courseItems = useMemo(() => {
    const hydrated = hydrateAsyncs(generateResult)
    if (hydrated.length === 0) return []
    return groupSectionsByCourse(hydrated)
  }, [generateResult])

  if (courseItems.length === 0) return null

  return (
    <div className="border-t px-4 py-3">
      <div className="text-muted-foreground mb-2 flex items-center gap-2 text-sm">
        <Monitor className="size-4" />
        <span className="font-medium">Async / TBD Courses</span>
      </div>
      <div className="space-y-2">
        {courseItems.map((item) => (
          <CourseListItem
            key={`${item.subject}:${item.courseNumber}`}
            course={item}
            onClick={() => onCourseClick(item.firstCrn)}
            hideAddButton
          />
        ))}
      </div>
    </div>
  )
}
