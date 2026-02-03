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
import "./index.css"
import App from "./App.tsx"

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
