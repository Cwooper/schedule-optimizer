import { useSyncExternalStore } from "react"

// Matches Tailwind's `md` breakpoint (768px)
const MD_BREAKPOINT = "(min-width: 768px)"

const desktopQuery =
  typeof globalThis.window !== "undefined"
    ? globalThis.matchMedia(MD_BREAKPOINT)
    : null

function subscribeToDesktop(callback: () => void) {
  desktopQuery?.addEventListener("change", callback)
  return () => desktopQuery?.removeEventListener("change", callback)
}

function getDesktopSnapshot(): boolean {
  return desktopQuery?.matches ?? false
}

function getServerSnapshot(): boolean {
  return false
}

/**
 * Returns true when viewport is at least md breakpoint (768px).
 * Uses useSyncExternalStore for consistent behavior with SSR.
 */
export function useIsDesktop(): boolean {
  return useSyncExternalStore(
    subscribeToDesktop,
    getDesktopSnapshot,
    getServerSnapshot
  )
}
