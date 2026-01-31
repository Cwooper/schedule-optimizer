import { Info } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"

export function AboutDialog() {
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="ghost" size="sm">
          <Info className="size-4" />
          <span className="hidden sm:inline">About</span>
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>About WWU Schedule Optimizer</DialogTitle>
          <DialogDescription>
            A course scheduling tool for WWU students
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 text-sm">
          {/* TODO: Add more about content */}
          <p>
            Schedule Optimizer helps Western Washington University students
            build conflict-free class schedules. Add your desired courses, set
            your preferences, and generate optimized schedule combinations.
          </p>
          <p className="text-muted-foreground text-xs">
            This project is not affiliated with Western Washington University.
            It is an independent initiative developed solely for educational and
            personal use. All data provided by this project is for informational
            purposes only and should not be considered official or binding. Use
            at your own discretion.
          </p>
        </div>
      </DialogContent>
    </Dialog>
  )
}
