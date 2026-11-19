import { useState } from "react"
import { ChevronDown, ChevronRight } from "lucide-react"
import { useStatsCourse } from "@/hooks/use-api"
import type { StatsCourse, StatsProfessorRow } from "@/lib/api"
import {
  currentTermCode,
  distributionColumns,
  formatMetric,
  formatPct,
  surname,
  termName,
  metricValue,
} from "@/lib/stats"
import { cn } from "@/lib/utils"
import { useAppStore } from "@/stores/app-store"
import { DistributionChart } from "./charts/DistributionChart"
import { DashedSquare } from "./charts/DashedSquare"
import { QuartersGrid } from "./charts/QuartersGrid"
import { TermChart } from "./charts/TermChart"
import {
  Card,
  DataRangeInfo,
  LinkButton,
  type RowScale,
  StatCells,
  StatCols,
  StatHeaders,
  StatTile,
  StatsMessage,
} from "./parts"
import { TD, TH } from "./table"

const HISTORY_PREVIEW = 4

export function CourseStats({
  subject,
  number,
}: {
  subject: string
  number: string
}) {
  const { data, error } = useStatsCourse(subject, number)
  if (error) return <StatsMessage error={error} />
  if (!data) return null
  // Keyed so a cached course does not inherit the last one's open rows.
  return <CourseBody key={`${data.subject} ${data.number}`} course={data} />
}

function CourseBody({ course }: { course: StatsCourse }) {
  const { metric, summary } = course
  const code = `${course.subject} ${course.number}`
  const columns = distributionColumns(summary.grades)
  const scale: RowScale = {
    metric,
    columns,
    mean: metricValue(summary, metric),
  }

  return (
    <>
      <div className="mt-5.5 mb-3.5 flex flex-wrap items-baseline gap-x-2.5">
        <h1 className="text-[26px] leading-8 font-bold tracking-tight">
          {code}
        </h1>
        <span className="text-muted-foreground text-base">
          {course.title}
          <DataRangeInfo />
        </span>
      </div>

      <div className="mb-3 grid grid-cols-2 gap-3 sm:grid-cols-3 lg:auto-cols-fr lg:grid-flow-col lg:grid-cols-none">
        {metric === "gpa" && (
          <StatTile
            label="GPA"
            value={summary.gpa != null ? formatMetric("gpa", summary.gpa) : "-"}
          />
        )}
        <StatTile
          label="Pass"
          value={summary.pass != null ? formatPct(summary.pass) : "-"}
        />
        <StatTile
          label="Withdrew"
          value={
            summary.withdrew != null ? formatPct(summary.withdrew, 1) : "-"
          }
        />
        <StatTile label="Students" value={summary.students.toLocaleString()} />
        <StatTile
          label="Sections"
          value={course.sections.toLocaleString()}
          detail={`${course.professors} ${course.professors === 1 ? "professor" : "professors"}`}
        />
      </div>

      <div className="mb-3 grid grid-cols-12 gap-3">
        <Card title="Grade distribution" className="col-span-12 lg:col-span-5">
          <DistributionChart
            series={summary.grades && { label: code, grades: summary.grades }}
            columns={columns}
          />
        </Card>
        <Card
          title={metric === "gpa" ? "GPA by term" : "Pass by term"}
          className="col-span-12 md:col-span-7 lg:col-span-4"
        >
          <TermChart terms={course.terms} mean={scale.mean} metric={metric} />
        </Card>
        <Card
          title="Quarters"
          className="col-span-12 flex flex-col md:col-span-5 lg:col-span-3"
        >
          <QuartersGrid course={course} currentTerm={currentTermCode()} />
          {course.upcoming.length > 0 && (
            <div className="mt-auto flex flex-col gap-1.5 border-t pt-2.5">
              {course.upcoming.map((u) => (
                <div key={u.term} className="flex items-center gap-2">
                  <DashedSquare
                    size={14}
                    radius={3}
                    dashesPerSide={2}
                    className="shrink-0"
                  />
                  <span className="font-semibold whitespace-nowrap">
                    {termName(u.term)}
                  </span>
                  <span
                    className="text-muted-foreground ml-auto truncate"
                    title={u.professors.join("; ")}
                  >
                    {u.professors.map(surname).join(", ")}
                  </span>
                </div>
              ))}
            </div>
          )}
        </Card>
      </div>

      <Card title="Professors" className="pb-0.5">
        <div className="-mx-4 overflow-x-auto px-4">
          <table className="w-full min-w-[1000px] table-fixed border-collapse text-sm tabular-nums">
            <colgroup>
              <col />
              <col className="w-[76px]" />
              <col className="w-[76px]" />
              <StatCols metric={metric} />
              <col className="w-[120px]" />
              <col className="w-[120px]" />
            </colgroup>
            <thead>
              <tr>
                <th className={cn(TH, "pl-9")}>Professor</th>
                <th className={cn(TH, "text-right")}>Sections</th>
                <th className={cn(TH, "text-right")}>Students</th>
                <StatHeaders metric={metric} />
                <th className={TH}>Last taught</th>
                <th className={TH}>Next</th>
              </tr>
            </thead>
            <tbody className="[&>tr:last-child>td]:border-b-0">
              {course.professorRows.map((p) => (
                <ProfessorRows key={p.name} prof={p} scale={scale} />
              ))}
            </tbody>
          </table>
        </div>
      </Card>
    </>
  )
}

