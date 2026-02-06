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
  className?: string
}

export function TermSelector({
  value,
  onChange,
  terms,
  isLoading = false,
  label = "Term",
  className,
}: TermSelectorProps) {
  return (
    <div className={className}>
      {label && (
        <Label htmlFor="term" className="mb-2 block">
          {label}
        </Label>
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
