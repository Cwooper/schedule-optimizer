import { SiGithub } from "@icons-pack/react-simple-icons"
import { version } from "../../package.json"

const GITHUB_URL = "https://github.com/Cwooper/schedule-optimizer"

export function Footer() {
  return (
    <footer className="bg-background border-t">
      <div className="text-muted-foreground mx-auto flex h-12 max-w-[1920px] items-center justify-center gap-4 px-4 text-sm">
        <a
          href={GITHUB_URL}
          target="_blank"
          rel="noopener noreferrer"
          className="hover:text-foreground flex items-center gap-1.5 transition-colors"
        >
          <SiGithub className="size-4" />
          <span className="hidden sm:inline">View on GitHub</span>
        </a>
        <span aria-hidden="true">·</span>
        <span>v{version}</span>
        <span aria-hidden="true">·</span>
        <span>Made for WWU students</span>
      </div>
    </footer>
  )
}
