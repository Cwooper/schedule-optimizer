import { useEffect, useRef } from "react"
import { useAppStore } from "@/stores/app-store"
import { CourseStats } from "./CourseStats"
import { SectionStats } from "./SectionStats"
import { StatsPickers } from "./StatsPickers"

export function StatsView() {
  const { subject, number, term, crn } = useAppStore((s) => s.statsSelection)
  const scroller = useRef<HTMLDivElement>(null)

  useEffect(() => {
    scroller.current?.scrollTo({ top: 0 })
  }, [subject, number, term, crn])

  let body = null
  if (term && crn) body = <SectionStats term={term} crn={crn} />
  else if (subject && number)
    body = <CourseStats subject={subject} number={number} />

  return (
    <div ref={scroller} className="scrollbar-styled h-full overflow-y-auto">
      <div className="mx-auto max-w-[1180px] px-4 pt-5 pb-7 text-sm sm:px-6">
        <StatsPickers />
        {body}
      </div>
    </div>
  )
}
