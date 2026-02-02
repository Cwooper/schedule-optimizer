import { describe, it, expect } from "vitest"
import { hydrateSection, hydrateSchedule, hydrateAsyncs } from "./schedule-utils"
import type {
  GenerateCourseInfo,
  GenerateSectionInfo,
  GenerateResponse,
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
