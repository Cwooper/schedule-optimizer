import { useRef, useState, useLayoutEffect, useCallback, useEffect } from "react"
import { Calendar, Search, BarChart3 } from "lucide-react"
import { motion } from "framer-motion"
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"
import { useAppStore, type Tab } from "@/stores/app-store"

const tabValues = ["schedule", "search", "statistics"] as const

const tabs: { value: Tab; label: string; icon: React.ReactNode }[] = [
  {
    value: "schedule",
    label: "Schedule",
    icon: <Calendar className="size-5" />,
  },
  { value: "search", label: "Search", icon: <Search className="size-5" /> },
  {
    value: "statistics",
    label: "Statistics",
    icon: <BarChart3 className="size-5" />,
  },
]

export function TabNav() {
  const { tab, setTab } = useAppStore()
  const itemRefs = useRef<Record<Tab, HTMLElement | null>>({
    schedule: null,
    search: null,
    statistics: null,
  })
  const [indicatorStyle, setIndicatorStyle] = useState({ left: 0, width: 0 })

  const updateIndicator = useCallback(() => {
    const el = itemRefs.current[tab]
    if (!el) return

    setIndicatorStyle({
      left: el.offsetLeft,
      width: el.offsetWidth,
    })
  }, [tab])

  useLayoutEffect(() => {
    updateIndicator()
  }, [updateIndicator])

  useEffect(() => {
    let timeoutId: ReturnType<typeof setTimeout>
    const onResize = () => {
      clearTimeout(timeoutId)
      timeoutId = setTimeout(updateIndicator, 100)
    }

    window.addEventListener("resize", onResize)
    return () => {
      window.removeEventListener("resize", onResize)
      clearTimeout(timeoutId)
    }
  }, [updateIndicator])

  const handleValueChange = useCallback(
    (v: string) => {
      if (v && tabValues.includes(v as Tab)) setTab(v as Tab)
    },
    [setTab]
  )

  return (
    <ToggleGroup
      type="single"
      value={tab}
      onValueChange={handleValueChange}
      className="relative rounded-lg bg-muted p-1"
    >
      {indicatorStyle.width > 0 && (
        <motion.span
          aria-hidden="true"
          className="absolute top-1 bottom-1 rounded-md ring-1 ring-black/10 bg-background shadow-sm dark:ring-white/20"
          initial={false}
          animate={{ left: indicatorStyle.left, width: indicatorStyle.width }}
          transition={{ type: "spring", bounce: 0.2, duration: 0.4 }}
        />
      )}
      {tabs.map((t) => (
        <ToggleGroupItem
          key={t.value}
          value={t.value}
          ref={(el) => {
            itemRefs.current[t.value] = el
          }}
          aria-label={t.label}
          className="relative z-10 hover:bg-transparent data-[state=on]:bg-transparent"
        >
          {t.icon}
          <span className="hidden sm:inline">{t.label}</span>
        </ToggleGroupItem>
      ))}
    </ToggleGroup>
  )
}
