import { Calendar, Search, BarChart3 } from "lucide-react"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useAppStore, type Tab } from "@/stores/app-store"

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

  return (
    <Tabs value={tab} onValueChange={(v) => setTab(v as Tab)}>
      <TabsList className="h-11">
        {tabs.map((t) => (
          <TabsTrigger
            key={t.value}
            value={t.value}
            className="gap-2 px-4 text-base"
          >
            {t.icon}
            <span className="hidden sm:inline">{t.label}</span>
          </TabsTrigger>
        ))}
      </TabsList>
    </Tabs>
  )
}
