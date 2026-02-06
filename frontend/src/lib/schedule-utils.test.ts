import { describe, it, expect } from "vitest"
import { hydrateSection, hydrateSchedule, hydrateAsyncs, groupSectionsByCourse, clampToAvoidOverlap, blocksToRanges, otherBlockRanges, mergeAdjacentBlocks, type TimeRange } from "./schedule-utils"
import type { BlockedTimeBlock } from "@/stores/app-store"
import type {
  GenerateCourseInfo,
  GenerateSectionInfo,
  GenerateResponse,
  HydratedSection,
  ScheduleRef,
} from "./api"

const mockCourses: Record<string, GenerateCourseInfo> = {
  "CSCI:241": {
    subject: "CSCI",
    courseNumber: "241",
    title: "Data Structures",
    credits: 4,
  },
  "MATH:204": {
    subject: "MATH",
    courseNumber: "204",
    title: "Linear Algebra",
    credits: 5,
  },
}

const mockSections: Record<string, GenerateSectionInfo> = {
  "12345": {
    crn: "12345",
    term: "202510",
    courseKey: "CSCI:241",
    instructor: "Smith",
    enrollment: 25,
    maxEnrollment: 30,
    seatsAvailable: 5,
    waitCount: 0,
    isOpen: true,
    meetingTimes: [
      {
        days: [false, true, false, true, false, false, false],
        startTime: "0900",
        endTime: "0950",
        building: "CF",
        room: "225",
      },
    ],
  },
  "67890": {
    crn: "67890",
    term: "202510",
    courseKey: "MATH:204",
    instructor: "Jones",
    enrollment: 20,
    maxEnrollment: 25,
    seatsAvailable: 5,
    waitCount: 0,
    isOpen: true,
    meetingTimes: [],
  },
}

describe("hydrateSection", () => {
  it("returns hydrated section for valid CRN", () => {
    const result = hydrateSection("12345", mockCourses, mockSections)

    expect(result).not.toBeNull()
    expect(result?.crn).toBe("12345")
    expect(result?.subject).toBe("CSCI")
    expect(result?.courseNumber).toBe("241")
    expect(result?.title).toBe("Data Structures")
    expect(result?.credits).toBe(4)
    expect(result?.instructor).toBe("Smith")
    expect(result?.enrollment).toBe(25)
    expect(result?.meetingTimes).toHaveLength(1)
  })

  it("returns null for missing CRN", () => {
    const result = hydrateSection("99999", mockCourses, mockSections)
    expect(result).toBeNull()
  })

  it("returns null for missing course key", () => {
    const sectionsWithBadKey = {
      "11111": {
        ...mockSections["12345"],
        crn: "11111",
        courseKey: "INVALID:999",
      },
    }
    const result = hydrateSection("11111", mockCourses, sectionsWithBadKey)
    expect(result).toBeNull()
  })

  it("handles section with empty instructor", () => {
    const sectionsWithNoInstructor = {
      "11111": {
        ...mockSections["12345"],
        crn: "11111",
        instructor: "",
      },
    }
    const result = hydrateSection("11111", mockCourses, sectionsWithNoInstructor)
    expect(result?.instructor).toBe("")
  })
})

describe("hydrateSchedule", () => {
  it("hydrates all valid CRNs in schedule", () => {
    const schedule: ScheduleRef = {
      crns: ["12345", "67890"],
      score: 0.85,
      weights: [{ name: "gaps", value: 0.5 }],
    }

    const result = hydrateSchedule(schedule, mockCourses, mockSections)

    expect(result.courses).toHaveLength(2)
    expect(result.score).toBe(0.85)
    expect(result.weights).toEqual([{ name: "gaps", value: 0.5 }])
  })

  it("filters out invalid CRNs", () => {
    const schedule: ScheduleRef = {
      crns: ["12345", "INVALID", "67890"],
      score: 0.75,
      weights: [],
    }

    const result = hydrateSchedule(schedule, mockCourses, mockSections)

    expect(result.courses).toHaveLength(2)
    expect(result.courses.map((c) => c.crn)).toEqual(["12345", "67890"])
  })

  it("handles empty CRNs array", () => {
    const schedule: ScheduleRef = {
      crns: [],
      score: 0,
      weights: [],
    }

    const result = hydrateSchedule(schedule, mockCourses, mockSections)

    expect(result.courses).toHaveLength(0)
    expect(result.score).toBe(0)
  })
})

