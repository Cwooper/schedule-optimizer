import { describe, it, expect } from "vitest"
import type { Term, TermCalendar } from "./api"
import { keyDateRows, lastDayOfClasses, termStatus } from "./term-status"

// Fall 2026 and Winter 2027 as the registrar published them.
const fall: TermCalendar = {
  term: "202640",
  start: "2026-09-23",
  end: "2026-12-11",
  registrationOpens: "2026-05-12",
  estimated: false,
  registration: [
    { phase: "I", start: "2026-05-12", end: "2026-07-13" },
    { phase: "II", start: "2026-08-25", end: "2026-09-22" },
    { phase: "III", start: "2026-09-23", end: "2026-09-29" },
  ],
  holidays: [
    { name: "Veterans Day", start: "2026-11-11", end: "2026-11-11" },
    { name: "Thanksgiving Break (Wed, Nov 25 - University open, no classes)", start: "2026-11-25", end: "2026-11-29" },
    { name: "Native American Heritage Day", start: "2026-11-27", end: "2026-11-27" },
  ],
  deadlines: [],
  finals: { start: "2026-12-07", end: "2026-12-11" },
}

const winter: TermCalendar = {
  term: "202710",
  start: "2027-01-05",
  end: "2027-03-19",
  registrationOpens: "2026-11-03",
  estimated: false,
  registration: [
    { phase: "I", start: "2026-11-03", end: "2026-11-16" },
    { phase: "II", start: "2026-11-17", end: "2027-01-04" },
    { phase: "III", start: "2027-01-05", end: "2027-01-11" },
  ],
  holidays: [
    { name: "Martin Luther King Jr. Day", start: "2027-01-18", end: "2027-01-18" },
    { name: "Presidents' Day", start: "2027-02-15", end: "2027-02-15" },
  ],
  deadlines: [],
  finals: { start: "2027-03-15", end: "2027-03-19" },
}

function term(code: string, name: string, phase: Term["phase"], registrationOpens: string): Term {
  return { code, name, phase, start: "", end: "", registrationOpens, estimated: false }
}

const fallTerm = term("202640", "Fall 2026", "in-session", "2026-05-12")
const winterTerm = term("202710", "Winter 2027", "pre-registration", "2026-11-03")
const springTerm = term("202720", "Spring 2027", "pre-registration", "2027-02-23")

