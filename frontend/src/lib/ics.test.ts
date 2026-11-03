import { describe, it, expect } from "vitest"
import type { HydratedSection, MeetingTime, TermCalendar } from "./api"
import {
  buildIcs,
  DEFAULT_CALENDAR_EXPORT,
  escapeText,
  foldLine,
  hasIcsEvents,
  icsFilename,
  type CalendarExportOptions,
} from "./ics"

const fall: TermCalendar = {
  term: "202640",
  start: "2026-09-23",
  end: "2026-12-11",
  registrationOpens: "2026-05-12",
  estimated: false,
  registration: [
    { phase: "I", name: "Phase I - registration for continuing students", start: "2026-05-12", end: "2026-07-13" },
    { phase: "II", name: "Phase II - registration for continuing, new transfer, and post-baccalaureate students", start: "2026-08-25", end: "2026-09-22" },
    { phase: "III", name: "Phase III - open registration", start: "2026-09-23", end: "2026-09-29" },
  ],
  holidays: [
    { name: "Veterans Day", start: "2026-11-11", end: "2026-11-11" },
    { name: "Thanksgiving Break (Wed, Nov 25 - University open, no classes)", start: "2026-11-25", end: "2026-11-29" },
    { name: "Native American Heritage Day", start: "2026-11-27", end: "2026-11-27" },
  ],
  deadlines: [
    { name: "Last day to add a class without an override", title: "Add deadline", date: "2026-09-29", kind: "add" },
    {
      name: "Last day to withdraw from a class. Online registration via Web4U ends at 11:59 p.m.",
      title: "Withdraw deadline",
      date: "2026-11-08",
      kind: "withdraw",
    },
    { name: "Last day to change to-or-from Audit", title: "Last day to change to-or-from Audit", date: "2026-09-29", kind: "other" },
  ],
  finals: { start: "2026-12-07", end: "2026-12-11" },
}

function meeting(days: string, startTime: string, endTime: string, extra: Partial<MeetingTime> = {}): MeetingTime {
  return {
    days: ["U", "M", "T", "W", "R", "F", "S"].map((d) => days.includes(d)),
    startTime,
    endTime,
    building: "CF",
    room: "105",
    startDate: "2026-09-23",
    endDate: "2026-12-11",
    ...extra,
  }
}

function section(overrides: Partial<HydratedSection>): HydratedSection {
  return {
    crn: "41216",
    term: "202640",
    subject: "CSCI",
    courseNumber: "247",
    title: "Computer Systems I",
    credits: 5,
    instructor: "Ada Lovelace",
    meetingTimes: [],
    enrollment: 0,
    maxEnrollment: 0,
    seatsAvailable: 0,
    waitCount: 0,
    isOpen: true,
    ...overrides,
  }
}

const csci247 = section({
  meetingTimes: [meeting("MTWF", "0900", "0950"), meeting("M", "1200", "1350", { room: "164" })],
  final: { date: "2026-12-07", start: "0800", end: "1000", certain: false, reason: "lab" },
})

const allOff: CalendarExportOptions = {
  classes: false,
  finals: false,
  holidays: false,
  registration: { I: false, II: false, III: false },
  deadlines: false,
}

function build(sections: HydratedSection[], options: Partial<CalendarExportOptions>, calendar = fall): string {
  return buildIcs({
    calendar,
    sections,
    options: { ...allOff, ...options },
    now: new Date(Date.UTC(2026, 10, 2, 12, 0, 0)),
  })
}

/** Unfolds and splits into VEVENT property lists. */
function events(ics: string): string[][] {
  const lines = ics.replace(/\r\n /g, "").split("\r\n")
  const out: string[][] = []
  let current: string[] | null = null
  for (const line of lines) {
    if (line === "BEGIN:VEVENT") current = []
    else if (line === "END:VEVENT" && current) {
      out.push(current)
      current = null
    } else current?.push(line)
  }
  return out
}

function prop(event: string[], name: string): string[] {
  return event.filter((l) => l.startsWith(name + ":") || l.startsWith(name + ";"))
}

