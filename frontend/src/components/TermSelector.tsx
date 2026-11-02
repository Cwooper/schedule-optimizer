import type { ReactNode } from "react"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import type { Term } from "@/lib/api"

interface TermSelectorProps {
  value: string
  onChange: (value: string) => void
  terms: Term[]
  isLoading?: boolean
  label?: string
  /** Right-aligned on the label row. */
  status?: ReactNode
  className?: string
}

export function TermSelector({
  value,
  onChange,
  terms,
  isLoading = false,
  label = "Term",
  status,
  className,
}: TermSelectorProps) {
  return (
    <div className={className}>
      {label && (
        <div className="mb-2 flex items-baseline justify-between gap-3">
          <Label htmlFor="term">{label}</Label>
          {status}
        </div>
      )}
      <Select value={value} onValueChange={onChange} disabled={isLoading}>
        <SelectTrigger id="term" className="w-full">
          <SelectValue placeholder="Select term..." />
        </SelectTrigger>
        <SelectContent>
          {terms.map((t) => (
            <SelectItem key={t.code} value={t.code}>
              {t.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  )
}
