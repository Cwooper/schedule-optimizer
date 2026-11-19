import { CircleHelp } from "lucide-react"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import type { Final } from "@/lib/api"
import { FINAL_GUESS_REASONS } from "@/lib/finals"

/** Marks a final placed by a guess, with the reason on hover or tap. */
export function FinalGuessIcon({ final }: { final: Final }) {
  if (final.certain || !final.reason) return null
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span className="inline-flex cursor-help" aria-label="Estimated final">
          <CircleHelp className="text-muted-foreground size-3" />
        </span>
      </TooltipTrigger>
      <TooltipContent>{FINAL_GUESS_REASONS[final.reason]}</TooltipContent>
    </Tooltip>
  )
}
