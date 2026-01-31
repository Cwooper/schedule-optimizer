import { useThemeSync } from "@/hooks/use-theme"
import { Header } from "@/components/Header"
import { TabNav } from "@/components/TabNav"
import { Footer } from "@/components/Footer"

function App() {
  useThemeSync()

  return (
    <div className="bg-background text-foreground flex min-h-screen flex-col">
      <Header />
      <TabNav />
      {/* TODO: Layout with Sidebar + MainContent */}
      <main className="flex flex-1 items-center justify-center">
        <p className="text-muted-foreground">Content area</p>
      </main>
      <Footer />
    </div>
  )
}

export default App
