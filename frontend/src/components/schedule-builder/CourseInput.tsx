import { useState } from "react"
import { ChevronsUpDown, Plus } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { CoursePreview } from "./CoursePreview"
import { CRNPreview } from "./CRNPreview"
import { useCourse, useCRN, useSubjects } from "@/hooks/use-api"
import type { Subject } from "@/lib/api"
import { cn, decodeHtmlEntities } from "@/lib/utils"

type InputMode = "subject" | "crn"

interface CourseInputProps {
  term: string
  selectedSubject: string
  onSubjectChange: (subject: string) => void
  onAddCourse: (course: {
    subject: string
    courseNumber: string
    title: string
  }) => void
  onAddCrn: (section: {
    crn: string
    term: string
    subject: string
    courseNumber: string
    title: string
    instructor: string
  }) => void
}

export function CourseInput({
  term,
  selectedSubject,
  onSubjectChange,
  onAddCourse,
  onAddCrn,
}: CourseInputProps) {
  const [inputMode, setInputMode] = useState<InputMode>("subject")
  const [subjectOpen, setSubjectOpen] = useState(false)
  const [numberInput, setNumberInput] = useState("")
  const [crnInput, setCrnInput] = useState("")

  const { data: subjectsData } = useSubjects(term)
  const subjects = subjectsData?.subjects ?? []

  // Course lookup
  const numberTrimmed = numberInput.trim().toUpperCase()
  const courseNumberPattern = /^\d{3}[A-Za-z]?$/
  const isValidCourseNumber = courseNumberPattern.test(numberTrimmed)
  const { data: courseData, isFetching: courseFetching } = useCourse(
    term,
    selectedSubject && isValidCourseNumber ? selectedSubject : "",
    selectedSubject && isValidCourseNumber ? numberTrimmed : ""
  )

  // CRN lookup
  const crnTrimmed = crnInput.trim()
  const isValidCrnFormat = /^\d{5}$/.test(crnTrimmed)
  const { data: crnData, isFetching: crnFetching } = useCRN(
    isValidCrnFormat ? crnTrimmed : "",
    term
  )

  const canAddSubject =
    selectedSubject && isValidCourseNumber && courseData?.course
  const canAddCrn = isValidCrnFormat && crnData?.section
  const canAdd = inputMode === "subject" ? canAddSubject : canAddCrn

  const handleAdd = () => {
    if (inputMode === "subject" && canAddSubject && courseData?.course) {
      onAddCourse(courseData.course)
      setNumberInput("")
    } else if (inputMode === "crn" && canAddCrn && crnData?.section) {
      onAddCrn(crnData.section)
      setCrnInput("")
    }
  }

  return (
    <div className="space-y-3 p-4">
      <div className="flex items-center justify-between">
        <Label>Add Course</Label>
        <TooltipProvider delayDuration={300}>
          <div className="flex gap-1 text-sm">
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  aria-pressed={inputMode === "subject"}
                  aria-label="Add course by subject"
                  className={cn(
                    "rounded px-2 py-0.5 transition-colors",
                    inputMode === "subject"
                      ? "bg-accent text-accent-foreground"
                      : "text-muted-foreground hover:text-foreground"
                  )}
                  onClick={() => setInputMode("subject")}
                >
                  Subject
                </button>
              </TooltipTrigger>
              <TooltipContent>
                Search by subject code and course number (e.g., CSCI 241)
              </TooltipContent>
            </Tooltip>
            <span className="text-muted-foreground">|</span>
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  aria-pressed={inputMode === "crn"}
                  aria-label="Add course by CRN"
                  className={cn(
                    "rounded px-2 py-0.5 transition-colors",
                    inputMode === "crn"
                      ? "bg-accent text-accent-foreground"
                      : "text-muted-foreground hover:text-foreground"
                  )}
                  onClick={() => setInputMode("crn")}
                >
                  CRN
                </button>
              </TooltipTrigger>
              <TooltipContent>
                Add a specific section by its Course Reference Number
              </TooltipContent>
            </Tooltip>
          </div>
        </TooltipProvider>
      </div>

      {inputMode === "subject" ? (
        <SubjectInput
          term={term}
          subjects={subjects}
          selectedSubject={selectedSubject}
          onSubjectChange={onSubjectChange}
          subjectOpen={subjectOpen}
          setSubjectOpen={setSubjectOpen}
          numberInput={numberInput}
          setNumberInput={setNumberInput}
          isValidCourseNumber={isValidCourseNumber}
          courseData={courseData}
          courseFetching={courseFetching}
          canAdd={!!canAdd}
          onAdd={handleAdd}
        />
      ) : (
        <CRNInput
          term={term}
          crnInput={crnInput}
          setCrnInput={setCrnInput}
          isValidCrnFormat={isValidCrnFormat}
          crnData={crnData}
          crnFetching={crnFetching}
          canAdd={!!canAdd}
          onAdd={handleAdd}
        />
      )}
    </div>
  )
}

