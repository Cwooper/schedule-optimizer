import { useState } from "react"
import { X, Info, AlertTriangle, FlaskConical } from "lucide-react"
import { useAnnouncement } from "@/hooks/use-api"

const DISMISS_KEY_PREFIX = "dismissed-announcement-"

const typeConfig = {
  info: {
    icon: Info,
    bg: "bg-blue-50 dark:bg-blue-950",
    border: "border-blue-200 dark:border-blue-800",
    text: "text-blue-900 dark:text-blue-100",
    iconColor: "text-blue-600 dark:text-blue-400",
  },
  warning: {
    icon: AlertTriangle,
    bg: "bg-amber-50 dark:bg-amber-950",
    border: "border-amber-200 dark:border-amber-800",
    text: "text-amber-900 dark:text-amber-100",
    iconColor: "text-amber-600 dark:text-amber-400",
  },
  beta: {
    icon: FlaskConical,
    bg: "bg-green-50 dark:bg-green-950",
    border: "border-green-200 dark:border-green-800",
    text: "text-green-900 dark:text-green-100",
    iconColor: "text-green-600 dark:text-green-400",
  },
} as const

function isDismissed(id: number): boolean {
  try {
    return localStorage.getItem(`${DISMISS_KEY_PREFIX}${id}`) === "1"
  } catch {
    return false
  }
}

export function AnnouncementBanner() {
  const { data } = useAnnouncement()
  const [dismissedId, setDismissedId] = useState<number | null>(null)

  const announcement = data?.announcement

  if (
    !announcement ||
    dismissedId === announcement.id ||
    isDismissed(announcement.id)
  )
    return null

  const config = typeConfig[announcement.type] ?? typeConfig.info

  const handleDismiss = () => {
    localStorage.setItem(`${DISMISS_KEY_PREFIX}${announcement.id}`, "1")
    setDismissedId(announcement.id)
  }

  const Icon = config.icon

  return (
    <div className="pointer-events-none fixed inset-x-0 top-0 z-50 flex justify-center p-3">
      <div
        className={`pointer-events-auto flex w-full max-w-xl items-center gap-3 rounded-lg border px-4 py-3 shadow-lg ${config.bg} ${config.border}`}
      >
        <Icon className={`size-4 shrink-0 ${config.iconColor}`} />
        <p className={`flex-1 text-sm ${config.text}`}>
          <span className="font-semibold">{announcement.title}</span>
          {announcement.body && (
            <span className="ml-1.5">{announcement.body}</span>
          )}
        </p>
        <button
          onClick={handleDismiss}
          className={`shrink-0 cursor-pointer rounded p-1 transition-colors hover:bg-black/10 dark:hover:bg-white/10 ${config.text}`}
          aria-label="Dismiss announcement"
        >
          <X className="size-4" />
        </button>
      </div>
    </div>
  )
}
