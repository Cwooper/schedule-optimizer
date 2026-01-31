import { useAppStore, type CourseSlot, type StoredSchedule } from "./app-store"

const mockSlot: CourseSlot = {
  id: "test-1",
  subject: "CSCI",
  courseNumber: "247",
  displayName: "CSCI 247",
  required: true,
  sections: null,
}

const mockSchedule: StoredSchedule = {
  sections: ["12345", "67890"],
  totalCredits: 10,
  score: 0.85,
}

describe("app-store", () => {
  beforeEach(() => {
    // Reset store to initial state before each test
    useAppStore.setState({
      tab: "schedule",
      term: "",
      minCredits: null,
      maxCredits: null,
      slots: [],
      theme: "system",
      sidebarCollapsed: false,
      schedules: null,
      currentScheduleIndex: 0,
    })
  })

  describe("initial state", () => {
    it("has correct defaults", () => {
      const state = useAppStore.getState()
      expect(state.tab).toBe("schedule")
      expect(state.term).toBe("")
      expect(state.slots).toEqual([])
      expect(state.schedules).toBeNull()
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

  describe("schedule clearing", () => {
    beforeEach(() => {
      // Set up schedules before each test in this group
      useAppStore.getState().setSchedules([mockSchedule])
      useAppStore.getState().setCurrentScheduleIndex(5)
      expect(useAppStore.getState().schedules).not.toBeNull()
    })

    it("clears schedules when adding a slot", () => {
      useAppStore.getState().addSlot(mockSlot)

      expect(useAppStore.getState().schedules).toBeNull()
      expect(useAppStore.getState().currentScheduleIndex).toBe(0)
    })

    it("clears schedules when removing a slot", () => {
      useAppStore.getState().addSlot(mockSlot)
      useAppStore.getState().setSchedules([mockSchedule]) // re-add schedules
      useAppStore.getState().removeSlot("test-1")

      expect(useAppStore.getState().schedules).toBeNull()
      expect(useAppStore.getState().currentScheduleIndex).toBe(0)
    })

    it("clears schedules when updating a slot", () => {
      useAppStore.getState().addSlot(mockSlot)
      useAppStore.getState().setSchedules([mockSchedule])
      useAppStore.getState().updateSlot("test-1", { required: false })

      expect(useAppStore.getState().schedules).toBeNull()
      expect(useAppStore.getState().currentScheduleIndex).toBe(0)
    })

    it("clears schedules when changing term", () => {
      useAppStore.getState().setTerm("202520")

      expect(useAppStore.getState().schedules).toBeNull()
      expect(useAppStore.getState().currentScheduleIndex).toBe(0)
      expect(useAppStore.getState().term).toBe("202520")
    })

    it("clears schedules when clearing slots", () => {
      useAppStore.getState().addSlot(mockSlot)
      useAppStore.getState().setSchedules([mockSchedule])
      useAppStore.getState().clearSlots()

      expect(useAppStore.getState().schedules).toBeNull()
    })

    it("does NOT clear schedules when changing tab", () => {
      useAppStore.getState().setTab("search")

      expect(useAppStore.getState().schedules).not.toBeNull()
      expect(useAppStore.getState().tab).toBe("search")
    })

    it("does NOT clear schedules when changing theme", () => {
      useAppStore.getState().setTheme("dark")

      expect(useAppStore.getState().schedules).not.toBeNull()
    })
  })

  describe("schedule navigation", () => {
    it("setSchedules resets index to 0", () => {
      useAppStore.getState().setCurrentScheduleIndex(5)
      useAppStore.getState().setSchedules([mockSchedule])

      expect(useAppStore.getState().currentScheduleIndex).toBe(0)
    })

    it("setCurrentScheduleIndex updates index", () => {
      useAppStore.getState().setSchedules([mockSchedule, mockSchedule])
      useAppStore.getState().setCurrentScheduleIndex(1)

      expect(useAppStore.getState().currentScheduleIndex).toBe(1)
    })

    it("allows setting index beyond bounds (consumer responsibility)", () => {
      useAppStore.getState().setSchedules([mockSchedule])
      useAppStore.getState().setCurrentScheduleIndex(99)

      // Store doesn't enforce bounds - consumers should handle this
      expect(useAppStore.getState().currentScheduleIndex).toBe(99)
    })
  })

  describe("credits", () => {
    it("sets min and max credits", () => {
      useAppStore.getState().setCredits(12, 18)

      expect(useAppStore.getState().minCredits).toBe(12)
      expect(useAppStore.getState().maxCredits).toBe(18)
    })

    it("allows null values", () => {
      useAppStore.getState().setCredits(12, 18)
      useAppStore.getState().setCredits(null, null)

      expect(useAppStore.getState().minCredits).toBeNull()
      expect(useAppStore.getState().maxCredits).toBeNull()
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
})
