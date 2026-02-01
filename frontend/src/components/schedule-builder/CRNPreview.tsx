import { AlertTriangle, Loader2 } from "lucide-react"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import type { CRNResponse } from "@/lib/api"
import { cn } from "@/lib/utils"

interface CRNPreviewProps {
  crnData: CRNResponse | undefined
  isLoading: boolean
  currentTerm: string
}

export function CRNPreview({
  crnData,
  isLoading,
  currentTerm,
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
    <div className={cn("p-4 text-sm", termMismatch && "bg-amber-500/10")}>
      <div className="flex items-center gap-2">
        <span className="font-medium">
          {section.subject} {section.courseNumber}
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
      <div className="text-muted-foreground">{section.title}</div>
      <div className="text-muted-foreground mt-2 space-y-0.5 text-xs">
        {section.instructor && <div>{section.instructor}</div>}
        <div>{section.credits} credits</div>
      </div>
    </div>
  )
}
