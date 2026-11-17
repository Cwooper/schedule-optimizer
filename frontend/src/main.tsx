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
