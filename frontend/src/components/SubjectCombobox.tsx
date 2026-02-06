import { useState, memo } from "react"
import { ChevronsUpDown } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { cn, decodeHtmlEntities } from "@/lib/utils"
import type { Subject } from "@/lib/api"

interface SubjectComboboxProps {
  subjects: Subject[]
  value: string
  onChange: (value: string) => void
  placeholder?: string
  disabled?: boolean
  showAnyOption?: boolean
  className?: string
}

export const SubjectCombobox = memo(function SubjectCombobox({
  subjects,
  value,
  onChange,
  placeholder = "Subject",
  disabled = false,
  showAnyOption = false,
  className,
}: SubjectComboboxProps) {
  const [open, setOpen] = useState(false)

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          aria-label={value ? `Subject: ${value}` : "Select subject"}
          disabled={disabled}
          className={cn("w-full justify-between", className)}
        >
          <span className="truncate">{value || placeholder}</span>
          <ChevronsUpDown className="size-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-64 p-0" align="start">
        <Command>
          <CommandInput placeholder="Search subjects..." />
          <CommandList>
            <CommandEmpty>No subject found.</CommandEmpty>
            <CommandGroup>
              {showAnyOption && (
                <CommandItem
                  value=""
                  onSelect={() => {
                    onChange("")
                    setOpen(false)
                  }}
                  className={cn(!value && "bg-accent")}
                >
                  <span className="text-muted-foreground">Any subject</span>
                </CommandItem>
              )}
              {subjects.map((subject) => (
                <CommandItem
                  key={subject.code}
                  value={`${subject.code} ${subject.name}`}
                  onSelect={() => {
                    onChange(subject.code)
                    setOpen(false)
                  }}
                  className={cn(value === subject.code && "bg-accent")}
                >
                  <span className="w-12 shrink-0 font-medium whitespace-nowrap">
                    {subject.code}
                  </span>
                  <span className="text-muted-foreground truncate">
                    {decodeHtmlEntities(subject.name)}
                  </span>
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
})
