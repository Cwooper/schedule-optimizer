import { useMemo, useState } from "react"
import { ChevronsUpDown } from "lucide-react"
import { SubjectCombobox } from "@/components/SubjectCombobox"
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
import { useStatsCourses } from "@/hooks/use-api"
import { cn } from "@/lib/utils"
import { useAppStore } from "@/stores/app-store"

function Field({
  label,
  children,
  className,
}: {
  label: string
  children: React.ReactNode
  className?: string
}) {
  return (
    <div className={cn("flex min-w-0 flex-col gap-1", className)}>
      <span className="text-muted-foreground pl-0.5 text-xs font-medium">
        {label}
      </span>
      {children}
    </div>
  )
}

export function StatsPickers() {
  const { data } = useStatsCourses()
  const selection = useAppStore((s) => s.statsSelection)
  const setSelection = useAppStore((s) => s.setStatsSelection)
  const [courseOpen, setCourseOpen] = useState(false)

  const courses = useMemo(
    () => data?.courses.filter((c) => c.subject === selection.subject) ?? [],
    [data, selection.subject]
  )
  const current = courses.find((c) => c.number === selection.number)

  return (
    <div className="flex items-end gap-2.5">
      <Field label="Subject" className="w-[130px] shrink-0">
        <SubjectCombobox
          subjects={data?.subjects ?? []}
          value={selection.subject}
          onChange={(subject) => {
            if (subject === selection.subject) return
            setSelection({ subject, number: "", term: "", crn: "" })
            setCourseOpen(true)
          }}
          className="font-medium"
        />
      </Field>
      <Field label="Course" className="max-w-[270px] flex-1">
        <Popover open={courseOpen} onOpenChange={setCourseOpen}>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              role="combobox"
              aria-expanded={courseOpen}
              aria-label={
                current
                  ? `Course: ${current.number} ${current.title}`
                  : "Select course"
              }
              disabled={!selection.subject}
              className="w-full justify-between"
            >
              <span className="truncate">
                {current ? `${current.number} ${current.title}` : "Course"}
              </span>
              <ChevronsUpDown className="size-4 shrink-0 opacity-50" />
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-80 p-0" align="start">
            <Command>
              <CommandInput placeholder="Search courses..." />
              <CommandList>
                <CommandEmpty>No course found.</CommandEmpty>
                <CommandGroup>
                  {courses.map((c) => (
                    <CommandItem
                      key={c.number}
                      value={`${c.number} ${c.title}`}
                      onSelect={() => {
                        setSelection({ number: c.number, term: "", crn: "" })
                        setCourseOpen(false)
                      }}
                      className={cn(
                        c.number === selection.number && "bg-accent"
                      )}
                    >
                      <span className="w-12 shrink-0 font-medium">
                        {c.number}
                      </span>
                      <span className="text-muted-foreground truncate">
                        {c.title}
                      </span>
                    </CommandItem>
                  ))}
                </CommandGroup>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>
      </Field>
    </div>
  )
}