describe("hydrateAsyncs", () => {
  it("hydrates async CRNs from response", () => {
    const response: GenerateResponse = {
      courses: mockCourses,
      sections: mockSections,
      schedules: [],
      asyncs: ["12345"],
      courseResults: [],
      stats: { totalGenerated: 0, timeMs: 0 },
    }

    const result = hydrateAsyncs(response)

    expect(result).toHaveLength(1)
    expect(result[0].crn).toBe("12345")
  })

  it("filters out invalid async CRNs", () => {
    const response: GenerateResponse = {
      courses: mockCourses,
      sections: mockSections,
      schedules: [],
      asyncs: ["12345", "INVALID"],
      courseResults: [],
      stats: { totalGenerated: 0, timeMs: 0 },
    }

    const result = hydrateAsyncs(response)

    expect(result).toHaveLength(1)
  })

  it("returns empty array for no asyncs", () => {
    const response: GenerateResponse = {
      courses: mockCourses,
      sections: mockSections,
      schedules: [],
      asyncs: [],
      courseResults: [],
      stats: { totalGenerated: 0, timeMs: 0 },
    }

    const result = hydrateAsyncs(response)

    expect(result).toHaveLength(0)
  })
})

describe("clampToAvoidOverlap", () => {
  describe("clamp mode", () => {
    it("returns proposed unchanged when no existing blocks", () => {
      const proposed: TimeRange = { day: 0, startMin: 540, endMin: 600 }
      const result = clampToAvoidOverlap(proposed, [], "clamp")
      expect(result).toEqual(proposed)
    })

    it("returns proposed unchanged when existing blocks are on different days", () => {
      const proposed: TimeRange = { day: 0, startMin: 540, endMin: 600 }
      const existing: TimeRange[] = [{ day: 1, startMin: 540, endMin: 600 }]
      const result = clampToAvoidOverlap(proposed, existing, "clamp")
      expect(result).toEqual(proposed)
    })

    it("clamps end when painting downward into existing block", () => {
      const proposed: TimeRange = { day: 0, startMin: 480, endMin: 600 }
      const existing: TimeRange[] = [{ day: 0, startMin: 540, endMin: 660 }]
      const result = clampToAvoidOverlap(proposed, existing, "clamp")
      expect(result).toEqual({ day: 0, startMin: 480, endMin: 540 })
    })

    it("clamps start when painting upward into existing block", () => {
      const proposed: TimeRange = { day: 0, startMin: 480, endMin: 600 }
      const existing: TimeRange[] = [{ day: 0, startMin: 420, endMin: 540 }]
      const result = clampToAvoidOverlap(proposed, existing, "clamp")
      expect(result).toEqual({ day: 0, startMin: 540, endMin: 600 })
    })

    it("returns null when proposed is completely inside existing block", () => {
      const proposed: TimeRange = { day: 0, startMin: 540, endMin: 600 }
      const existing: TimeRange[] = [{ day: 0, startMin: 480, endMin: 660 }]
      const result = clampToAvoidOverlap(proposed, existing, "clamp")
      expect(result).toBeNull()
    })

    it("clamps end when dragging downward fully over a block (anchor-aware)", () => {
      // Drag from 9am to 1pm over a 10am-11am block, anchor at 9am
      const proposed: TimeRange = { day: 0, startMin: 540, endMin: 780 }
      const existing: TimeRange[] = [{ day: 0, startMin: 600, endMin: 660 }]
      const result = clampToAvoidOverlap(proposed, existing, "clamp", undefined, 540)
      expect(result).toEqual({ day: 0, startMin: 540, endMin: 600 })
    })

    it("clamps start when dragging upward fully over a block (anchor-aware)", () => {
      // Drag from 1pm up to 9am over a 10am-11am block, anchor at 1pm
      const proposed: TimeRange = { day: 0, startMin: 540, endMin: 780 }
      const existing: TimeRange[] = [{ day: 0, startMin: 600, endMin: 660 }]
      const result = clampToAvoidOverlap(proposed, existing, "clamp", undefined, 780)
      expect(result).toEqual({ day: 0, startMin: 660, endMin: 780 })
    })

    it("returns proposed unchanged when no overlap on same day", () => {
      const proposed: TimeRange = { day: 0, startMin: 480, endMin: 540 }
      const existing: TimeRange[] = [{ day: 0, startMin: 600, endMin: 660 }]
      const result = clampToAvoidOverlap(proposed, existing, "clamp")
      expect(result).toEqual(proposed)
    })

    it("clamps at the nearest block when multiple blocks overlap", () => {
      // Two blocks: 10-11am and 12-1pm. Paint from 9am to 2pm with anchor at 9am.
      const proposed: TimeRange = { day: 0, startMin: 540, endMin: 840 }
      const existing: TimeRange[] = [
        { day: 0, startMin: 600, endMin: 660 },
        { day: 0, startMin: 720, endMin: 780 },
      ]
      const result = clampToAvoidOverlap(proposed, existing, "clamp", undefined, 540)
      // Should stop at the first block boundary (10am = 600)
      expect(result).toEqual({ day: 0, startMin: 540, endMin: 600 })
    })

    it("returns null for zero-duration proposed range", () => {
      const proposed: TimeRange = { day: 0, startMin: 600, endMin: 600 }
      const existing: TimeRange[] = [{ day: 0, startMin: 540, endMin: 660 }]
      const result = clampToAvoidOverlap(proposed, existing, "clamp")
      expect(result).toBeNull()
    })

    it("handles blocks at day boundaries (midnight)", () => {
      const proposed: TimeRange = { day: 0, startMin: 0, endMin: 60 }
      const result = clampToAvoidOverlap(proposed, [], "clamp")
      expect(result).toEqual({ day: 0, startMin: 0, endMin: 60 })
    })

    it("handles blocks at end of day (1440)", () => {
      const proposed: TimeRange = { day: 0, startMin: 1380, endMin: 1440 }
      const result = clampToAvoidOverlap(proposed, [], "clamp")
      expect(result).toEqual({ day: 0, startMin: 1380, endMin: 1440 })
    })
  })

  describe("snap mode", () => {
    it("returns proposed unchanged when no overlap", () => {
      const proposed: TimeRange = { day: 0, startMin: 480, endMin: 540 }
      const existing: TimeRange[] = [{ day: 0, startMin: 600, endMin: 660 }]
      const result = clampToAvoidOverlap(proposed, existing, "snap")
      expect(result).toEqual(proposed)
    })

    it("snaps above overlapping block when closer to top", () => {
      const proposed: TimeRange = { day: 0, startMin: 520, endMin: 580 }
      const existing: TimeRange[] = [{ day: 0, startMin: 540, endMin: 660 }]
      const result = clampToAvoidOverlap(proposed, existing, "snap", 60)
      expect(result).toEqual({ day: 0, startMin: 480, endMin: 540 })
    })

    it("snaps below overlapping block when closer to bottom", () => {
      const proposed: TimeRange = { day: 0, startMin: 620, endMin: 680 }
      const existing: TimeRange[] = [{ day: 0, startMin: 540, endMin: 660 }]
      const result = clampToAvoidOverlap(proposed, existing, "snap", 60)
      expect(result).toEqual({ day: 0, startMin: 660, endMin: 720 })
    })

    it("returns null when no valid snap position exists", () => {
      // Block fills entire day, leaving no room
      const proposed: TimeRange = { day: 0, startMin: 600, endMin: 720 }
      const existing: TimeRange[] = [
        { day: 0, startMin: 0, endMin: 720 },
        { day: 0, startMin: 720, endMin: 1440 },
      ]
      const result = clampToAvoidOverlap(proposed, existing, "snap", 120)
      expect(result).toBeNull()
    })

    it("does not interact with blocks on different days", () => {
      const proposed: TimeRange = { day: 0, startMin: 540, endMin: 600 }
      const existing: TimeRange[] = [{ day: 1, startMin: 540, endMin: 600 }]
      const result = clampToAvoidOverlap(proposed, existing, "snap")
      expect(result).toEqual(proposed)
    })

    it("snaps into tight gap between two blocks", () => {
      // Two blocks with exactly 60min gap: 9-10am and 11am-12pm. Move 60min block into gap.
      const proposed: TimeRange = { day: 0, startMin: 570, endMin: 630 }
      const existing: TimeRange[] = [
        { day: 0, startMin: 540, endMin: 600 },
        { day: 0, startMin: 660, endMin: 720 },
      ]
      const result = clampToAvoidOverlap(proposed, existing, "snap", 60)
      expect(result).toEqual({ day: 0, startMin: 600, endMin: 660 })
    })

    it("snaps above when gap below is too small", () => {
      // 30min gap between blocks, 60min block can't fit in gap but can snap above
      const proposed: TimeRange = { day: 0, startMin: 590, endMin: 650 }
      const existing: TimeRange[] = [
        { day: 0, startMin: 540, endMin: 600 },
        { day: 0, startMin: 630, endMin: 720 },
      ]
      const result = clampToAvoidOverlap(proposed, existing, "snap", 60)
      expect(result).toEqual({ day: 0, startMin: 480, endMin: 540 })
    })
  })
})

