import { Plus, Check } from "lucide-react"
import { cn, decodeHtmlEntities } from "@/lib/utils"
import { Button } from "@/components/ui/button"

// Flexible section info - works with both search and generate responses
export interface CourseListSectionInfo {
  crn: string
  term: string
  instructor?: string
  seatsAvailable: number
  isOpen: boolean
}

export interface CourseListItemData {
  subject: string
  courseNumber: string
  title: string
  credits: number
  creditsHigh?: number // For variable credit courses
  sections?: CourseListSectionInfo[]
}

interface CourseListItemProps {
  course: CourseListItemData
  /** Called when the row is clicked (e.g., to open detail dialog) */
  onClick?: () => void
  /** Called when the "+" button is clicked to add course to schedule */
  onAdd?: () => void
  /** Hide the add button even if onAdd is provided */
  hideAddButton?: boolean
  /** Whether this course is already added to schedule */
  isAdded?: boolean
  className?: string
}

/** Format credits, handling variable credit courses */
function formatCredits(credits: number, creditsHigh?: number): string {
  if (creditsHigh && creditsHigh !== credits) {
    return `${credits}-${creditsHigh}`
  }
  return String(credits)
}

/** Get unique instructors from sections */
function getInstructors(sections: CourseListSectionInfo[]): string {
  const instructorSet = new Set<string>()
  for (const s of sections) {
    if (s.instructor) {
      // Get last name only for brevity
      const name = s.instructor.split(",")[0]?.trim()
      if (name) instructorSet.add(name)
    }
  }
  const instructors = [...instructorSet]
  if (instructors.length === 0) return "TBA"
  if (instructors.length === 1) return instructors[0]
  if (instructors.length <= 2) return instructors.join(", ")
  return `${instructors.slice(0, 2).join(", ")} +${instructors.length - 2}`
}

/** Get unique terms from sections */
function getTerms(sections: CourseListSectionInfo[]): string {
  const terms = new Set<string>()
  for (const s of sections) {
    if (s.term) terms.add(s.term)
  }
  if (terms.size === 0) return ""
  if (terms.size === 1) {
    // Format single term nicely
    return formatTermCode([...terms][0])
  }
  return `${terms.size} terms`
}

/** Format term code to readable name */
function formatTermCode(code: string): string {
  if (code.length !== 6) return code
  const year = code.slice(0, 4)
  const quarter = code.slice(4)
  const quarterNames: Record<string, string> = {
    "10": "Winter",
    "20": "Spring",
    "30": "Summer",
    "40": "Fall",
  }
  const name = quarterNames[quarter]
  return name ? `${name} ${year}` : code
}

/** Check if course has any open sections */
function hasOpenSections(sections: CourseListSectionInfo[]): boolean {
  return sections.some((s) => s.isOpen)
}

export function CourseListItem({
  course,
  onClick,
  onAdd,
  hideAddButton = false,
  isAdded = false,
  className,
}: CourseListItemProps) {
  const sections = course.sections ?? []
  const hasSections = sections.length > 0

  const instructors = hasSections ? getInstructors(sections) : null
  const terms = hasSections ? getTerms(sections) : null
  const isOpen = hasSections ? hasOpenSections(sections) : null
  const creditsText = formatCredits(course.credits, course.creditsHigh)

  const showAddButton = onAdd && !hideAddButton

  const handleAddClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (!isAdded) {
      onAdd?.()
    }
  }

  const content = (
    <div className="flex min-w-0 flex-1 flex-col gap-1">
      {/* Top row: Course code, title, and add button */}
      <div className="flex items-start gap-2">
        <div className="flex min-w-0 flex-1 items-baseline gap-2">
          <span className="shrink-0 font-medium">
            {course.subject} {course.courseNumber}
          </span>
          <span className="text-muted-foreground">—</span>
          <span className="truncate text-muted-foreground">
            {decodeHtmlEntities(course.title)}
          </span>
        </div>

        {/* Add button - top right */}
        {showAddButton && (
          <Button
            variant="ghost"
            size="icon"
            className={cn(
              "size-7 shrink-0 -mt-0.5 -mr-1",
              isAdded && "text-emerald-600 dark:text-emerald-400"
            )}
            onClick={handleAddClick}
            disabled={isAdded}
            title={isAdded ? "Already in schedule" : "Add to schedule"}
          >
            {isAdded ? <Check className="size-4" /> : <Plus className="size-4" />}
          </Button>
        )}
      </div>

      {/* Metadata row - only if we have sections */}
      {hasSections && (
        <div className="flex items-center gap-x-3 text-xs text-muted-foreground">
          <div className="flex flex-wrap items-center gap-x-2 gap-y-0.5">
            {terms && <span>{terms}</span>}
            {terms && <span className="text-muted-foreground/50">•</span>}
            <span>{creditsText} cr</span>
            {instructors && (
              <>
                <span className="text-muted-foreground/50">•</span>
                <span>{decodeHtmlEntities(instructors)}</span>
              </>
            )}
          </div>
          {/* Open/Closed badge - right aligned */}
          {isOpen !== null && (
            <span
              className={cn(
                "ml-auto shrink-0 rounded-full px-2 py-0.5 text-xs font-medium",
                isOpen
                  ? "bg-emerald-500/15 text-emerald-700 dark:text-emerald-400"
                  : "bg-red-500/15 text-red-700 dark:text-red-400"
              )}
            >
              {isOpen ? "Open" : "Closed"}
            </span>
          )}
        </div>
      )}
    </div>
  )

  if (onClick) {
    return (
      <button
        type="button"
        onClick={onClick}
        className={cn(
          "flex w-full items-center py-3 px-3 text-left text-sm transition-colors rounded-lg border cursor-pointer",
          "hover:bg-muted/50 hover:border-muted-foreground/20",
          className
        )}
      >
        {content}
      </button>
    )
  }

  return (
    <div className={cn("flex items-center py-3 px-3 text-sm rounded-lg border", className)}>
      {content}
    </div>
  )
}
