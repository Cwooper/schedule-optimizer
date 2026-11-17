import type { CalendarExportOptions } from "@/lib/ics"
import { selectScheduleTerm, useAppStore, type Tab } from "@/stores/app-store"

/** Each event's target; mirrors eventTargets in backend/internal/api/events.go. */
interface EventTargets {
  visit: "mobile" | "desktop"
  tab: Tab
  ics_export: string // from icsExportTarget
  png_export: "schedule" | "finals"
  finals_view: undefined
  feedback_open: undefined
}

type EventName = keyof EventTargets
type TargetArgs<N extends EventName> = EventTargets[N] extends undefined
  ? []
  : [target: EventTargets[N]]

/**
 * The included .ics options, comma-joined in the order of icsOptions in
 * backend/internal/api/events.go, which rejects any other order.
 */
export function icsExportTarget(options: CalendarExportOptions): string {
  const flags: [string, boolean][] = [
    ["classes", options.classes],
    ["finals", options.finals],
    ["holidays", options.holidays],
    ["reg1", options.registration.I],
    ["reg2", options.registration.II],
    ["reg3", options.registration.III],
    ["deadlines", options.deadlines],
  ]
  return flags
    .filter(([, on]) => on)
    .map(([key]) => key)
    .join(",")
}

/** Records a usage event. Never throws: a lost event is not the user's problem. */
export function track<N extends EventName>(name: N, ...[target]: TargetArgs<N>): void {
  // Automated browsers (headless crawlers, Playwright) aren't visitors.
  if (navigator.webdriver) return
  try {
    const url = `${__APP_BASE_URL__}api/events`
    const state = useAppStore.getState()
    const body = JSON.stringify({
      session_id: state.sessionId,
      name,
      target,
      term: selectScheduleTerm(state) || undefined,
    })
    // sendBeacon returns false when the browser's beacon queue is full.
    if (navigator.sendBeacon?.(url, new Blob([body], { type: "application/json" }))) return
    fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body,
      keepalive: true,
    }).catch(() => {})
  } catch {
    // Same as a rejected fetch: drop the event.
  }
}