// ── Block utility tests ──────────────────────────────────────────────

function makeBlock(id: string, day: number, start: string, end: string): BlockedTimeBlock {
  return { id, day, startTime: start, endTime: end }
}

describe("blocksToRanges", () => {
  it("converts blocks to time ranges", () => {
    const blocks = [
      makeBlock("a", 0, "0900", "1000"),
      makeBlock("b", 2, "1400", "1530"),
    ]
    expect(blocksToRanges(blocks)).toEqual([
      { day: 0, startMin: 540, endMin: 600 },
      { day: 2, startMin: 840, endMin: 930 },
    ])
  })

  it("returns empty array for empty input", () => {
    expect(blocksToRanges([])).toEqual([])
  })
})

describe("otherBlockRanges", () => {
  it("excludes block with matching id", () => {
    const blocks = [
      makeBlock("a", 0, "0900", "1000"),
      makeBlock("b", 0, "1100", "1200"),
    ]
    const result = otherBlockRanges(blocks, "a")
    expect(result).toEqual([{ day: 0, startMin: 660, endMin: 720 }])
  })

  it("returns all ranges when excludeId matches nothing", () => {
    const blocks = [makeBlock("a", 0, "0900", "1000")]
    const result = otherBlockRanges(blocks, "z")
    expect(result).toHaveLength(1)
  })
})

