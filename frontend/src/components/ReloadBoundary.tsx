import { Component, type ReactNode } from "react"
import { Button } from "@/components/ui/button"

/**
 * Catches a lazy tab that failed to load, most often because a deploy
 * removed the chunk an open page asked for, and offers a reload.
 */
export class ReloadBoundary extends Component<
  { children: ReactNode },
  { failed: boolean }
> {
  state = { failed: false }

  static getDerivedStateFromError() {
    return { failed: true }
  }

  render() {
    if (!this.state.failed) return this.props.children
    return (
      <div className="text-muted-foreground mt-10 flex flex-col items-center gap-3 text-sm">
        <p>This tab failed to load.</p>
        <Button variant="outline" size="sm" onClick={() => location.reload()}>
          Reload
        </Button>
      </div>
    )
  }
}