function ProfessorRows({
  prof,
  scale,
}: {
  prof: StatsProfessorRow
  scale: RowScale
}) {
  const setSelection = useAppStore((s) => s.setStatsSelection)
  const [open, setOpen] = useState(false)
  const [showAll, setShowAll] = useState(false)
  const shown = showAll ? prof.history : prof.history.slice(0, HISTORY_PREVIEW)
  const hidden = prof.history.length - shown.length
  const Chevron = open ? ChevronDown : ChevronRight
  const sub = cn(TD, "bg-muted border-b-background h-10")

  return (
    <>
      <tr>
        <td className={TD}>
          {/* TODO: link the name to the professor view once it exists (issue #27). */}
          <button
            type="button"
            onClick={() => setOpen(!open)}
            aria-expanded={open}
            className="flex max-w-full cursor-pointer items-center font-medium"
          >
            <Chevron
              className="text-muted-foreground mr-1.5 size-3.5 w-5 shrink-0"
              aria-hidden
            />
            <span
              className={cn("truncate", !prof.name && "text-muted-foreground")}
              title={prof.name || undefined}
            >
              {prof.name || "Unlisted"}
            </span>
          </button>
        </td>
        <td className={cn(TD, "text-right")}>{prof.sections}</td>
        <td className={cn(TD, "text-right")}>
          {prof.students.toLocaleString()}
        </td>
        <StatCells stats={prof} scale={scale} />
        <td className={TD}>{prof.lastTaught && termName(prof.lastTaught)}</td>
        <td className={TD}>{prof.nextTerm && termName(prof.nextTerm)}</td>
      </tr>
      {open &&
        shown.map((s) => (
          <tr key={`${s.term}${s.crn}`}>
            <td className={cn(sub, "pl-10")}>
              <LinkButton
                onClick={() => setSelection({ term: s.term, crn: s.crn })}
              >
                {termName(s.term)}
              </LinkButton>
              <span className="text-muted-foreground ml-2.5">{s.crn}</span>
            </td>
            <td className={sub} />
            <td className={cn(sub, "text-right")}>{s.students}</td>
            <StatCells
              stats={s}
              scale={scale}
              cellClass="bg-muted border-b-background h-10"
            />
            <td className={sub} />
            <td className={sub} />
          </tr>
        ))}
      {open && hidden > 0 && (
        <tr>
          <td
            className={cn(sub, "pl-10")}
            colSpan={scale.metric === "gpa" ? 10 : 9}
          >
            <button
              type="button"
              onClick={() => setShowAll(true)}
              className="text-muted-foreground hover:text-foreground cursor-pointer"
            >
              + {hidden} more
            </button>
          </td>
        </tr>
      )}
    </>
  )
}
