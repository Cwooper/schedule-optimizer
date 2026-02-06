package generator

import (
	"testing"

	"schedule-optimizer/internal/cache"
)

func TestParseTimeToMins(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"0000", 0},
		{"0100", 60},
		{"0830", 510},
		{"1030", 630},
		{"10:30", 630},
		{"1200", 720},
		{"1700", 1020},
		{"2359", 1439},
		{"", -1},
		{"invalid", -1},
		{"25:00", -1},
		{"12:60", -1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseTimeToMins(tt.input)
			if got != tt.want {
				t.Errorf("parseTimeToMins(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestTimeToSlot(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"0700", 0},  // 7:00 AM = slot 0
		{"0710", 1},  // 7:10 AM = slot 1
		{"0800", 6},  // 8:00 AM = slot 6
		{"0830", 9},  // 8:30 AM = slot 9
		{"1000", 18}, // 10:00 AM = slot 18
		{"1200", 30}, // 12:00 PM = slot 30
		{"1700", 60}, // 5:00 PM = slot 60
		{"2150", 89}, // 9:50 PM = slot 89 (last valid slot)
		{"0600", -1}, // Before 7 AM
		{"2200", -1}, // After 10 PM
		{"", -1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := timeToSlot(tt.input)
			if got != tt.want {
				t.Errorf("timeToSlot(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestTimeMask_SetSlotAndConflicts(t *testing.T) {
	var m1 TimeMask
	m1.SetSlot(0, 10) // Monday slot 10

	var m2 TimeMask
	m2.SetSlot(0, 10) // Monday slot 10 (same)

	var m3 TimeMask
	m3.SetSlot(0, 11) // Monday slot 11 (different)

	var m4 TimeMask
	m4.SetSlot(1, 10) // Tuesday slot 10 (different day)

	if !m1.Conflicts(m2) {
		t.Error("m1 should conflict with m2 (same day/slot)")
	}

	if m1.Conflicts(m3) {
		t.Error("m1 should not conflict with m3 (different slot)")
	}

	if m1.Conflicts(m4) {
		t.Error("m1 should not conflict with m4 (different day)")
	}
}

func TestTimeMask_Merge(t *testing.T) {
	var m1 TimeMask
	m1.SetSlot(0, 10)
	m1.SetSlot(0, 11)

	var m2 TimeMask
	m2.SetSlot(0, 15)
	m2.SetSlot(1, 10)

	merged := m1.Merge(m2)

	// Check that merged has all slots from both
	var expected TimeMask
	expected.SetSlot(0, 10)
	expected.SetSlot(0, 11)
	expected.SetSlot(0, 15)
	expected.SetSlot(1, 10)

	for i := range 8 {
		if merged[i] != expected[i] {
			t.Errorf("Merge result differs at index %d: got %v, want %v", i, merged[i], expected[i])
		}
	}
}

func TestFromMeetingTimes(t *testing.T) {
	meetings := []cache.MeetingTime{
		{
			Days:      [7]bool{false, true, false, true, false, true, false}, // Mon, Wed, Fri
			StartTime: "0900",
			EndTime:   "0950",
		},
	}

	mask := FromMeetingTimes(meetings)

	// 9:00 AM = slot 12, 9:50 AM = slot 17 (so slots 12-16 should be set)
	// Check Monday (day 0)
	var expected TimeMask
	for slot := 12; slot < 17; slot++ {
		expected.SetSlot(0, slot) // Monday
		expected.SetSlot(2, slot) // Wednesday
		expected.SetSlot(4, slot) // Friday
	}

	for i := range 8 {
		if mask[i] != expected[i] {
			t.Errorf("FromMeetingTimes differs at index %d: got %v, want %v", i, mask[i], expected[i])
		}
	}
}

func TestFromBlockedTimes(t *testing.T) {
	blocked := []BlockedTime{
		{Day: 0, StartTime: "0800", EndTime: "0900"}, // Monday 8-9am
		{Day: 2, StartTime: "1200", EndTime: "1300"}, // Wednesday 12-1pm
	}

	mask := FromBlockedTimes(blocked)

	// Monday 8am (slot 6) to 9am (slot 12) -> slots 6-11
	var expected TimeMask
	for slot := 6; slot < 12; slot++ {
		expected.SetSlot(0, slot)
	}
	// Wednesday 12pm (slot 30) to 1pm (slot 36) -> slots 30-35
	for slot := 30; slot < 36; slot++ {
		expected.SetSlot(2, slot)
	}

	for i := range 8 {
		if mask[i] != expected[i] {
			t.Errorf("FromBlockedTimes differs at index %d: got %v, want %v", i, mask[i], expected[i])
		}
	}
}

func TestFromMeetingTimes_EmptyOrTBD(t *testing.T) {
	// Empty meetings
	mask := FromMeetingTimes(nil)
	if mask != (TimeMask{}) {
		t.Error("Empty meetings should return empty mask")
	}

	// TBD times (empty strings)
	meetings := []cache.MeetingTime{
		{
			Days:      [7]bool{false, true, false, false, false, false, false},
			StartTime: "",
			EndTime:   "",
		},
	}
	mask = FromMeetingTimes(meetings)
	if mask != (TimeMask{}) {
		t.Error("TBD meetings should return empty mask")
	}
}
