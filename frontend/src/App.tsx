import { useThemeSync } from "@/hooks/use-theme"

function App() {
  useThemeSync()

  return (
    <div className="min-h-screen bg-background text-foreground flex items-center justify-center">
      <div className="text-center">
        <h1 className="text-4xl font-bold mb-4">
          Schedule Optimizer
        </h1>
        <p className="text-muted-foreground">Frontend initialized</p>
      </div>
    </div>
  )
}

export default App