describe("buildIcs classes", () => {
  const [lecture, lab] = events(build([csci247], { classes: true }))

  it("starts each pattern on its first meeting day in Pacific time", () => {
    // Sep 23, 2026 is a Wednesday; the Monday lab first meets Sep 28.
    expect(prop(lecture, "DTSTART")).toEqual(["DTSTART;TZID=America/Los_Angeles:20260923T090000"])
    expect(prop(lecture, "DTEND")).toEqual(["DTEND;TZID=America/Los_Angeles:20260923T095000"])
    expect(prop(lab, "DTSTART")).toEqual(["DTSTART;TZID=America/Los_Angeles:20260928T120000"])
  })

  it("repeats weekly until the end of the Sunday before finals, in UTC", () => {
    // Dec 6 23:59:59 PST is Dec 7 07:59:59 UTC; the November DST change
    // does not move the 9:00 class because DTSTART is TZID-local.
    expect(prop(lecture, "RRULE")).toEqual(["RRULE:FREQ=WEEKLY;BYDAY=MO,TU,WE,FR;UNTIL=20261207T075959Z"])
  })

  it("excludes each class day inside a holiday once", () => {
    // Thanksgiving Break covers Wed and Fri; Native American Heritage Day
    // repeats that Friday.
    expect(prop(lecture, "EXDATE")).toEqual([
      "EXDATE;TZID=America/Los_Angeles:20261111T090000",
      "EXDATE;TZID=America/Los_Angeles:20261125T090000",
      "EXDATE;TZID=America/Los_Angeles:20261127T090000",
    ])
    expect(prop(lab, "EXDATE")).toEqual([])
  })

  it("gives stable UIDs and readable text", () => {
    expect(prop(lecture, "UID")).toEqual(["UID:202640-41216-MTWF-0900-0950-20260923-20261211@schedule-optimizer.cwooper.dev"])
    expect(prop(lecture, "SUMMARY")).toEqual(["SUMMARY:CSCI 247 - Computer Systems I"])
    expect(prop(lecture, "LOCATION")).toEqual(["LOCATION:CF 105"])
    expect(prop(lecture, "DESCRIPTION")).toEqual(["DESCRIPTION:CRN 41216\\nAda Lovelace"])
  })

  it("labels a lab meeting in its description, not its title", () => {
    const withLab = section({
      meetingTimes: [meeting("MWF", "0900", "0950"), meeting("R", "1400", "1550", { lab: true })],
    })
    const [lec, labEvent] = events(build([withLab], { classes: true }))
    expect(prop(lec, "DESCRIPTION")).toEqual(["DESCRIPTION:CRN 41216\\nAda Lovelace"])
    expect(prop(labEvent, "DESCRIPTION")).toEqual(["DESCRIPTION:Lab\\nCRN 41216\\nAda Lovelace"])
    expect(prop(labEvent, "SUMMARY")).toEqual(["SUMMARY:CSCI 247 - Computer Systems I"])
  })

  it("ends a short session at its own last day", () => {
    const short = section({ meetingTimes: [meeting("MW", "1000", "1050", { endDate: "2026-10-02" })] })
    // Oct 2 23:59:59 PDT is Oct 3 06:59:59 UTC.
    expect(prop(events(build([short], { classes: true }))[0], "RRULE")).toEqual([
      "RRULE:FREQ=WEEKLY;BYDAY=MO,WE;UNTIL=20261003T065959Z",
    ])
  })

  it("ends an evening class on either side of the November DST change", () => {
    // DST ends at 2:00 on Sun Nov 1, 2026: 23:59:59 is PDT on Oct 31 and PST
    // on Nov 1, so the 19:00 class on the last day falls inside UNTIL both times.
    const until = (endDate: string) => {
      const evening = section({ meetingTimes: [meeting("SU", "1900", "2150", { endDate })] })
      return prop(events(build([evening], { classes: true }))[0], "RRULE")[0]
    }
    expect(until("2026-10-31")).toContain("UNTIL=20261101T065959Z")
    expect(until("2026-11-01")).toContain("UNTIL=20261102T075959Z")
  })

  it("ends an evening class on either side of the March DST change", () => {
    // DST starts at 2:00 on Sun Mar 14, 2027: 23:59:59 is PST on Mar 13 and
    // PDT on Mar 14, the last class day before Winter 2027 finals.
    const winter2027 = { ...fall, term: "202710", start: "2027-01-05", end: "2027-03-19", holidays: [], finals: { start: "2027-03-15", end: "2027-03-19" } }
    const until = (endDate: string) => {
      const evening = section({ meetingTimes: [meeting("SU", "1900", "2150", { startDate: "2027-01-05", endDate })] })
      return prop(events(build([evening], { classes: true }, winter2027))[0], "RRULE")[0]
    }
    expect(until("2027-03-13")).toContain("UNTIL=20270314T075959Z")
    expect(until("2027-03-14")).toContain("UNTIL=20270315T065959Z")
  })

  it("merges rows that differ only by room into one event", () => {
    // ACCT 321 CRN 41526 meets in PH 210 and PH 210B at once.
    const split = section({
      meetingTimes: [
        meeting("MW", "1000", "1150", { building: "PH", room: "210" }),
        meeting("MW", "1000", "1150", { building: "PH", room: "210B" }),
      ],
    })
    const evs = events(build([split], { classes: true }))
    expect(evs).toHaveLength(1)
    expect(prop(evs[0], "LOCATION")).toEqual(["LOCATION:PH 210\\, PH 210B"])
  })

  it("keeps the sessions of a split term apart with distinct, stable UIDs", () => {
    // ECON 470 CRN 13702 (Winter 2026) meets W 10:00 twice with different dates.
    const winter2026 = { ...fall, term: "202610", start: "2026-01-06", end: "2026-03-20", holidays: [], finals: { start: "2026-03-16", end: "2026-03-20" } }
    const econ = section({
      crn: "13702",
      meetingTimes: [
        meeting("W", "1000", "1150", { startDate: "2026-01-06", endDate: "2026-01-20" }),
        meeting("W", "1000", "1150", { startDate: "2026-01-21", endDate: "2026-03-20" }),
      ],
    })
    const uids = () => events(build([econ], { classes: true }, winter2026)).map((e) => prop(e, "UID")[0])
    expect(uids()).toEqual([
      "UID:202610-13702-W-1000-1150-20260106-20260120@schedule-optimizer.cwooper.dev",
      "UID:202610-13702-W-1000-1150-20260121-20260320@schedule-optimizer.cwooper.dev",
    ])
    expect(uids()).toEqual(uids())
  })

  it("ends at the term's last day without finals", () => {
    const summer = { ...fall, finals: null }
    const rrule = prop(events(build([csci247], { classes: true }, summer))[0], "RRULE")
    expect(rrule).toEqual(["RRULE:FREQ=WEEKLY;BYDAY=MO,TU,WE,FR;UNTIL=20261212T075959Z"])
  })

  it("falls back to term dates for meetings without dates", () => {
    const undated = section({ meetingTimes: [meeting("TR", "1400", "1550", { startDate: undefined, endDate: undefined })] })
    const [e] = events(build([undated], { classes: true }))
    expect(prop(e, "DTSTART")).toEqual(["DTSTART;TZID=America/Los_Angeles:20260924T140000"])
    expect(prop(e, "RRULE")[0]).toContain("UNTIL=20261207T075959Z")
  })

  it("skips unscheduled meetings", () => {
    const async = section({ meetingTimes: [meeting("", "", "", { room: "ASNC" })] })
    expect(events(build([async], { classes: true }))).toEqual([])
  })
})

