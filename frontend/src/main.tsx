import { StrictMode } from "react"
import { createRoot } from "react-dom/client"
import {
  QueryClient,
  QueryClientProvider,
  QueryCache,
  MutationCache,
} from "@tanstack/react-query"
import { toast } from "sonner"
import { TooltipProvider } from "@/components/ui/tooltip"
import { Toaster } from "@/components/ui/sonner"
import { isDesktopViewport } from "@/hooks/use-media-query"
import { track } from "@/lib/track"
import { useAppStore } from "@/stores/app-store"
import "./index.css"
import App from "./App.tsx"

track("visit", isDesktopViewport() ? "desktop" : "mobile")

// A deploy replaces the hashed chunks, so a page open across one fails to
// load a lazy tab; reload for the new ones. The timestamp stops a loop when
// the reload fails too, and ReloadBoundary covers that case.
const PRELOAD_RELOAD_KEY = "preload-error-reload"
window.addEventListener("vite:preloadError", (e) => {
  const last = Number(sessionStorage.getItem(PRELOAD_RELOAD_KEY) ?? 0)
  if (Date.now() - last < 10_000) return
  sessionStorage.setItem(PRELOAD_RELOAD_KEY, String(Date.now()))
  e.preventDefault()
  location.reload()
})
// On the store rather than TabNav, so programmatic tab switches count too.
useAppStore.subscribe((state, prev) => {
  if (state.tab !== prev.tab) track("tab", state.tab)
})

const queryClient = new QueryClient({
  queryCache: new QueryCache({
    onError: (error) => {
      toast.error(error.message)
    },
  }),
  mutationCache: new MutationCache({
    onError: (error, _variables, _context, mutation) => {
      if (mutation.options.meta?.handlesError) return
      toast.error(error.message)
    },
  }),
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      retry: 1,
    },
  },
})

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <TooltipProvider>
        <App />
        <Toaster />
      </TooltipProvider>
    </QueryClientProvider>
  </StrictMode>
)
