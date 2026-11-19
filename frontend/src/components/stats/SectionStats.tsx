import { ChevronLeft } from "lucide-react"
import { useStatsSection } from "@/hooks/use-api"
import type { StatsSection } from "@/lib/api"
import { formatMonthDay, formatWeekday } from "@/lib/dates"
import { formatLocation, formatMeetingTime, formatTime } from "@/lib/meetings"
import {
  axisDomain,
  distributionColumns,
  formatPct,
  surname,
  termName,
  metricValue,
} from "@/lib/stats"
import { cn } from "@/lib/utils"
import { useAppStore } from "@/stores/app-store"
import { DistributionChart } from "./charts/DistributionChart"
import { DotStrip } from "./charts/DotStrip"
import {
  Card,
  LegendItem,
  LinkButton,
  type RowScale,
  StatCells,
  StatHeaders,
  StatTile,
  StatsMessage,
  TileOf,
} from "./parts"
import { TD, TH } from "./table"

export function SectionStats({ term, crn }: { term: string; crn: string }) {
  const { data, error } = useStatsSection(term, crn)
  if (error) return <StatsMessage error={error} />
  if (!data) return null
  return <SectionBody section={data} />
}

function SectionBody({ section: v }: { section: StatsSection }) {
  const setSelection = useAppStore((s) => s.setStatsSelection)
  const { metric, section, course } = v
  const code = `${v.subject} ${v.number}`
  const term = termName(v.term)
  const columns = distributionColumns(course.grades)
  const scale: RowScale = {
    metric,
    columns,
    mean: metricValue(course, metric),
    domain: axisDomain(
      metric,
      v.dots.map((d) => d.value)
    ),
  }
  const capacity = v.class?.capacity ?? 0
  const rate = (value: number | null, digits = 0) =>
    value == null ? "-" : formatPct(value, digits)

  return (
    <>
      <div className="text-muted-foreground mt-5 flex items-center text-[13px]">
        <LinkButton
          onClick={() => setSelection({ term: "", crn: "" })}
          className="flex items-center gap-1"
        >
          <ChevronLeft className="size-3.5" aria-hidden />
          {code}
        </LinkButton>
      </div>
      <div className="mt-1.5 mb-3.5 flex flex-wrap items-baseline gap-x-2.5">
        <h1 className="text-[26px] leading-8 font-bold tracking-tight">
          {term}
        </h1>
        {/* TODO: link the professor to their view once it exists (issue #27). */}
        {v.professor && (
          <span className="text-base font-medium">{v.professor}</span>
        )}
        <span className="text-muted-foreground text-base">CRN {v.crn}</span>
      </div>

      <div className="mb-3 grid grid-cols-2 gap-3 sm:grid-cols-3 lg:auto-cols-fr lg:grid-flow-col lg:grid-cols-none">
        {metric === "gpa" && (
          <StatTile
            label="GPA"
            value={section.gpa?.toFixed(2) ?? "-"}
            detail={course.gpa != null && `${code} ${course.gpa.toFixed(2)}`}
          />
        )}
        <StatTile
          label="Pass"
          value={rate(section.pass)}
          detail={`${code} ${rate(course.pass)}`}
        />
        <StatTile
          label="Withdrew"
          value={rate(section.withdrew, 1)}
          detail={`${code} ${rate(course.withdrew, 1)}`}
        />
        <StatTile
          label="Students"
          value={
            <>
              {section.students}
              {capacity > 0 && <TileOf>{capacity}</TileOf>}
            </>
          }
          detail={
            capacity > 0 &&
            `${Math.round((section.students / capacity) * 100)}% full`
          }
        >
          {capacity > 0 && (
            <div className="bg-muted mt-2 h-1.5 overflow-hidden rounded-[3px]">
              <div
                className="bg-foreground h-full"
                style={{
                  width: `${Math.min(100, (section.students / capacity) * 100)}%`,
                }}
              />
            </div>
          )}
        </StatTile>
        <StatTile
          label="Rank"
          value={
            v.rank ? (
              <>
                {v.rank}
                <TileOf>{v.dots.length}</TileOf>
              </>
            ) : (
              "-"
            )
          }
          detail={
            metric === "gpa" ? "sections by GPA" : "sections by pass rate"
          }
        />
      </div>

      <div className="mb-3 grid grid-cols-12 gap-3">
        <Card
          title="Grade distribution"
          className="col-span-12 lg:col-span-6"
          legend={
            <>
              {section.grades && (
                <LegendItem swatch="bg-foreground rounded-[3px]">
                  {term}
                </LegendItem>
              )}
              <LegendItem swatch="bg-chart-soft rounded-[3px]">
                {code}
              </LegendItem>
            </>
          }
        >
          <DistributionChart
            grades={section.grades}
            background={course.grades}
            columns={columns}
          />
        </Card>
        <Card
          title={`All ${code} sections`}
          className="col-span-12 lg:col-span-6"
          legend={
            <>
              {v.rank > 0 && (
                <LegendItem swatch="border-foreground rounded-full border-2">
                  This one
                </LegendItem>
              )}
              {v.professor && (
                <LegendItem swatch="bg-foreground/75 rounded-full">
                  {surname(v.professor)}
                </LegendItem>
              )}
              <LegendItem swatch="bg-chart-soft rounded-full">
                Others
              </LegendItem>
            </>
          }
        >
          <DotStrip dots={v.dots} current={v} metric={metric} />
        </Card>
      </div>

      <div className="grid grid-cols-12 gap-3">
        <Card
          title={`${term} sections`}
          className={cn("col-span-12 pb-0.5", v.class && "lg:col-span-8")}
        >
          <div className="-mx-4 overflow-x-auto px-4">
            <table className="w-full border-collapse text-sm tabular-nums">
              <thead>
                <tr>
                  <th className={TH}>CRN</th>
                  <th className={TH}>Professor</th>
                  <th className={cn(TH, "text-right")}>Students</th>
                  <StatHeaders metric={metric} />
                </tr>
              </thead>
              <tbody className="[&>tr:last-child>td]:border-b-0">
                {v.siblings.map((s) => {
                  const current = s.crn === v.crn
                  return (
                    <tr key={s.crn} className={cn(current && "font-semibold")}>
                      <td className={TD}>
                        {current ? (
                          s.crn
                        ) : (
                          <LinkButton
                            onClick={() => setSelection({ crn: s.crn })}
                          >
                            {s.crn}
                          </LinkButton>
                        )}
                      </td>
                      <td
                        className={cn(
                          TD,
                          !s.professor && "text-muted-foreground"
                        )}
                      >
                        {s.professor || "Unlisted"}
                      </td>
                      <td className={cn(TD, "text-right")}>{s.students}</td>
                      <StatCells stats={s} scale={scale} />
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </Card>
        {v.class && <ClassCard details={v.class} />}
      </div>
    </>
  )
}

function ClassCard({
  details,
}: {
  details: NonNullable<StatsSection["class"]>
}) {
  const rooms = [
    ...new Set(details.meetings.map(formatLocation).filter(Boolean)),
  ]
  const final = details.final
  const rows: [string, string[]][] = [
    ["Meets", details.meetings.map(formatMeetingTime)],
    ["Room", rooms as string[]],
    [
      "Final",
      final
        ? [
            `${formatWeekday(final.date)} ${formatMonthDay(final.date)}, ${formatTime(final.start)}`,
          ]
        : [],
    ],
    ["Credits", details.credits ? [String(details.credits)] : []],
  ]
  return (
    <Card title="Class" className="col-span-12 lg:col-span-4">
      <dl className="grid grid-cols-[auto_1fr]">
        {rows
          .filter(([, lines]) => lines.length > 0)
          .map(([label, lines]) => (
            <div
              key={label}
              className="col-span-2 grid grid-cols-subgrid border-b py-[9px] last:border-b-0"
            >
              <dt className="text-muted-foreground pr-4">{label}</dt>
              <dd className="text-right font-medium tabular-nums">
                {lines.map((l) => (
                  <div key={l}>{l}</div>
                ))}
              </dd>
            </div>
          ))}
      </dl>
    </Card>
  )
}
