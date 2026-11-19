import * as React from "react"
import * as TooltipPrimitive from "@radix-ui/react-tooltip"

import { cn } from "@/lib/utils"

// Context for passing touch handlers from Tooltip to TooltipTrigger
interface TooltipTouchContextValue {
  triggerRef: React.RefObject<HTMLElement | null>
  onTap: () => void
  onSwipe: () => void
}

const TooltipTouchContext = React.createContext<TooltipTouchContextValue | null>(null)

function TooltipProvider({
  delayDuration = 0,
  ...props
}: React.ComponentProps<typeof TooltipPrimitive.Provider>) {
  return (
    <TooltipPrimitive.Provider
      data-slot="tooltip-provider"
      delayDuration={delayDuration}
      {...props}
    />
  )
}

interface TooltipProps extends Omit<React.ComponentProps<typeof TooltipPrimitive.Root>, 'open' | 'onOpenChange'> {
  open?: boolean
  onOpenChange?: (open: boolean) => void
  /**
   * Whether a tap toggles the tooltip. Turn it off where a tap does
   * something else (an action button, a popover trigger), so its label
   * does not linger over the result.
   */
  touchToggle?: boolean
}

/**
 * Tooltip with mobile touch support.
 * - Desktop: normal hover behavior
 * - Mobile: tap trigger to toggle tooltip, tap outside to close
 */
function Tooltip({
  children,
  open: controlledOpen,
  onOpenChange: controlledOnOpenChange,
  touchToggle = true,
  ...props
}: TooltipProps) {
  const [uncontrolledOpen, setUncontrolledOpen] = React.useState(false)
  const triggerRef = React.useRef<HTMLElement | null>(null)
  const touchOpenRef = React.useRef(false)

  const isControlled = controlledOpen !== undefined
  const open = isControlled ? controlledOpen : uncontrolledOpen
  const rawSetOpen = isControlled ? controlledOnOpenChange : setUncontrolledOpen

  // A touch-opened tooltip ignores Radix's closes. They are built for a
  // mouse (pointerdown, click, scroll, leaving the trigger) and the events a
  // tap generates after touchend set them off, on iOS often well after it.
  // It closes on a touch elsewhere, a second tap, Escape, or the mouse.
  const setOpen = React.useCallback((newOpen: boolean) => {
    if (touchOpenRef.current && !newOpen) return
    rawSetOpen?.(newOpen)
  }, [rawSetOpen])

  const closeTouchOpened = React.useCallback(() => {
    touchOpenRef.current = false
    rawSetOpen?.(false)
  }, [rawSetOpen])

  const handleTap = React.useCallback(() => {
    if (open) {
      closeTouchOpened()
    } else {
      touchOpenRef.current = true
      rawSetOpen?.(true)
    }
  }, [open, rawSetOpen, closeTouchOpened])

  // A swipe from the trigger scrolls the page, so the tooltip would be left
  // floating over it.
  const handleSwipe = React.useCallback(() => {
    if (touchOpenRef.current) closeTouchOpened()
  }, [closeTouchOpened])

  React.useEffect(() => {
    if (!open) return

    const handleOutsideTouch = (e: TouchEvent) => {
      if (triggerRef.current && !triggerRef.current.contains(e.target as Node)) {
        closeTouchOpened()
      }
    }
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === "Escape" && touchOpenRef.current) closeTouchOpened()
    }
    const handlePointer = (e: PointerEvent) => {
      if (e.pointerType === "mouse" && touchOpenRef.current) closeTouchOpened()
    }

    // Small delay to prevent the opening touch from immediately closing
    const timeoutId = setTimeout(() => {
      document.addEventListener("touchstart", handleOutsideTouch)
    }, 100)
    document.addEventListener("keydown", handleKey)
    document.addEventListener("pointermove", handlePointer)
    document.addEventListener("pointerdown", handlePointer)

    return () => {
      clearTimeout(timeoutId)
      document.removeEventListener("touchstart", handleOutsideTouch)
      document.removeEventListener("keydown", handleKey)
      document.removeEventListener("pointermove", handlePointer)
      document.removeEventListener("pointerdown", handlePointer)
    }
  }, [open, closeTouchOpened])

  const touchContextValue = React.useMemo(
    () => (touchToggle ? { triggerRef, onTap: handleTap, onSwipe: handleSwipe } : null),
    [touchToggle, handleTap, handleSwipe]
  )

  return (
    <TooltipTouchContext.Provider value={touchContextValue}>
      <TooltipPrimitive.Root
        data-slot="tooltip"
        open={open}
        onOpenChange={setOpen}
        {...props}
      >
        {children}
      </TooltipPrimitive.Root>
    </TooltipTouchContext.Provider>
  )
}