describe("buildIcs finals and WWU dates", () => {
  it("marks a guessed final and says why", () => {
    const [e] = events(build([csci247], { finals: true }))
    expect(prop(e, "SUMMARY")).toEqual(["SUMMARY:? CSCI 247 - Final Exam"])
    expect(prop(e, "DTSTART")).toEqual(["DTSTART;TZID=America/Los_Angeles:20261207T080000"])
    expect(prop(e, "DESCRIPTION")).toEqual(["DESCRIPTION:Has a lab"])
    expect(prop(e, "LOCATION")).toEqual([])
  })

  it("places a final in the rooms of the meeting it is based on", () => {
    const split = section({
      final: { date: "2026-12-08", start: "1030", end: "1230", certain: true, location: "PH 210, PH 210B" },
    })
    expect(prop(events(build([split], { finals: true }))[0], "LOCATION")).toEqual(["LOCATION:PH 210\\, PH 210B"])
  })

  it("titles a certain final without the mark", () => {
    const certain = section({ final: { date: "2026-12-08", start: "1030", end: "1230", certain: true } })
    const [e] = events(build([certain], { finals: true }))
    expect(prop(e, "SUMMARY")).toEqual(["SUMMARY:CSCI 247 - Final Exam"])
    expect(prop(e, "DESCRIPTION")).toEqual([])
  })

  it("adds holidays as all-day events with an exclusive end", () => {
    const ics = build([], { holidays: true })
    const thanksgiving = events(ics)[1]
    expect(prop(thanksgiving, "DTSTART")).toEqual(["DTSTART;VALUE=DATE:20261125"])
    expect(prop(thanksgiving, "DTEND")).toEqual(["DTEND;VALUE=DATE:20261130"])
    expect(prop(thanksgiving, "SUMMARY")).toEqual(["SUMMARY:F26 - Thanksgiving Break"])
    expect(prop(thanksgiving, "DESCRIPTION")).toEqual([
      "DESCRIPTION:Thanksgiving Break (Wed\\, Nov 25 - University open\\, no classes)",
    ])
    // All-day events need no time zone.
    expect(ics).not.toContain("VTIMEZONE")
  })

  it("adds only the chosen registration phases, on their opening day", () => {
    const evs = events(build([], { registration: { I: false, II: true, III: false } }))
    expect(evs.map((e) => [prop(e, "SUMMARY")[0], prop(e, "DTSTART")[0], prop(e, "DESCRIPTION")[0]])).toEqual([
      [
        "SUMMARY:F26 - Phase II registration",
        "DTSTART;VALUE=DATE:20260825",
        "DESCRIPTION:Phase II - registration for continuing\\, new transfer\\, and post-baccalaureate students",
      ],
    ])
  })

  it("adds add, drop, and withdraw deadlines only, titled short", () => {
    const evs = events(build([], { deadlines: true }))
    expect(evs.map((e) => [prop(e, "SUMMARY")[0], prop(e, "DESCRIPTION")[0]])).toEqual([
      ["SUMMARY:F26 - Add deadline", "DESCRIPTION:Last day to add a class without an override"],
      [
        "SUMMARY:F26 - Withdraw deadline",
        "DESCRIPTION:Last day to withdraw from a class. Online registration via Web4U ends at 11:59 p.m.",
      ],
    ])
  })

  it("prefixes each term's titles with its abbreviation", () => {
    const prefix = (term: string) => prop(events(build([], { holidays: true }, { ...fall, term }))[0], "SUMMARY")[0]
    expect(prefix("202710")).toBe("SUMMARY:W27 - Veterans Day")
    expect(prefix("202720")).toBe("SUMMARY:Sp27 - Veterans Day")
    expect(prefix("202730")).toBe("SUMMARY:Su27 - Veterans Day")
  })

  it("emits the time zone once when any event is timed", () => {
    const ics = build([csci247], { ...DEFAULT_CALENDAR_EXPORT })
    expect(ics.match(/BEGIN:VTIMEZONE/g)).toHaveLength(1)
    expect(events(ics)).toHaveLength(2 + 1 + 3)
  })

  it("reports whether there is anything to export", () => {
    const input = { calendar: fall, sections: [] }
    expect(hasIcsEvents({ ...input, options: { ...allOff, classes: true, finals: true } })).toBe(false)
    expect(hasIcsEvents({ ...input, options: { ...allOff, holidays: true } })).toBe(true)
  })
})

