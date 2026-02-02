import { useAppStore, type CourseSlot } from "./app-store"
import type { GenerateResponse } from "@/lib/api"

const mockSlot: CourseSlot = {
  id: "test-1",
  subject: "CSCI",
  courseNumber: "247",
  displayName: "CSCI 247",
  required: true,
  sections: null,
}

const mockGenerateResult: GenerateResponse = {
  courses: {},
  sections: {},
  schedules: [
    {
      crns: [],
      score: 0.85,
      weights: [],
    },
  ],
  asyncs: [],
  courseResults: [],
  stats: {
    totalGenerated: 1,
    timeMs: 50,
  },
}

describe("app-store", () => {
  beforeEach(() => {
    // Reset store to initial state before each test
    useAppStore.setState({
      tab: "schedule",
      term: "",
      selectedSubject: "",
      minCourses: null,
      maxCourses: null,
      slots: [],
      theme: "system",
      sidebarCollapsed: false,
      generateResult: null,
      isGenerateResultStale: false,
      currentScheduleIndex: 0,
      slotsVersion: 0,
    })
  })

  describe("initial state", () => {
    it("has correct defaults", () => {
      const state = useAppStore.getState()
      expect(state.tab).toBe("schedule")
      expect(state.term).toBe("")
      expect(state.slots).toEqual([])
      expect(state.generateResult).toBeNull()
      expect(state.isGenerateResultStale).toBe(false)
      expect(state.theme).toBe("system")
    })
  })

  describe("slot management", () => {
    it("adds a slot", () => {
      useAppStore.getState().addSlot(mockSlot)
      expect(useAppStore.getState().slots).toHaveLength(1)
      expect(useAppStore.getState().slots[0]).toEqual(mockSlot)
    })

    it("removes a slot by id", () => {
      useAppStore.getState().addSlot(mockSlot)
      useAppStore.getState().addSlot({ ...mockSlot, id: "test-2" })
      expect(useAppStore.getState().slots).toHaveLength(2)

      useAppStore.getState().removeSlot("test-1")
      expect(useAppStore.getState().slots).toHaveLength(1)
      expect(useAppStore.getState().slots[0].id).toBe("test-2")
    })

    it("updates a slot", () => {
      useAppStore.getState().addSlot(mockSlot)
      useAppStore.getState().updateSlot("test-1", { required: false })

      expect(useAppStore.getState().slots[0].required).toBe(false)
      expect(useAppStore.getState().slots[0].subject).toBe("CSCI") // unchanged
    })

    it("clears all slots", () => {
      useAppStore.getState().addSlot(mockSlot)
      useAppStore.getState().addSlot({ ...mockSlot, id: "test-2" })
      useAppStore.getState().clearSlots()

      expect(useAppStore.getState().slots).toEqual([])
    })
  })

  describe("schedule staleness", () => {
    beforeEach(() => {
      // Set up generateResult before each test in this group
      useAppStore.getState().setGenerateResult(mockGenerateResult)
      useAppStore.setState({ currentScheduleIndex: 5 })
      expect(useAppStore.getState().generateResult).not.toBeNull()
      expect(useAppStore.getState().isGenerateResultStale).toBe(false)
    })

    it("marks stale when adding a slot", () => {
      useAppStore.getState().addSlot(mockSlot)

      expect(useAppStore.getState().generateResult).not.toBeNull()
      expect(useAppStore.getState().isGenerateResultStale).toBe(true)
    })

    it("marks stale when removing a slot", () => {
      useAppStore.getState().addSlot(mockSlot)
      useAppStore.getState().setGenerateResult(mockGenerateResult)
      useAppStore.getState().removeSlot("test-1")

      expect(useAppStore.getState().generateResult).not.toBeNull()
      expect(useAppStore.getState().isGenerateResultStale).toBe(true)
    })

    it("marks stale when updating a slot", () => {
      useAppStore.getState().addSlot(mockSlot)
      useAppStore.getState().setGenerateResult(mockGenerateResult)
      useAppStore.getState().updateSlot("test-1", { required: false })

      expect(useAppStore.getState().generateResult).not.toBeNull()
      expect(useAppStore.getState().isGenerateResultStale).toBe(true)
    })

    it("marks stale when changing term", () => {
      useAppStore.getState().setTerm("202520")

      expect(useAppStore.getState().generateResult).not.toBeNull()
      expect(useAppStore.getState().isGenerateResultStale).toBe(true)
      expect(useAppStore.getState().term).toBe("202520")
    })

    it("clears generateResult when clearing slots", () => {
      useAppStore.getState().addSlot(mockSlot)
      useAppStore.getState().setGenerateResult(mockGenerateResult)
      useAppStore.getState().clearSlots()

      expect(useAppStore.getState().generateResult).toBeNull()
      expect(useAppStore.getState().isGenerateResultStale).toBe(false)
    })

    it("clears stale flag when setting new result", () => {
      useAppStore.getState().addSlot(mockSlot)
      expect(useAppStore.getState().isGenerateResultStale).toBe(true)

      useAppStore.getState().setGenerateResult(mockGenerateResult)
      expect(useAppStore.getState().isGenerateResultStale).toBe(false)
    })

    it("does NOT mark stale when changing tab", () => {
      useAppStore.getState().setTab("search")

      expect(useAppStore.getState().generateResult).not.toBeNull()
      expect(useAppStore.getState().isGenerateResultStale).toBe(false)
      expect(useAppStore.getState().tab).toBe("search")
    })

    it("does NOT mark stale when changing theme", () => {
      useAppStore.getState().setTheme("dark")

      expect(useAppStore.getState().generateResult).not.toBeNull()
      expect(useAppStore.getState().isGenerateResultStale).toBe(false)
    })
  })

  describe("schedule navigation", () => {
    it("setGenerateResult resets index to 0", () => {
      useAppStore.setState({ currentScheduleIndex: 5 })
      useAppStore.getState().setGenerateResult(mockGenerateResult)

      expect(useAppStore.getState().currentScheduleIndex).toBe(0)
    })

    it("setCurrentScheduleIndex updates index", () => {
      const twoSchedules: GenerateResponse = {
        ...mockGenerateResult,
        schedules: [mockGenerateResult.schedules[0], mockGenerateResult.schedules[0]],
      }
      useAppStore.getState().setGenerateResult(twoSchedules)
      useAppStore.getState().setCurrentScheduleIndex(1)

      expect(useAppStore.getState().currentScheduleIndex).toBe(1)
    })

    it("clamps index to valid range", () => {
      useAppStore.getState().setGenerateResult(mockGenerateResult)
      useAppStore.getState().setCurrentScheduleIndex(99)

      // Store now enforces bounds
      expect(useAppStore.getState().currentScheduleIndex).toBe(0)
    })
  })

  describe("course bounds", () => {
    it("sets min and max courses", () => {
      useAppStore.getState().setCourseBounds(3, 5)

      expect(useAppStore.getState().minCourses).toBe(3)
      expect(useAppStore.getState().maxCourses).toBe(5)
    })

    it("allows null values", () => {
      useAppStore.getState().setCourseBounds(3, 5)
      useAppStore.getState().setCourseBounds(null, null)

      expect(useAppStore.getState().minCourses).toBeNull()
      expect(useAppStore.getState().maxCourses).toBeNull()
    })
  })

  describe("UI state", () => {
    it("sets sidebar collapsed state", () => {
      expect(useAppStore.getState().sidebarCollapsed).toBe(false)

      useAppStore.getState().setSidebarCollapsed(true)
      expect(useAppStore.getState().sidebarCollapsed).toBe(true)

      useAppStore.getState().setSidebarCollapsed(false)
      expect(useAppStore.getState().sidebarCollapsed).toBe(false)
    })
  })

  describe("edge cases", () => {
    it("removing non-existent slot does nothing", () => {
      useAppStore.getState().addSlot(mockSlot)
      const slotsBefore = useAppStore.getState().slots

      useAppStore.getState().removeSlot("non-existent-id")

      expect(useAppStore.getState().slots).toEqual(slotsBefore)
    })

    it("updating non-existent slot does nothing", () => {
      useAppStore.getState().addSlot(mockSlot)

      useAppStore.getState().updateSlot("non-existent-id", { required: false })

      // Original slot unchanged
      expect(useAppStore.getState().slots[0].required).toBe(true)
    })

    it("preserves slot order after operations", () => {
      useAppStore.getState().addSlot({ ...mockSlot, id: "a", subject: "CSCI" })
      useAppStore.getState().addSlot({ ...mockSlot, id: "b", subject: "MATH" })
      useAppStore.getState().addSlot({ ...mockSlot, id: "c", subject: "PHYS" })

      useAppStore.getState().updateSlot("b", { required: false })

      const subjects = useAppStore.getState().slots.map((s) => s.subject)
      expect(subjects).toEqual(["CSCI", "MATH", "PHYS"])
    })
  })

  describe("slotsVersion", () => {
    it("increments when slots change", () => {
      const v0 = useAppStore.getState().getSlotsVersion()

      useAppStore.getState().addSlot(mockSlot)
      const v1 = useAppStore.getState().getSlotsVersion()
      expect(v1).toBe(v0 + 1)

      useAppStore.getState().updateSlot("test-1", { required: false })
      const v2 = useAppStore.getState().getSlotsVersion()
      expect(v2).toBe(v1 + 1)

      useAppStore.getState().removeSlot("test-1")
      const v3 = useAppStore.getState().getSlotsVersion()
      expect(v3).toBe(v2 + 1)
    })

    it("increments when term changes", () => {
      const v0 = useAppStore.getState().getSlotsVersion()

      useAppStore.getState().setTerm("202520")
      const v1 = useAppStore.getState().getSlotsVersion()
      expect(v1).toBe(v0 + 1)
    })
  })
})