describe("mergeAdjacentBlocks", () => {
  it("returns empty array unchanged", () => {
    expect(mergeAdjacentBlocks([])).toEqual([])
  })

  it("returns single block unchanged", () => {
    const blocks = [makeBlock("a", 0, "0900", "1000")]
    expect(mergeAdjacentBlocks(blocks)).toEqual(blocks)
  })

  it("merges two adjacent blocks on same day", () => {
    const blocks = [
      makeBlock("a", 0, "0900", "1000"),
      makeBlock("b", 0, "1000", "1100"),
    ]
    const result = mergeAdjacentBlocks(blocks)
    expect(result).toHaveLength(1)
    expect(result[0]).toEqual(makeBlock("a", 0, "0900", "1100"))
  })

  it("keeps earlier block's ID after merge", () => {
    const blocks = [
      makeBlock("b", 0, "1000", "1100"),
      makeBlock("a", 0, "0900", "1000"),
    ]
    const result = mergeAdjacentBlocks(blocks)
    expect(result).toHaveLength(1)
    expect(result[0].id).toBe("a")
  })

  it("does not merge blocks on different days", () => {
    const blocks = [
      makeBlock("a", 0, "0900", "1000"),
      makeBlock("b", 1, "1000", "1100"),
    ]
    const result = mergeAdjacentBlocks(blocks)
    expect(result).toHaveLength(2)
  })

  it("does not merge non-adjacent blocks on same day", () => {
    const blocks = [
      makeBlock("a", 0, "0900", "1000"),
      makeBlock("b", 0, "1100", "1200"),
    ]
    const result = mergeAdjacentBlocks(blocks)
    expect(result).toHaveLength(2)
  })

  it("chain-merges three adjacent blocks into one", () => {
    const blocks = [
      makeBlock("c", 0, "1100", "1200"),
      makeBlock("a", 0, "0900", "1000"),
      makeBlock("b", 0, "1000", "1100"),
    ]
    const result = mergeAdjacentBlocks(blocks)
    expect(result).toHaveLength(1)
    expect(result[0]).toEqual(makeBlock("a", 0, "0900", "1200"))
  })

  it("merges adjacent pair but keeps separate non-adjacent block", () => {
    const blocks = [
      makeBlock("a", 0, "0900", "1000"),
      makeBlock("b", 0, "1000", "1100"),
      makeBlock("c", 0, "1300", "1400"),
    ]
    const result = mergeAdjacentBlocks(blocks)
    expect(result).toHaveLength(2)
    expect(result[0]).toEqual(makeBlock("a", 0, "0900", "1100"))
    expect(result[1]).toEqual(makeBlock("c", 0, "1300", "1400"))
  })

  it("merges across multiple days independently", () => {
    const blocks = [
      makeBlock("a", 0, "0900", "1000"),
      makeBlock("b", 0, "1000", "1100"),
      makeBlock("c", 1, "0900", "1000"),
      makeBlock("d", 1, "1000", "1100"),
    ]
    const result = mergeAdjacentBlocks(blocks)
    expect(result).toHaveLength(2)
    expect(result[0]).toEqual(makeBlock("a", 0, "0900", "1100"))
    expect(result[1]).toEqual(makeBlock("c", 1, "0900", "1100"))
  })
})

