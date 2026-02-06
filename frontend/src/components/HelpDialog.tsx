import { CircleHelp } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"

export function HelpDialog() {
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="ghost" size="sm">
          <CircleHelp className="size-4" />
          <span className="hidden sm:inline">Help</span>
        </Button>
      </DialogTrigger>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>How to Use</DialogTitle>
          <DialogDescription>
            Get started with Schedule Optimizer
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 text-sm">
          {/* Quick Start */}
          <div>
            <h3 className="mb-1.5 font-medium">Quick Start</h3>
            <ol className="text-muted-foreground list-inside list-decimal space-y-1">
              <li>
                Pick a <strong className="text-foreground">term</strong> from
                the sidebar
              </li>
              <li>
                Add courses by{" "}
                <strong className="text-foreground">subject + number</strong> or{" "}
                <strong className="text-foreground">CRN</strong>
              </li>
              <li>
                Hit <strong className="text-foreground">Generate</strong>
              </li>
              <li>
                Browse results with the{" "}
                <strong className="text-foreground">left/right arrows</strong>
              </li>
            </ol>
          </div>

          <hr className="border-border" />

          {/* Schedule Builder */}
          <div>
            <h3 className="mb-1.5 font-medium">Schedule Builder</h3>
            <ul className="text-muted-foreground list-inside list-disc space-y-1">
              <li>
                <strong className="text-foreground">
                  Required vs optional
                </strong>
                : toggle the cicle on a course to guarantee it appears in every
                schedule
              </li>
              <li>
                <strong className="text-foreground">Pin sections</strong>: add
                specific CRNs to restrict which sections are used
              </li>
              <li>
                <strong className="text-foreground">Min / Max courses</strong>:
                allow partial combinations (e.g. "best 4 of 6")
              </li>
            </ul>
          </div>

          <hr className="border-border" />

          {/* Blocked Times */}
          <div>
            <h3 className="mb-1.5 font-medium">Blocked Times</h3>
            <ul className="text-muted-foreground list-inside list-disc space-y-1">
              <li>
                Open the{" "}
                <strong className="text-foreground">&#8942; menu</strong> on the
                schedule view and choose{" "}
                <strong className="text-foreground">Blocked Times</strong>
              </li>
              <li>
                Create a group, then{" "}
                <strong className="text-foreground">paint</strong> directly on
                the grid to block off time slots
              </li>
              <li>Toggle groups on/off without deleting them</li>
            </ul>
          </div>

          <hr className="border-border" />

          {/* Search */}
          <div>
            <h3 className="mb-1.5 font-medium">Course Search</h3>
            <ul className="text-muted-foreground list-inside list-disc space-y-1">
              <li>
                Use the <strong className="text-foreground">Search</strong> tab
                to browse the course catalog
              </li>
              <li>
                Filter by subject, number, title, instructor, credits, or open
                seats
              </li>
              <li>
                Click a result to view details, or add it straight to your
                schedule
              </li>
            </ul>
          </div>

          <hr className="border-border" />

          {/* Tips */}
          <div>
            <h3 className="mb-1.5 font-medium">Tips</h3>
            <ul className="text-muted-foreground list-inside list-disc space-y-1">
              <li>
                Click any course on the schedule grid to view section details
              </li>
              <li>
                Download your schedule as a{" "}
                <strong className="text-foreground">PNG</strong> from the
                <strong className="text-foreground">&#8942;</strong>menu
              </li>
              <li>
                Your selections are saved automatically and persist between
                visits
              </li>
            </ul>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
