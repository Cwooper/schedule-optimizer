import { useState } from "react"
import { Menu } from "lucide-react"
import { useThemeSync } from "@/hooks/use-theme"
import { useSidebarVisible } from "@/hooks/use-sidebar"
import { Header } from "@/components/Header"
import { TabNav } from "@/components/TabNav"
import { Footer } from "@/components/Footer"
import { ScheduleBuilder } from "@/components/ScheduleBuilder"
import { ScheduleView, CourseInfoDialog } from "@/components/schedule"
import { SearchView } from "@/components/SearchView"
import { Button } from "@/components/ui/button"
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from "@/components/ui/drawer"
import { useAppStore } from "@/stores/app-store"

// TODO: Use framer-motion for more animations (course list enter/exit, tab content transitions, schedule navigation)

function App() {
  useThemeSync()
  const [drawerOpen, setDrawerOpen] = useState(false)
  const tab = useAppStore((s) => s.tab)
  const showSidebar = useSidebarVisible()
  const term = useAppStore((s) => s.term)
  const generateResult = useAppStore((s) => s.generateResult)
  const courseDialog = useAppStore((s) => s.courseDialog)
  const closeCourseDialog = useAppStore((s) => s.closeCourseDialog)

  return (
    <div className="bg-background text-foreground flex h-screen flex-col overflow-hidden">
      <Header />

      {/* Tab nav + main content wrapper with max-width constraint */}
      <div className="mx-auto flex w-full min-h-0 max-w-[1920px] flex-1 flex-col overflow-hidden min-[1920px]:border-x">
        {/* Tab Navigation with mobile hamburger */}
        <div className="relative flex items-center justify-center border-b px-4 py-3">
          {/* Mobile drawer trigger - positioned left of tabs */}
          {showSidebar && (
            <Drawer direction="left" open={drawerOpen} onOpenChange={setDrawerOpen}>
              <DrawerTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="absolute left-4 md:hidden"
                >
                  <Menu className="size-5" />
                  <span className="sr-only">Open course builder</span>
                </Button>
              </DrawerTrigger>
              <DrawerContent className="h-full w-80">
                <DrawerHeader className="sr-only">
                  <DrawerTitle>Schedule Builder</DrawerTitle>
                  <DrawerDescription>Add and configure courses for schedule generation</DrawerDescription>
                </DrawerHeader>
                <ScheduleBuilder />
              </DrawerContent>
            </Drawer>
          )}

          <TabNav />
        </div>

        {/* Main content area with sidebar */}
        <div className="flex min-h-0 flex-1 overflow-hidden">
          {/* Desktop sidebar */}
          {showSidebar && (
            <aside className="hidden w-80 min-h-0 flex-col overflow-hidden border-r md:flex">
              <ScheduleBuilder />
            </aside>
          )}

          {/* Main content */}
          <main className="flex-1 overflow-hidden">
            {tab === "schedule" && <ScheduleView />}
            {tab === "search" && <SearchView />}
            {tab === "statistics" && (
              <div className="flex h-full items-center justify-center">
                <p className="text-muted-foreground">Statistics (coming soon)</p>
              </div>
            )}
          </main>
        </div>
      </div>

      <Footer />

      {/* Course info dialog - rendered at app level so it's accessible from anywhere */}
      <CourseInfoDialog
        open={courseDialog.open}
        onOpenChange={(open) => !open && closeCourseDialog()}
        courses={generateResult?.courses}
        sections={generateResult?.sections}
        selectedCrn={courseDialog.selectedCrn}
        selectedCourseKey={courseDialog.selectedCourseKey}
        term={term}
      />
    </div>
  )
}

export default App