// ── groupSectionsByCourse tests ──────────────────────────────────────

function makeHydratedSection(overrides: Partial<HydratedSection> = {}): HydratedSection {
  return {
    crn: "11111",
    term: "202510",
    subject: "CSCI",
    courseNumber: "101",
    title: "Intro to CS",
    credits: 4,
    instructor: "Smith",
    meetingTimes: [],
    enrollment: 20,
    maxEnrollment: 30,
    seatsAvailable: 10,
    waitCount: 0,
    isOpen: true,
    ...overrides,
  }
}

describe("groupSectionsByCourse", () => {
  it("returns empty array for no sections", () => {
    expect(groupSectionsByCourse([])).toEqual([])
  })

  it("returns single course item for single section", () => {
    const result = groupSectionsByCourse([makeHydratedSection()])

    expect(result).toHaveLength(1)
    expect(result[0].subject).toBe("CSCI")
    expect(result[0].courseNumber).toBe("101")
    expect(result[0].title).toBe("Intro to CS")
    expect(result[0].credits).toBe(4)
    expect(result[0].sections).toHaveLength(1)
    expect(result[0].sections[0].crn).toBe("11111")
    expect(result[0].firstCrn).toBe("11111")
  })

  it("groups multiple sections of the same course", () => {
    const sections = [
      makeHydratedSection({ crn: "11111", instructor: "Smith" }),
      makeHydratedSection({ crn: "22222", instructor: "Jones" }),
    ]

    const result = groupSectionsByCourse(sections)

    expect(result).toHaveLength(1)
    expect(result[0].sections).toHaveLength(2)
    expect(result[0].sections[0].crn).toBe("11111")
    expect(result[0].sections[1].crn).toBe("22222")
    expect(result[0].firstCrn).toBe("11111")
  })

  it("separates different courses", () => {
    const sections = [
      makeHydratedSection({ crn: "11111", subject: "CSCI", courseNumber: "101" }),
      makeHydratedSection({ crn: "22222", subject: "MATH", courseNumber: "204", title: "Linear Algebra", credits: 5 }),
    ]

    const result = groupSectionsByCourse(sections)

    expect(result).toHaveLength(2)
    expect(result[0].subject).toBe("CSCI")
    expect(result[1].subject).toBe("MATH")
    expect(result[1].title).toBe("Linear Algebra")
  })

  it("converts empty instructor to undefined", () => {
    const result = groupSectionsByCourse([makeHydratedSection({ instructor: "" })])

    expect(result[0].sections[0].instructor).toBeUndefined()
  })

  it("preserves section enrollment info", () => {
    const result = groupSectionsByCourse([
      makeHydratedSection({ seatsAvailable: 5, isOpen: false }),
    ])

    expect(result[0].sections[0].seatsAvailable).toBe(5)
    expect(result[0].sections[0].isOpen).toBe(false)
  })
})
