import { Loader2 } from "lucide-react"
import type { CourseResponse } from "@/lib/api"

interface CoursePreviewProps {
  courseData: CourseResponse | undefined
  isLoading: boolean
}

export function CoursePreview({ courseData, isLoading }: CoursePreviewProps) {
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
    <div className="p-3 text-sm">
      <div className="font-medium">
        {course.subject} {course.courseNumber} – {course.title}
      </div>
      <div className="text-muted-foreground mt-1 flex items-center justify-between text-xs">
        <span>
          {courseData.sectionCount} {sectionLabel}
        </span>
        <span>{course.credits} cr</span>
      </div>
    </div>
  )
}
