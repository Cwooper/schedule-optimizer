import { useState } from "react"
import { CircleHelp, Bug, ArrowLeft, ExternalLink, Loader2 } from "lucide-react"
import { SiGithub } from "@icons-pack/react-simple-icons"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { submitFeedback } from "@/lib/api"

const GITHUB_ISSUES_URL =
  "https://github.com/Cwooper/schedule-optimizer/issues/new?title=Bug%3A+&labels=bug&body=Be+descriptive+and+provide+images+or+screenshots+if+possible."
const MAX_LENGTH = 1000

export function HelpDialog() {
  const [view, setView] = useState<"help" | "feedback">("help")
  const [message, setMessage] = useState("")
  const [submitting, setSubmitting] = useState(false)

  const resetFeedback = () => {
    setView("help")
    setMessage("")
  }

  const handleSubmit = async () => {
    const trimmed = message.trim()
    if (!trimmed) return

    setSubmitting(true)
    try {
      await submitFeedback(trimmed)
      toast.success("Thank you for the feedback!")
      resetFeedback()
    } catch {
      toast.error("Failed to submit feedback. Please try again.")
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog onOpenChange={(open) => !open && resetFeedback()}>
      <DialogTrigger asChild>
        <Button variant="ghost" size="sm">
          <CircleHelp className="size-4" />
          <span className="hidden sm:inline">Help</span>
        </Button>
      </DialogTrigger>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-lg">
        {view === "help" ? (
          <>
            <DialogHeader>
              <DialogTitle>Help & Feedback</DialogTitle>
              <DialogDescription>
                Get started with Schedule Optimizer
              </DialogDescription>
            </DialogHeader>

            {/* Feedback & Bugs */}
            <div>
              <h3 className="mb-1.5 font-medium">Feedback & Bugs</h3>
              <p className="text-muted-foreground mb-3">
                Let us know what you like, what you don&apos;t like, or if
                something isn&apos;t working. Please be descriptive on bugs! It
                helps a lot! You can also open an issue on GitHub.
              </p>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  className="flex-1"
                  onClick={() => setView("feedback")}
                >
                  <Bug className="size-4" />
                  Send Feedback
                </Button>
                <Button variant="outline" size="sm" className="flex-1" asChild>
                  <a
                    href={GITHUB_ISSUES_URL}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <SiGithub className="size-3.5" />
                    GitHub Issues
                    <ExternalLink className="size-3" />
                  </a>
                </Button>
              </div>
            </div>

            <hr className="border-border" />

            <HelpContent />
          </>
        ) : (
          <>
            <DialogHeader>
              <div className="flex items-center gap-2">
                <button
                  onClick={resetFeedback}
                  className="text-muted-foreground hover:text-foreground rounded p-0.5 transition-colors"
                  aria-label="Back to help"
                >
                  <ArrowLeft className="size-4" />
                </button>
                <DialogTitle>Send Feedback</DialogTitle>
              </div>
              <DialogDescription>
                What do you like, dislike, or what&apos;s broken? Be as
                descriptive as you can.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-3">
              <Textarea
                value={message}
                onChange={(e) =>
                  setMessage(e.target.value.slice(0, MAX_LENGTH))
                }
                placeholder="What went wrong? What could be better?"
                className="min-h-32 resize-none"
                maxLength={MAX_LENGTH}
                autoFocus
              />
              <div className="flex items-center justify-between">
                <span
                  className={`text-xs ${
                    message.length > MAX_LENGTH * 0.9
                      ? "text-red-500"
                      : "text-muted-foreground"
                  }`}
                >
                  {message.length}/{MAX_LENGTH}
                </span>
                <Button
                  size="sm"
                  onClick={handleSubmit}
                  disabled={!message.trim() || submitting}
                >
                  {submitting && <Loader2 className="size-4 animate-spin" />}
                  Submit
                </Button>
              </div>
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}

function HelpContent() {
  return (
    <div className="space-y-4 text-sm">
      {/* Quick Start */}
      <div>
        <h3 className="mb-1.5 font-medium">Quick Start</h3>
        <ol className="text-muted-foreground list-inside list-decimal space-y-1">
          <li>
            Pick a <strong className="text-foreground">term</strong> from the
            sidebar
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
            <strong className="text-foreground">Required vs optional</strong>:
            toggle the cicle on a course to guarantee it appears in every
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
            Open the <strong className="text-foreground">&#8942; menu</strong>{" "}
            on the schedule view and choose{" "}
            <strong className="text-foreground">Blocked Times</strong>
          </li>
          <li>
            Create a group, then{" "}
            <strong className="text-foreground">paint</strong> directly on the
            grid to block off time slots
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
            Use the <strong className="text-foreground">Search</strong> tab to
            browse the course catalog
          </li>
          <li>
            Filter by subject, number, title, instructor, credits, or open seats
          </li>
          <li>
            Click a result to view details, or add it straight to your schedule
          </li>
        </ul>
      </div>

      <hr className="border-border" />

      {/* Tips */}
      <div>
        <h3 className="mb-1.5 font-medium">Tips</h3>
        <ul className="text-muted-foreground list-inside list-disc space-y-1">
          <li>Click any course on the schedule grid to view section details</li>
          <li>
            Download your schedule as a{" "}
            <strong className="text-foreground">PNG</strong> from the
            <strong className="text-foreground">&#8942;</strong>menu
          </li>
          <li>
            Your selections are saved automatically and persist between visits
          </li>
        </ul>
      </div>
    </div>
  )
}
