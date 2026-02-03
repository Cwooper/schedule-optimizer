import {
  ChevronDown,
  ChevronRight,
  Circle,
  CircleDot,
  X,
  AlertTriangle,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"
import { Toggle } from "@/components/ui/toggle"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import type { CourseSlot } from "@/stores/app-store"
import type { Term, CourseValidationResult } from "@/lib/api"
import { cn, decodeHtmlEntities } from "@/lib/utils"

interface CourseRowProps {
  slot: CourseSlot
  expanded: boolean
  onToggleExpand: () => void
  onToggleRequired: () => void
  onRemove: () => void
  onRemoveSection: (crn: string) => void
  onToggleSectionRequired: (crn: string) => void
  onCourseClick?: (courseKey: string) => void
  onSectionClick?: (crn: string) => void
  currentTerm: string
  terms: Term[]
  validation?: CourseValidationResult
}

export function CourseRow({
  slot,
  expanded,
  onToggleExpand,
  onToggleRequired,
  onRemove,
  onRemoveSection,
  onToggleSectionRequired,
  onCourseClick,
  onSectionClick,
  currentTerm,
  terms,
  validation,
}: CourseRowProps) {
  const hasSections = slot.sections && slot.sections.length > 0
  const hasTermMismatch =
    slot.sections?.some((s) => s.term !== currentTerm) ?? false
  const courseNotInTerm = validation?.exists === false

  const getTermName = (termCode: string) =>
    terms.find((t) => t.code === termCode)?.name ?? termCode

  return (
    <Collapsible open={expanded} onOpenChange={onToggleExpand}>
      <div
        className={cn(
          "hover:bg-muted/50 flex items-center gap-2 rounded px-2 py-1.5 transition-colors",
          (hasTermMismatch || courseNotInTerm) && "text-muted-foreground"
        )}
      >
        {/* Expand/collapse trigger */}
        <CollapsibleTrigger asChild disabled={!hasSections}>
          <button
            aria-expanded={expanded}
            aria-label={expanded ? "Collapse sections" : "Expand sections"}
            className={cn(
              "size-5 shrink-0",
              !hasSections && "cursor-default opacity-0"
            )}
          >
            {expanded ? (
              <ChevronDown className="size-4" />
            ) : (
              <ChevronRight className="size-4" />
            )}
          </button>
        </CollapsibleTrigger>

        {/* Warning icon for course not in term or section term mismatch */}
        {courseNotInTerm && (
          <Tooltip>
            <TooltipTrigger asChild>
              <AlertTriangle className="size-4 shrink-0 text-amber-500" />
            </TooltipTrigger>
            <TooltipContent>
              This course is not offered in {getTermName(currentTerm)}
            </TooltipContent>
          </Tooltip>
        )}
        {!courseNotInTerm && hasTermMismatch && (
          <Tooltip>
            <TooltipTrigger asChild>
              <AlertTriangle className="size-4 shrink-0 text-amber-500" />
            </TooltipTrigger>
            <TooltipContent>
              Some sections are from a different term
            </TooltipContent>
          </Tooltip>
        )}

        {/* Course name */}
        {onCourseClick ? (
          <button
            type="button"
            className="flex-1 truncate text-left text-sm font-medium hover:underline"
            title={slot.title ? decodeHtmlEntities(slot.title) : undefined}
            onClick={() => onCourseClick(`${slot.subject}:${slot.courseNumber}`)}
          >
            {slot.displayName}
          </button>
        ) : (
          <span
            className="flex-1 truncate text-sm font-medium"
            title={slot.title ? decodeHtmlEntities(slot.title) : undefined}
          >
            {slot.displayName}
          </span>
        )}

        {/* Required toggle */}
        <Toggle
          size="sm"
          pressed={slot.required}
          onPressedChange={onToggleRequired}
          className="size-7 p-0"
          title={slot.required ? "Required" : "Optional"}
        >
          {slot.required ? (
            <CircleDot className="size-4" />
          ) : (
            <Circle className="size-4" />
          )}
        </Toggle>

        {/* Remove button */}
        <Button
          variant="ghost"
          size="icon"
          className="size-7"
          onClick={onRemove}
          aria-label={`Remove ${slot.displayName}`}
        >
          <X className="size-4" />
        </Button>
      </div>

      {/* Expanded section list - tree style */}
      {hasSections && (
        <CollapsibleContent>
          <div className="ml-4 pl-3 py-1">
            {slot.sections!.map((section, idx) => {
              const isLast = idx === slot.sections!.length - 1
              return (
                <div
                  key={section.crn}
                  className={cn(
                    "hover:bg-muted/50 relative flex items-center gap-2 rounded px-2 py-1.5 text-sm",
                    section.term !== currentTerm && "text-muted-foreground"
                  )}
                >
                  {/* Tree connectors */}
                  <div className="bg-border absolute -left-3 top-0 h-1/2 w-px" />
                  <div className="bg-border absolute -left-3 top-1/2 h-px w-2" />
                  {!isLast && (
                    <div className="bg-border absolute -left-3 top-1/2 h-1/2 w-px" />
                  )}

                  {/* Warning icon for term mismatch */}
                  {section.term !== currentTerm && (
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <AlertTriangle className="size-4 shrink-0 text-amber-500" />
                      </TooltipTrigger>
                      <TooltipContent>
                        This section is from {getTermName(section.term)}
                      </TooltipContent>
                    </Tooltip>
                  )}

                  {/* CRN and instructor */}
                  {onSectionClick ? (
                    <button
                      type="button"
                      className="flex-1 truncate text-left font-medium hover:underline"
                      onClick={() => onSectionClick(section.crn)}
                    >
                      {section.crn}
                      {section.instructor && (
                        <span className="text-muted-foreground font-normal">
                          {" "}
                          ({decodeHtmlEntities(section.instructor)})
                        </span>
                      )}
                    </button>
                  ) : (
                    <span className="flex-1 truncate font-medium">
                      {section.crn}
                      {section.instructor && (
                        <span className="text-muted-foreground font-normal">
                          {" "}
                          ({decodeHtmlEntities(section.instructor)})
                        </span>
                      )}
                    </span>
                  )}

                  {/* Required toggle */}
                  <Toggle
                    size="sm"
                    pressed={section.required}
                    onPressedChange={() => onToggleSectionRequired(section.crn)}
                    className="size-7 p-0"
                    title={section.required ? "Required" : "Optional"}
                  >
                    {section.required ? (
                      <CircleDot className="size-4" />
                    ) : (
                      <Circle className="size-4" />
                    )}
                  </Toggle>

                  {/* Remove button */}
                  <Button
                    variant="ghost"
                    size="icon"
                    className="size-7"
                    onClick={() => onRemoveSection(section.crn)}
                    title="Remove section"
                    aria-label={`Remove section ${section.crn}`}
                  >
                    <X className="size-4" />
                  </Button>
                </div>
              )
            })}
          </div>
        </CollapsibleContent>
      )}
    </Collapsible>
  )
}
