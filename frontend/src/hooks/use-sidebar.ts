import { useAppStore, type Tab } from "@/stores/app-store"

/** Tabs that should show the sidebar */
const SIDEBAR_TABS: Tab[] = ["schedule", "search"]

/** Hook to determine if the sidebar should be visible based on current tab */
export function useSidebarVisible(): boolean {
  const tab = useAppStore((s) => s.tab)
  return SIDEBAR_TABS.includes(tab)
}
