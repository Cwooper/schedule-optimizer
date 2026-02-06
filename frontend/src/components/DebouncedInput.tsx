import { memo } from "react"
import { Input } from "@/components/ui/input"
import { useDebouncedState } from "@/hooks/use-debounced-state"

interface DebouncedInputProps
  extends Omit<React.ComponentProps<typeof Input>, "value" | "onChange"> {
  storeValue: string
  onSync: (value: string) => void
  /** Called after flushing on Enter key */
  onSubmit?: () => void
  delay?: number
}

export const DebouncedInput = memo(function DebouncedInput({
  storeValue,
  onSync,
  onSubmit,
  delay = 200,
  onKeyDown,
  onBlur,
  ...inputProps
}: DebouncedInputProps) {
  const [local, setLocal, flush] = useDebouncedState(storeValue, onSync, delay)

  return (
    <Input
      value={local}
      onChange={(e) => setLocal(e.target.value)}
      onBlur={(e) => {
        flush()
        onBlur?.(e)
      }}
      onKeyDown={(e) => {
        if (e.key === "Enter" && onSubmit) {
          flush()
          onSubmit()
        }
        onKeyDown?.(e)
      }}
      {...inputProps}
    />
  )
})
