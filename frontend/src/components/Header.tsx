import { Sun, Moon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { AboutDialog } from "@/components/AboutDialog"
import { HelpDialog } from "@/components/HelpDialog"
import { useAppStore } from "@/stores/app-store"
import { useEffectiveTheme } from "@/hooks/use-theme"

export function Header() {
  const { setTheme, setTab } = useAppStore()
  const effectiveTheme = useEffectiveTheme()

  const toggleTheme = () => {
    // Simple toggle between light and dark based on effective appearance
    setTheme(effectiveTheme === "dark" ? "light" : "dark")
  }

  const ThemeIcon = effectiveTheme === "dark" ? Moon : Sun
  const logoSrc =
    effectiveTheme === "dark"
      ? `${import.meta.env.BASE_URL}schopt-logo-dark.svg`
      : `${import.meta.env.BASE_URL}schopt-logo-light.svg`

  return (
    <header className="bg-background border-b">
      <div className="mx-auto flex h-14 max-w-[1920px] items-center justify-between px-4">
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
            onClick={toggleTheme}
            aria-label={`Switch to ${effectiveTheme === "dark" ? "light" : "dark"} mode`}
          >
            <ThemeIcon className="size-4" />
          </Button>
        </div>
      </div>
    </header>
  )
}
