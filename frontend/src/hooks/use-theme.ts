import { useEffect } from "react"
import { useAppStore } from "@/stores/app-store"

/**
 * Syncs the theme from Zustand store to the document.
 * Applies "dark" class to <html> element based on theme preference.
 * Call this once at the app root level.
 */
export function useThemeSync() {
  const theme = useAppStore((state) => state.theme)

  useEffect(() => {
    const root = document.documentElement

    function applyTheme(isDark: boolean) {
      if (isDark) {
        root.classList.add("dark")
      } else {
        root.classList.remove("dark")
      }
    }

    if (theme === "system") {
      const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)")
      applyTheme(mediaQuery.matches)

      function handleChange(e: MediaQueryListEvent) {
        applyTheme(e.matches)
      }

      mediaQuery.addEventListener("change", handleChange)
      return () => mediaQuery.removeEventListener("change", handleChange)
    } else {
      applyTheme(theme === "dark")
    }
  }, [theme])
}

/**
 * Returns the current effective theme (resolved "system" to actual value).
 */
export function useEffectiveTheme(): "light" | "dark" {
  const theme = useAppStore((state) => state.theme)

  if (theme === "system") {
    // This won't react to system changes, but useThemeSync handles that
    return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light"
  }

  return theme
}
