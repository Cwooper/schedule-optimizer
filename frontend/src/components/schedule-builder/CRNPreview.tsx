import { AlertTriangle, Loader2 } from "lucide-react"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import type { CRNResponse } from "@/lib/api"
import { cn, decodeHtmlEntities } from "@/lib/utils"

interface CRNPreviewProps {
  crnData: CRNResponse | undefined
  isLoading: boolean
  currentTerm: string
  onAdd?: () => void
}

export function CRNPreview({
  crnData,
  isLoading,
  currentTerm,
  onAdd,
}: CRNPreviewProps) {
  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-4">
        <Loader2 className="text-muted-foreground size-4 animate-spin" />
      </div>
    )
  }

  if (!crnData) {
    return null
  }

  if (!crnData.section) {
    return (
      <div className="text-muted-foreground p-4 text-sm">CRN not found</div>
    )
  }

  const section = crnData.section
  const termMismatch = section.term !== currentTerm

  return (
    <button
      type="button"
      className={cn(
        "hover:bg-muted/50 w-full cursor-pointer p-3 text-left text-sm transition-colors",
        termMismatch && "bg-amber-500/10 hover:bg-amber-500/20"
      )}
      onClick={() => onAdd?.()}
    >
      <div className="flex items-center gap-2">
        <span className="font-medium">
          {section.subject} {section.courseNumber} – {decodeHtmlEntities(section.title)}
        </span>
        {termMismatch && (
          <Tooltip>
            <TooltipTrigger asChild>
              <AlertTriangle className="size-4 shrink-0 text-amber-500" />
            </TooltipTrigger>
            <TooltipContent>This CRN is from a different term</TooltipContent>
          </Tooltip>
        )}
      </div>
      <div className="text-muted-foreground mt-1 flex items-center justify-between text-xs">
        <span>
          {section.instructor ? decodeHtmlEntities(section.instructor) : "TBA"}
        </span>
        <span>{section.credits} cr</span>
      </div>
    </button>
  )
}
