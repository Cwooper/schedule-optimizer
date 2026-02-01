import { useState } from "react"
import { Menu } from "lucide-react"
import { useThemeSync } from "@/hooks/use-theme"
import { Header } from "@/components/Header"
import { TabNav } from "@/components/TabNav"
import { Footer } from "@/components/Footer"
import { ScheduleBuilder } from "@/components/ScheduleBuilder"
import { Button } from "@/components/ui/button"
import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from "@/components/ui/drawer"

function App() {
  useThemeSync()
  const [drawerOpen, setDrawerOpen] = useState(false)

  return (
    <div className="bg-background text-foreground flex min-h-screen flex-col">
      <Header />

      {/* Tab Navigation with mobile hamburger */}
      <div className="relative flex items-center justify-center border-b px-4 py-3">
        {/* Mobile drawer trigger - positioned left of tabs */}
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
            </DrawerHeader>
            <ScheduleBuilder />
          </DrawerContent>
        </Drawer>

        <TabNav />
      </div>

      {/* Main content area with sidebar */}
      <div className="flex flex-1 overflow-hidden">
        {/* Desktop sidebar */}
        <aside className="hidden w-80 flex-col border-r md:flex">
          <ScheduleBuilder />
        </aside>

        {/* Main content */}
        <main className="flex flex-1 items-center justify-center overflow-auto">
          <p className="text-muted-foreground">Content area</p>
        </main>
      </div>

      <Footer />
    </div>
  )
}

export default App
