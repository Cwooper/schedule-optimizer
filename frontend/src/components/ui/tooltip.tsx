import * as React from "react"
import * as TooltipPrimitive from "@radix-ui/react-tooltip"

import { cn } from "@/lib/utils"

// Context for passing touch handlers from Tooltip to TooltipTrigger
interface TooltipTouchContextValue {
  triggerRef: React.RefObject<HTMLElement | null>
  onTouchEnd: (e: React.TouchEvent) => void
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
  ...props
}: TooltipProps) {
  const [uncontrolledOpen, setUncontrolledOpen] = React.useState(false)
  const triggerRef = React.useRef<HTMLElement | null>(null)
  const blockCloseRef = React.useRef(false)

  const isControlled = controlledOpen !== undefined
  const open = isControlled ? controlledOpen : uncontrolledOpen
  const rawSetOpen = isControlled ? controlledOnOpenChange : setUncontrolledOpen

  // Wrap setOpen to block Radix's close-on-activation during touch
  const setOpen = React.useCallback((newOpen: boolean) => {
    if (blockCloseRef.current && !newOpen) return
    rawSetOpen?.(newOpen)
  }, [rawSetOpen])

  // Toggle tooltip on touch (mobile support)
  const handleTouchEnd = React.useCallback(() => {
    blockCloseRef.current = true
    const nextOpen = !open
    rawSetOpen?.(nextOpen)
    // Keep blocking Radix's close attempts until handlers settle
    setTimeout(() => {
      blockCloseRef.current = false
    }, 50)
  }, [open, rawSetOpen])

  // Close tooltip when touching outside the trigger
  React.useEffect(() => {
    if (!open) return

    const handleOutsideTouch = (e: TouchEvent) => {
      if (triggerRef.current && !triggerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }

    // Small delay to prevent the opening touch from immediately closing
    const timeoutId = setTimeout(() => {
      document.addEventListener("touchstart", handleOutsideTouch)
    }, 100)

    return () => {
      clearTimeout(timeoutId)
      document.removeEventListener("touchstart", handleOutsideTouch)
    }
  }, [open, setOpen])

  const touchContextValue = React.useMemo(
    () => ({ triggerRef, onTouchEnd: handleTouchEnd }),
    [handleTouchEnd]
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

const TooltipTrigger = React.forwardRef<
  HTMLButtonElement,
  React.ComponentProps<typeof TooltipPrimitive.Trigger>
>(({ onTouchEnd: onTouchEndProp, ...props }, forwardedRef) => {
  const touchContext = React.useContext(TooltipTouchContext)

  // Merge the forwarded ref with the context's triggerRef
  const mergedRef = mergeRefs(forwardedRef, touchContext?.triggerRef)

  // Combine touch handlers - stopPropagation prevents bubbling to parent touch handlers
  const handleTouchEnd = (e: React.TouchEvent<HTMLButtonElement>) => {
    e.stopPropagation()
    touchContext?.onTouchEnd(e)
    onTouchEndProp?.(e)
  }

  return (
    <TooltipPrimitive.Trigger
      ref={mergedRef}
      data-slot="tooltip-trigger"
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
