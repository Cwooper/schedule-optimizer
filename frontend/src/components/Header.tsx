import { Sun, Moon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { AboutDialog } from "@/components/AboutDialog"
import { HelpDialog } from "@/components/HelpDialog"
import { useAppStore } from "@/stores/app-store"
import { useEffectiveTheme } from "@/hooks/use-theme"

export function Header() {
  const { theme, setTheme, setTab } = useAppStore()
  const effectiveTheme = useEffectiveTheme()

  const cycleTheme = () => {
    const next =
      theme === "light" ? "dark" : theme === "dark" ? "system" : "light"
    setTheme(next)
  }

  const ThemeIcon = effectiveTheme === "dark" ? Moon : Sun
  const logoSrc =
    effectiveTheme === "dark"
      ? "/schopt-logo-dark.svg"
      : "/schopt-logo-light.svg"

  return (
    <header className="bg-background border-b">
      <div className="mx-auto flex h-14 max-w-7xl items-center justify-between px-4">
        {/* Logo + Title */}
        <button
          onClick={() => setTab("schedule")}
          className="flex items-center gap-1.5 sm:gap-2"
        >
          <img src={logoSrc} alt="" className="size-6" />
          <span className="text-lg font-semibold whitespace-nowrap sm:text-xl">
            <span className="hidden sm:inline">WWU </span>Schedule Optimizer
          </span>
        </button>

        {/* Nav Links + Theme Toggle */}
        <div className="flex items-center gap-1">
          <AboutDialog />
          <HelpDialog />
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={cycleTheme}
            aria-label={`Current theme: ${theme}. Click to cycle.`}
          >
            <ThemeIcon className="size-4" />
          </Button>
        </div>
      </div>
    </header>
  )
}
