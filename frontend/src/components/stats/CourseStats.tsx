import { useState } from "react"
import { ChevronDown, ChevronRight } from "lucide-react"
import { useStatsCourse } from "@/hooks/use-api"
import type { StatsCourse, StatsProfessorRow } from "@/lib/api"
import {
  axisDomain,
  currentTermCode,
  distributionColumns,
  formatPct,
  surname,
  termName,
  termShort,
  metricValue,
} from "@/lib/stats"
import { cn } from "@/lib/utils"
import { useAppStore } from "@/stores/app-store"
import { DistributionChart } from "./charts/DistributionChart"
import { QuartersGrid } from "./charts/QuartersGrid"
import { TermLineChart } from "./charts/TermLineChart"
import {
  Card,
  LinkButton,
  type RowScale,
  StatCells,
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
  return <CourseBody course={data} />
}

function CourseBody({ course }: { course: StatsCourse }) {
  const { metric, summary, level } = course
  const code = `${course.subject} ${course.number}`
  const levelName = level && `${course.subject} ${course.number[0]}00-level`
  const columns = distributionColumns(summary.grades)
  const rowValues = course.professorRows
    .flatMap((p) => [p, ...p.history])
    .map((s) => metricValue(s, metric))
  const scale: RowScale = {
    metric,
    columns,
    mean: metricValue(summary, metric),
    domain: axisDomain(
      metric,
      rowValues.filter((v): v is number => v != null)
    ),
  }
  const next = course.next

  return (
    <>
      <div className="mt-5.5 mb-3.5 flex flex-wrap items-baseline gap-x-2.5">
        <h1 className="text-[26px] leading-8 font-bold tracking-tight">
          {code}
        </h1>
        <span className="text-muted-foreground text-base">{course.title}</span>
      </div>

      <div className="mb-3 grid grid-cols-2 gap-3 sm:grid-cols-3 lg:auto-cols-fr lg:grid-flow-col lg:grid-cols-none">
        {metric === "gpa" && (
          <StatTile
            label="GPA"
            value={summary.gpa?.toFixed(2) ?? "-"}
            detail={
              levelName &&
              level.gpa != null &&
              `${levelName} ${level.gpa.toFixed(2)}`
            }
          />
        )}
        <StatTile
          label="Pass"
          value={summary.pass != null ? formatPct(summary.pass) : "-"}
          detail={
            levelName &&
            level.pass != null &&
            `${levelName} ${formatPct(level.pass)}`
          }
        />
        <StatTile
          label="Withdrew"
          value={
            summary.withdrew != null ? formatPct(summary.withdrew, 1) : "-"
          }
          detail={
            levelName &&
            level.withdrew != null &&
            `${levelName} ${formatPct(level.withdrew, 1)}`
          }
        />
        <StatTile
          label="Students"
          value={summary.students.toLocaleString()}
          detail={`since ${termName(course.firstTerm)}`}
        />
        <StatTile
          label="Sections"
          value={course.sections.toLocaleString()}
          detail={`${course.professors} ${course.professors === 1 ? "professor" : "professors"}`}
        />
      </div>

      <div className="mb-3 grid grid-cols-12 gap-3">
        <Card title="Grade distribution" className="col-span-12 lg:col-span-5">
          <DistributionChart grades={summary.grades} columns={columns} />
        </Card>
        <Card
          title={metric === "gpa" ? "GPA by term" : "Pass by term"}
          className="col-span-12 md:col-span-7 lg:col-span-4"
        >
          <TermLineChart
            points={course.terms}
            mean={scale.mean}
            metric={metric}
          />
        </Card>
        <Card
          title="Quarters"
          className="col-span-12 flex flex-col md:col-span-5 lg:col-span-3"
        >
          <QuartersGrid
            offered={course.offered}
            next={next?.term}
            currentTerm={currentTermCode()}
            seasons={course.seasons}
            metric={metric}
          />
          {next && (
            <div className="mt-auto flex items-center gap-2 border-t pt-2.5">
              <svg width="14" height="14" className="shrink-0" aria-hidden>
                <rect
                  x="1"
                  y="1"
                  width="12"
                  height="12"
                  rx="3"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="1.5"
                  strokeDasharray="3 2"
                />
              </svg>
              <span className="font-semibold whitespace-nowrap">
                {termName(next.term)}
              </span>
              <span
                className="text-muted-foreground ml-auto truncate"
                title={next.professors.join("; ")}
              >
                {next.professors.map(surname).join(", ")}
              </span>
            </div>
          )}
        </Card>
      </div>

      <Card title="Professors" className="pb-0.5">
        <div className="-mx-4 overflow-x-auto px-4">
          <table className="w-full border-collapse text-sm tabular-nums">
            <thead>
              <tr>
                <th className={TH} aria-label="Professor" />
                <th className={cn(TH, "text-right")}>Sections</th>
                <th className={cn(TH, "text-right")}>Students</th>
                <StatHeaders metric={metric} />
                <th className={TH}>Last taught</th>
              </tr>
            </thead>
            <tbody className="[&>tr:last-child>td]:border-b-0">
              {course.professorRows.map((p) => (
                <ProfessorRows
                  key={p.name}
                  prof={p}
                  scale={scale}
                  nextTerm={next?.term}
                />
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
  nextTerm,
}: {
  prof: StatsProfessorRow
  scale: RowScale
  nextTerm?: string
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
          <span className="flex items-center">
            {/* TODO: link the name to the professor view once it exists (issue #27). */}
            <button
              type="button"
              onClick={() => setOpen(!open)}
              aria-expanded={open}
              className="flex cursor-pointer items-center font-medium"
            >
              <Chevron
                className="text-muted-foreground mr-1.5 size-3.5 w-5"
                aria-hidden
              />
              {prof.name || (
                <span className="text-muted-foreground">Unlisted</span>
              )}
            </button>
            {prof.teachesNext && nextTerm && (
              <span className="border-muted-foreground ml-2 inline-flex h-5 items-center rounded-full border border-dashed px-[7px] text-[11px] font-medium">
                {termShort(nextTerm)}
              </span>
            )}
          </span>
        </td>
        <td className={cn(TD, "text-right")}>{prof.sections}</td>
        <td className={cn(TD, "text-right")}>
          {prof.students.toLocaleString()}
        </td>
        <StatCells stats={prof} scale={scale} />
        <td className={TD}>{prof.lastTaught && termName(prof.lastTaught)}</td>
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
          </tr>
        ))}
      {open && hidden > 0 && (
        <tr>
          <td
            className={cn(sub, "pl-10")}
            colSpan={scale.metric === "gpa" ? 9 : 8}
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