interface SubjectInputProps {
  term: string
  subjects: Subject[]
  selectedSubject: string
  onSubjectChange: (subject: string) => void
  subjectOpen: boolean
  setSubjectOpen: (open: boolean) => void
  numberInput: string
  setNumberInput: (value: string) => void
  isValidCourseNumber: boolean
  courseData: ReturnType<typeof useCourse>["data"]
  courseFetching: boolean
  canAdd: boolean
  onAdd: () => void
}

function SubjectInput({
  term,
  subjects,
  selectedSubject,
  onSubjectChange,
  subjectOpen,
  setSubjectOpen,
  numberInput,
  setNumberInput,
  isValidCourseNumber,
  courseData,
  courseFetching,
  canAdd,
  onAdd,
}: SubjectInputProps) {
  const [numberFocused, setNumberFocused] = useState(false)

  const showPreview =
    numberFocused &&
    isValidCourseNumber &&
    selectedSubject !== "" &&
    (courseFetching || courseData !== undefined)

  return (
    <Popover open={showPreview} modal={false}>
      <PopoverTrigger asChild>
        <div className="flex gap-2">
          <Popover open={subjectOpen} onOpenChange={setSubjectOpen}>
            <PopoverTrigger asChild>
              <Button
                variant="outline"
                role="combobox"
                aria-expanded={subjectOpen}
                className="flex-1 justify-between px-2"
                disabled={!term}
              >
                <span className="truncate">{selectedSubject || "Subject"}</span>
                <ChevronsUpDown className="size-4 shrink-0 opacity-50" />
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-64 p-0" align="start">
              <Command>
                <CommandInput placeholder="Search subjects..." />
                <CommandList>
                  <CommandEmpty>No subject found.</CommandEmpty>
                  <CommandGroup>
                    {subjects.map((subject) => (
                      <CommandItem
                        key={subject.code}
                        value={`${subject.code} ${subject.name}`}
                        onSelect={() => {
                          onSubjectChange(subject.code)
                          setSubjectOpen(false)
                        }}
                        className={cn(
                          selectedSubject === subject.code && "bg-accent"
                        )}
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
          <Input
            placeholder="241"
            value={numberInput}
            onChange={(e) => setNumberInput(e.target.value)}
            onFocus={() => setNumberFocused(true)}
            onBlur={() => setNumberFocused(false)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && canAdd) onAdd()
              if (e.key === "Escape") e.currentTarget.blur()
            }}
            className="w-16 flex-none"
            maxLength={4}
          />
          <Button
            size="icon"
            variant="outline"
            onClick={onAdd}
            disabled={!canAdd}
          >
            <Plus className="size-4" />
          </Button>
        </div>
      </PopoverTrigger>
      <PopoverContent
        className="w-72 p-0"
        align="start"
        onOpenAutoFocus={(e) => e.preventDefault()}
        onCloseAutoFocus={(e) => e.preventDefault()}
      >
        <CoursePreview courseData={courseData} isLoading={courseFetching} />
      </PopoverContent>
    </Popover>
  )
}

interface CRNInputProps {
  term: string
  crnInput: string
  setCrnInput: (value: string) => void
  isValidCrnFormat: boolean
  crnData: ReturnType<typeof useCRN>["data"]
  crnFetching: boolean
  canAdd: boolean
  onAdd: () => void
}

function CRNInput({
  term,
  crnInput,
  setCrnInput,
  isValidCrnFormat,
  crnData,
  crnFetching,
  canAdd,
  onAdd,
}: CRNInputProps) {
  const [crnFocused, setCrnFocused] = useState(false)

  const showPreview =
    crnFocused && isValidCrnFormat && (crnFetching || crnData !== undefined)

  return (
    <Popover open={showPreview} modal={false}>
      <PopoverTrigger asChild>
        <div className="flex gap-2">
          <Input
            placeholder="12345"
            value={crnInput}
            onChange={(e) => setCrnInput(e.target.value.replace(/\D/g, ""))}
            onFocus={() => setCrnFocused(true)}
            onBlur={() => setCrnFocused(false)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && canAdd) onAdd()
              if (e.key === "Escape") e.currentTarget.blur()
            }}
            inputMode="numeric"
            className="flex-1"
            maxLength={5}
          />
          <Button
            size="icon"
            variant="outline"
            onClick={onAdd}
            disabled={!canAdd}
          >
            <Plus className="size-4" />
          </Button>
        </div>
      </PopoverTrigger>
      <PopoverContent
        className="w-72 p-0"
        align="start"
        onOpenAutoFocus={(e) => e.preventDefault()}
        onCloseAutoFocus={(e) => e.preventDefault()}
      >
        <TooltipProvider delayDuration={300}>
          <CRNPreview crnData={crnData} isLoading={crnFetching} currentTerm={term} />
        </TooltipProvider>
      </PopoverContent>
    </Popover>
  )
}
