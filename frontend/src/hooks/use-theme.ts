import { useEffect, useSyncExternalStore } from "react"
import { useAppStore } from "@/stores/app-store"

const mediaQuery =
  typeof globalThis.window !== "undefined"
    ? globalThis.matchMedia("(prefers-color-scheme: dark)")
    : null

function subscribeToSystemTheme(callback: () => void) {
  mediaQuery?.addEventListener("change", callback)
  return () => mediaQuery?.removeEventListener("change", callback)
}

function getSystemThemeSnapshot(): "light" | "dark" {
  return mediaQuery?.matches ? "dark" : "light"
}

function getServerSnapshot(): "light" | "dark" {
  return "light"
}

/**
 * Syncs the theme from Zustand store to the document.
 * Applies "dark" class to <html> element based on theme preference.
 * Call this once at the app root level.
 */
export function useThemeSync() {
  const theme = useAppStore((state) => state.theme)
  const systemTheme = useSyncExternalStore(
    subscribeToSystemTheme,
    getSystemThemeSnapshot,
    getServerSnapshot
  )

  useEffect(() => {
    const root = document.documentElement
    const isDark =
      theme === "system" ? systemTheme === "dark" : theme === "dark"

    if (isDark) {
      root.classList.add("dark")
    } else {
      root.classList.remove("dark")
    }
  }, [theme, systemTheme])
}

/**
 * Returns the current effective theme (resolved "system" to actual value).
 * Reacts to system theme changes when theme is set to "system".
 */
export function useEffectiveTheme(): "light" | "dark" {
  const theme = useAppStore((state) => state.theme)
  const systemTheme = useSyncExternalStore(
    subscribeToSystemTheme,
    getSystemThemeSnapshot,
    getServerSnapshot
  )

  if (theme === "system") {
    return systemTheme
  }

  return theme === "dark" ? "dark" : "light"
}
