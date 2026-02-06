import { Info, ExternalLink } from "lucide-react"
import { SiGithub } from "@icons-pack/react-simple-icons"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"

const GITHUB_URL = "https://github.com/Cwooper/schedule-optimizer"

export function AboutDialog() {
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="ghost" size="sm">
          <Info className="size-4" />
          <span className="hidden sm:inline">About</span>
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>About WWU Schedule Optimizer</DialogTitle>
          <DialogDescription>
            A course scheduling tool for WWU students
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 text-sm">
          <p>
            Schedule Optimizer helps Western Washington University students
            build conflict-free class schedules. Add your desired courses, set
            your preferences, and generate optimized schedule combinations
            instantly.
          </p>

          <p className="text-muted-foreground">
            Go · React · SQLite · MIT License ·{" "}
            <a
              href={GITHUB_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="text-foreground inline-flex items-center gap-1 hover:underline"
            >
              <SiGithub className="size-3" />
              GitHub
            </a>
          </p>

          {/* Contributors */}
          <div className="space-y-3">
            <div>
              <h3 className="text-muted-foreground mb-1.5 text-xs font-medium uppercase tracking-wide">
                Author &amp; Maintainer
              </h3>
              <ContributorCard name="Cooper Morgan" href="https://cwooper.me" />
            </div>
            <div>
              <h3 className="text-muted-foreground mb-1.5 text-xs font-medium uppercase tracking-wide">
                Previous Contributor
              </h3>
              <ContributorCard
                name="Konnor Kooi"
                href="https://konnorkooi.com"
              />
            </div>
          </div>

          <hr className="border-border" />

          <p className="text-muted-foreground text-center text-xs">
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

function ContributorCard({ name, href }: { name: string; href: string }) {
  return (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      className="flex items-center justify-between rounded-lg border bg-gradient-to-br from-muted/20 to-transparent p-3 transition-all duration-300 hover:border-green-400/50 hover:shadow-[0_0_15px_-5px_rgba(74,222,128,0.25)]"
    >
      <span className="font-medium">{name}</span>
      <ExternalLink className="text-muted-foreground size-3.5 shrink-0" />
    </a>
  )
}
