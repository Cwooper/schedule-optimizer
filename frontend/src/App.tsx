import { useState, useCallback } from "react"
import { Menu } from "lucide-react"
import { toast } from "sonner"
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
import { genId } from "@/lib/utils"
import type { MeetingTime } from "@/lib/api"

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
  const addSlot = useAppStore((s) => s.addSlot)
  const courseDialog = useAppStore((s) => s.courseDialog)
  const closeCourseDialog = useAppStore((s) => s.closeCourseDialog)

  // Merge both datasets so the dialog works regardless of where it was opened.
  // Course keys use the same format (SUBJECT:NUMBER) — search data (fresher) overrides.
  // Section keys differ (CRN vs term:CRN) so they coexist without collision.
  // Search sections lack meetingTimes — pad with empty array for dialog compatibility.
  // The dialog's CRN fallback fetch handles loading them on expand.
  const paddedSearchSections = searchResult?.sections
    ? Object.fromEntries(
        Object.entries(searchResult.sections).map(([key, s]) => [
          key,
          { ...s, meetingTimes: [] as MeetingTime[] },
        ])
      )
    : {}
  const dialogCourses = { ...generateResult?.courses, ...searchResult?.courses }
  const dialogSections = { ...generateResult?.sections, ...paddedSearchSections }

  // Self-contained handler — receives all info from the dialog, no data source lookup
  const handleAddSection = useCallback(
    (crn: string, sectionTerm: string, course: { subject: string; courseNumber: string; title: string }, instructor: string | null) => {
      addSlot({
        id: genId(),
        subject: course.subject,
        courseNumber: course.courseNumber,
        displayName: `${course.subject} ${course.courseNumber}`,
        title: course.title,
        required: false,
        sections: [{ crn, term: sectionTerm, instructor, required: false }],
      })
      toast.success(`Added ${course.subject} ${course.courseNumber} (CRN: ${crn}) to schedule`)
    },
    [addSlot]
  )

  const isSectionAdded = useCallback(
    (crn: string) => slots.some((s) => s.sections?.some((sec) => sec.crn === crn)),
    [slots]
  )

  return (
    <div className="bg-background text-foreground flex h-screen flex-col overflow-hidden">
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
              <div className="flex h-full items-center justify-center">
                <p className="text-muted-foreground">
                  Statistics (coming soon)
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
