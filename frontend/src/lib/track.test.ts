import { describe, it, expect } from "vitest"
import { DEFAULT_CALENDAR_EXPORT, type CalendarExportOptions } from "./ics"
import { icsExportTarget } from "./track"

const NONE: CalendarExportOptions = {
  classes: false,
  finals: false,
  holidays: false,
  registration: { I: false, II: false, III: false },
  deadlines: false,
}

describe("icsExportTarget", () => {
  it.each<[string, CalendarExportOptions, string]>([
    ["defaults", DEFAULT_CALENDAR_EXPORT, "classes,finals,holidays"],
    [
      "every option, in the backend's order",
      { classes: true, finals: true, holidays: true, registration: { I: true, II: true, III: true }, deadlines: true },
      "classes,finals,holidays,reg1,reg2,reg3,deadlines",
    ],
    ["some phases", { ...NONE, registration: { I: true, II: false, III: true } }, "reg1,reg3"],
    ["deadlines only", { ...NONE, deadlines: true }, "deadlines"],
    ["nothing", NONE, ""],
  ])("%s", (_, options, want) => {
    expect(icsExportTarget(options)).toBe(want)
  })
})