describe("termStatus", () => {
  it("counts down to Phase I", () => {
    expect(termStatus(winterTerm, winter, [], "2026-10-20")?.text).toBe("Phase I opens Nov 3 - in 14 days")
    expect(termStatus(winterTerm, winter, [], "2026-11-02")?.text).toBe("Phase I opens Nov 3 - tomorrow")
  })

  it("says today, never a negative count, once the opening day arrives", () => {
    // The cached phase still says pre-registration until /api/terms refetches.
    expect(termStatus(winterTerm, winter, [], "2026-11-03")?.text).toBe("Phase I opens Nov 3 - today")
    expect(termStatus(winterTerm, winter, [], "2026-11-04")?.text).toBe("Phase I opens Nov 3 - today")
  })

  it("marks an estimated opening without a Phase I row", () => {
    const estimated = { ...winter, registration: [], registrationOpens: "2026-10-27", estimated: true }
    expect(termStatus(winterTerm, estimated, [], "2026-10-20")?.text).toBe("Registration opens ~Oct 27 - in 7 days")
    const published = { ...estimated, estimated: false }
    expect(termStatus(winterTerm, published, [], "2026-10-20")?.text).toBe("Registration opens Oct 27 - in 7 days")
  })

  it("names the next phase during registration, then the last phase's end", () => {
    const registering = { ...winterTerm, phase: "registration" as const }
    expect(termStatus(registering, winter, [], "2026-11-03")?.text).toBe("Phase II opens Nov 17")
    expect(termStatus(registering, winter, [], "2026-12-01")?.text).toBe("Phase III opens Jan 5")
    expect(termStatus(registering, winter, [], "2027-01-06")?.text).toBe("Phase III open until Jan 11")
  })

  it("names the coming phase in a gap between phases", () => {
    const registering = { ...fallTerm, phase: "registration" as const }
    expect(termStatus(registering, fall, [], "2026-08-01")?.text).toBe("Phase II opens Aug 25")
  })

  it("shows finals week in session", () => {
    expect(termStatus(fallTerm, fall, [fallTerm], "2026-10-15")?.text).toBe("Finals Dec 7-11")
  })

  it("shows the last day for a term without finals", () => {
    const summer = { ...fall, finals: null }
    expect(termStatus(fallTerm, summer, [], "2026-10-15")?.text).toBe("Classes end Dec 11")
  })

  it("points to a newer term whose registration opens within two weeks", () => {
    const terms = [springTerm, winterTerm, fallTerm]
    expect(termStatus(fallTerm, fall, terms, "2026-10-20")).toEqual({
      text: "Winter 2027 registration opens Nov 3",
      switchTo: "202710",
    })
    expect(termStatus(fallTerm, fall, terms, "2026-10-19")?.text).toBe("Finals Dec 7-11")
  })

  it("marks a newer term's estimated opening", () => {
    const estimated = { ...winterTerm, registrationOpens: "2026-10-27", estimated: true }
    expect(termStatus(fallTerm, fall, [estimated, fallTerm], "2026-10-20")?.text).toBe(
      "Winter 2027 registration opens ~Oct 27"
    )
  })

  it("points to a newer term registering now", () => {
    const terms = [{ ...winterTerm, phase: "registration" as const }, fallTerm]
    expect(termStatus(fallTerm, fall, terms, "2026-11-05")).toEqual({
      text: "Winter 2027 registration open",
      switchTo: "202710",
    })
  })

  it("says nothing for a past term or before its calendar loads", () => {
    expect(termStatus({ ...fallTerm, phase: "past" }, fall, [], "2027-02-01")).toBeNull()
    expect(termStatus(fallTerm, undefined, [], "2026-10-15")).toBeNull()
  })
})

describe("keyDateRows", () => {
  it("lists registration, classes, holidays, and finals with today's state", () => {
    expect(keyDateRows(winter, "2026-11-20")).toEqual([
      { label: "Phase I", dates: "Tue, Nov 3 - Mon, Nov 16", state: "past" },
      { label: "Phase II", dates: "Tue, Nov 17 - Mon, Jan 4", state: "current" },
      { label: "Phase III", dates: "Tue, Jan 5 - Mon, Jan 11", state: "upcoming" },
      { label: "Classes", dates: "Tue, Jan 5 - Fri, Mar 12", state: "upcoming" },
      { label: "No classes", dates: "Mon, Jan 18", note: "MLK Jr. Day", state: "upcoming" },
      { label: "", dates: "Mon, Feb 15", note: "Presidents' Day", state: "upcoming" },
      { label: "Finals", dates: "Mon, Mar 15 - Fri, Mar 19", state: "upcoming" },
    ])
  })

  it("lists one of two holidays with the same dates", () => {
    const twice = {
      ...winter,
      holidays: [
        { name: "Presidents' Day", start: "2027-02-15", end: "2027-02-15" },
        { name: "Washington's Birthday", start: "2027-02-15", end: "2027-02-15" },
      ],
    }
    expect(keyDateRows(twice, "2026-11-20").filter((r) => r.note).map((r) => r.note)).toEqual(["Presidents' Day"])
  })

  it("drops a holiday inside another and the trailing note", () => {
    const holidays = keyDateRows(fall, "2026-10-01").filter((r) => r.note)
    expect(holidays.map((r) => [r.dates, r.note])).toEqual([
      ["Wed, Nov 11", "Veterans Day"],
      ["Wed, Nov 25 - Sun, Nov 29", "Thanksgiving Break"],
    ])
  })
})

describe("lastDayOfClasses", () => {
  it("is the weekday before finals week", () => {
    expect(lastDayOfClasses(fall)).toBe("2026-12-04")
  })

  it("is the term's last day without finals", () => {
    expect(lastDayOfClasses({ ...fall, finals: null })).toBe("2026-12-11")
  })
})
