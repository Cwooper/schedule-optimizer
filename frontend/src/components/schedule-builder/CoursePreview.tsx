import { Loader2 } from "lucide-react"
import type { CourseResponse } from "@/lib/api"
import { decodeHtmlEntities } from "@/lib/utils"

interface CoursePreviewProps {
  courseData: CourseResponse | undefined
  isLoading: boolean
  onAdd?: () => void
}

export function CoursePreview({
  courseData,
  isLoading,
  onAdd,
}: CoursePreviewProps) {
  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-4">
        <Loader2 className="text-muted-foreground size-4 animate-spin" />
      </div>
    )
  }

  if (!courseData) {
    return null
  }

  if (!courseData.course) {
    return (
      <div className="text-muted-foreground p-4 text-sm">Course not found</div>
    )
  }

  const course = courseData.course
  const sectionLabel =
    courseData.sectionCount === 1 ? "section" : "sections"

  return (
    <button
      type="button"
      className="hover:bg-muted/50 w-full cursor-pointer p-3 text-left text-sm transition-colors"
      onClick={() => onAdd?.()}
    >
      <div className="font-medium">
        {course.subject} {course.courseNumber} – {decodeHtmlEntities(course.title)}
      </div>
      <div className="text-muted-foreground mt-1 flex items-center justify-between text-xs">
        <span>
          {courseData.sectionCount} {sectionLabel}
        </span>
        <span>{course.credits} cr</span>
      </div>
    </button>
  )
}
