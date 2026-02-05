import { cn } from "@/lib/utils"

export interface CourseListItemData {
  subject: string
  courseNumber: string
  title: string
}

interface CourseListItemProps {
  course: CourseListItemData
  onClick?: () => void
  className?: string
}

export function CourseListItem({
  course,
  onClick,
  className,
}: CourseListItemProps) {
  const content = (
    <>
      <span className="font-medium">
        {course.subject} {course.courseNumber}
      </span>
      <span className="text-muted-foreground mx-2">—</span>
      <span className="truncate">{course.title}</span>
    </>
  )

  if (onClick) {
    return (
      <button
        type="button"
        onClick={onClick}
        className={cn(
          "hover:bg-muted/50 flex w-full items-center py-2 text-left text-sm transition-colors",
          className
        )}
      >
        {content}
      </button>
    )
  }

  return (
    <div className={cn("flex items-center py-2 text-sm", className)}>
      {content}
    </div>
  )
}