// Utility to merge multiple refs into one callback ref
function mergeRefs<T>(...refs: (React.Ref<T> | undefined)[]): React.RefCallback<T> {
  return (value) => {
    refs.forEach((ref) => {
      if (typeof ref === "function") {
        ref(value)
      } else if (ref != null) {
        (ref as React.MutableRefObject<T | null>).current = value
      }
    })
  }
}

// How far a touch may travel and still count as a tap rather than a swipe.
const TAP_SLOP_PX = 10

const TooltipTrigger = React.forwardRef<
  HTMLButtonElement,
  React.ComponentProps<typeof TooltipPrimitive.Trigger>
>(({ onTouchStart: onTouchStartProp, onTouchMove: onTouchMoveProp, onTouchEnd: onTouchEndProp, ...props }, forwardedRef) => {
  const touchContext = React.useContext(TooltipTouchContext)
  const startRef = React.useRef<{ x: number; y: number } | null>(null)
  const movedRef = React.useRef(false)

  // Merge the forwarded ref with the context's triggerRef
  const mergedRef = mergeRefs(forwardedRef, touchContext?.triggerRef)

  const strayed = (t: React.Touch | undefined) => {
    const start = startRef.current
    return !start || !t || Math.hypot(t.clientX - start.x, t.clientY - start.y) > TAP_SLOP_PX
  }

  const handleTouchStart = (e: React.TouchEvent<HTMLButtonElement>) => {
    const t = e.touches[0]
    startRef.current = { x: t.clientX, y: t.clientY }
    movedRef.current = false
    onTouchStartProp?.(e)
  }

  const handleTouchMove = (e: React.TouchEvent<HTMLButtonElement>) => {
    if (strayed(e.touches[0])) movedRef.current = true
    onTouchMoveProp?.(e)
  }

  // stopPropagation prevents bubbling to parent touch handlers
  const handleTouchEnd = (e: React.TouchEvent<HTMLButtonElement>) => {
    e.stopPropagation()
    if (touchContext) {
      if (movedRef.current || strayed(e.changedTouches[0])) touchContext.onSwipe()
      else touchContext.onTap()
    }
    startRef.current = null
    onTouchEndProp?.(e)
  }

  return (
    <TooltipPrimitive.Trigger
      ref={mergedRef}
      data-slot="tooltip-trigger"
      onTouchStart={handleTouchStart}
      onTouchMove={handleTouchMove}
      onTouchEnd={handleTouchEnd}
      {...props}
    />
  )
})
TooltipTrigger.displayName = "TooltipTrigger"

function TooltipContent({
  className,
  sideOffset = 0,
  children,
  ...props
}: React.ComponentProps<typeof TooltipPrimitive.Content>) {
  return (
    <TooltipPrimitive.Portal>
      <TooltipPrimitive.Content
        data-slot="tooltip-content"
        sideOffset={sideOffset}
        className={cn(
          // Animations handled via CSS with @media (hover: hover) for mobile compat
          "tooltip-content bg-foreground text-background z-50 w-fit origin-(--radix-tooltip-content-transform-origin) rounded-md px-3 py-1.5 text-xs text-balance",
          className
        )}
        {...props}
      >
        {children}
        <TooltipPrimitive.Arrow className="bg-foreground fill-foreground z-50 size-2.5 translate-y-[calc(-50%_-_2px)] rotate-45 rounded-[2px]" />
      </TooltipPrimitive.Content>
    </TooltipPrimitive.Portal>
  )
}

export { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider }
