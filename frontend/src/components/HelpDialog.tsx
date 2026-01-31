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
      <DialogContent>
        <DialogHeader>
          <DialogTitle>How to Use</DialogTitle>
          <DialogDescription>
            Get started with Schedule Optimizer
          </DialogDescription>
        </DialogHeader>
        {/* TODO: Help content */}
        <div className="text-muted-foreground text-sm">
          Content coming soon.
        </div>
      </DialogContent>
    </Dialog>
  )
}