describe("ics encoding", () => {
  it("uses CRLF line endings throughout", () => {
    const ics = build([csci247], DEFAULT_CALENDAR_EXPORT)
    expect(ics.endsWith("END:VCALENDAR\r\n")).toBe(true)
    expect(ics.replace(/\r\n/g, "")).not.toMatch(/[\r\n]/)
  })

  it("escapes TEXT values", () => {
    expect(escapeText("Ethics, Law; and \\ Society\nRoom 2")).toBe("Ethics\\, Law\\; and \\\\ Society\\nRoom 2")
  })

  it("escapes every line break style, including a bare CR", () => {
    expect(escapeText("a\r\nb\rc\nd")).toBe("a\\nb\\nc\\nd")
  })

  it("folds at 75 octets without splitting a character", () => {
    const line = `SUMMARY:${"Þórður Guðjónsson ".repeat(8)}`
    const folded = foldLine(line)
    const encoder = new TextEncoder()
    for (const part of folded.split("\r\n")) {
      expect(encoder.encode(part).length).toBeLessThanOrEqual(75)
    }
    expect(folded.split("\r\n").length).toBeGreaterThan(2)
    expect(folded.replace(/\r\n /g, "")).toBe(line)
  })

  it("moves a two-octet character that would straddle octet 75", () => {
    const x = "X".repeat(73)
    expect(foldLine(`${x}ð`)).toBe(`${x}ð`)
    expect(foldLine(`${x}Xð`)).toBe(`${x}X\r\n ð`)
  })
})

describe("icsFilename", () => {
  it("slugs the term name", () => {
    expect(icsFilename("Winter 2027")).toBe("winter-2027.ics")
  })
})
