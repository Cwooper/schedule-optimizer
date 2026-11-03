import { describe, it, expect } from "vitest"
import { decodeRepeated, termAbbreviation } from "./utils"

// One layer per call, like the browser decoder: the regex never rescans its own output.
const ENTITIES: Record<string, string> = { "&amp;": "&", "&#39;": "'", "&nbsp;": " " }
const decodeOnce = (s: string) => s.replace(/&(amp|#39|nbsp);/g, (m) => ENTITIES[m])

describe("decodeRepeated", () => {
  it.each([
    ["plain text", "Data Structures", "Data Structures"],
    ["single escape", "B&amp;W Photography", "B&W Photography"],
    ["double escape", "Children&amp;#39;s Literature", "Children's Literature"],
    ["triple escape", "R&amp;amp;amp;D", "R&D"],
    ["nbsp becomes space", "Theories&nbsp;and Techniques", "Theories and Techniques"],
  ])("%s", (_, input, want) => {
    expect(decodeRepeated(input, decodeOnce)).toBe(want)
  })
})

describe("termAbbreviation", () => {
  it.each([
    ["202640", "F26"],
    ["202710", "W27"],
    ["202720", "Sp27"],
    ["202730", "Su27"],
    ["209950", "209950"],
  ])("%s -> %s", (code, want) => {
    expect(termAbbreviation(code)).toBe(want)
  })
})
