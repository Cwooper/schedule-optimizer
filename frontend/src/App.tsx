import { useState, useCallback } from "react"
import { Menu } from "lucide-react"
import { toast } from "sonner"
import { useThemeSync } from "@/hooks/use-theme"
import { useSidebarVisible } from "@/hooks/use-sidebar"
import { AnnouncementBanner } from "@/components/AnnouncementBanner"
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
  const searchResult = useAppStore((s) => s.searchResult)
  const slots = useAppStore((s) => s.slots)
  const addSectionToSlot = useAppStore((s) => s.addSectionToSlot)
  const courseDialog = useAppStore((s) => s.courseDialog)
  const closeCourseDialog = useAppStore((s) => s.closeCourseDialog)

  // Select the correct data source based on which view opened the dialog.
  // This prevents duplicate sections when the same course exists in both
  // generate and search results (they use different section key formats).
  const dialogSource = courseDialog.source
  const dialogCourses = (() => {
    if (dialogSource === "search") return searchResult?.courses
    if (dialogSource === "schedule") return generateResult?.courses
    return undefined
  })()
  const dialogSections = (() => {
    if (dialogSource === "schedule") return generateResult?.sections
    if (dialogSource === "search") return searchResult?.sections
    return undefined
  })()

  // Self-contained handler — receives all info from the dialog, no data source lookup
  const handleAddSection = useCallback(
    (
      crn: string,
      sectionTerm: string,
      course: { subject: string; courseNumber: string; title: string },
      instructor: string | null
    ) => {
      const result = addSectionToSlot(
        crn,
        sectionTerm,
        course.subject,
        course.courseNumber,
        course.title,
        instructor
      )
      if (result === "added") {
        toast.success(
          `Added ${course.subject} ${course.courseNumber} (CRN: ${crn}) to schedule`
        )
      } else if (result === "updated") {
        toast.success(
          `Pinned CRN ${crn} to ${course.subject} ${course.courseNumber}`
        )
      }
    },
    [addSectionToSlot]
  )

  const isSectionAdded = useCallback(
    (crn: string, sectionTerm: string) =>
      slots.some((s) =>
        s.sections?.some((sec) => sec.crn === crn && sec.term === sectionTerm)
      ),
    [slots]
  )

  return (
    <div className="bg-background text-foreground flex h-screen flex-col overflow-hidden">
      <AnnouncementBanner />
      <Header />

      {/* Tab nav + main content wrapper with max-width constraint */}
      <div className="mx-auto flex min-h-0 w-full max-w-480 flex-1 flex-col overflow-hidden min-[1920px]:border-x">
        {/* Tab Navigation with mobile hamburger */}
        <div className="relative flex items-center justify-center border-b px-4 py-3">
          {/* Mobile drawer trigger - positioned left of tabs */}
          {showSidebar && (
            <Drawer
              direction="left"
              open={drawerOpen}
              onOpenChange={setDrawerOpen}
            >
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
                  <DrawerDescription>
                    Add and configure courses for schedule generation
                  </DrawerDescription>
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
            <aside className="hidden min-h-0 w-80 flex-col overflow-hidden border-r md:flex">
              <ScheduleBuilder />
            </aside>
          )}

          {/* Main content */}
          <main className="flex-1 overflow-hidden">
            {tab === "schedule" && <ScheduleView />}
            {tab === "search" && <SearchView />}
            {tab === "statistics" && (
              <div className="flex h-full flex-col items-center justify-center">
                <p className="text-muted-foreground">
                  Statistics (coming soon)
                </p>
                <p className="text-muted mt-8 text-xs">
                  I'm waiting for Western to get back to me :(
                </p>
              </div>
            )}
          </main>
        </div>
      </div>

      <Footer />

      {/* Course info dialog - single instance, accessible from anywhere */}
      <CourseInfoDialog
        open={courseDialog.open}
        onOpenChange={(open) => !open && closeCourseDialog()}
        courses={dialogCourses}
        sections={dialogSections}
        selectedCrn={courseDialog.selectedCrn}
        selectedCourseKey={courseDialog.selectedCourseKey}
        term={term}
        onAddSection={handleAddSection}
        isSectionAdded={isSectionAdded}
      />
    </div>
  )
}

export default App
