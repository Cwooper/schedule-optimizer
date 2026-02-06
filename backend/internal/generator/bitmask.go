package generator

import (
	"strconv"
	"strings"

	"schedule-optimizer/internal/cache"
)

// Time slot constants for bitmask representation.
// Using 10-minute granularity: classes can start at :00, :10, :20, :30, :40, :50.
const (
	slotsPerDay  = 90  // 7am to 10pm in 10-min increments (15 hours * 6 slots)
	dayStartMins = 420 // 7:00 AM in minutes from midnight
)

// TimeMask represents a week's schedule as a bitmask.
// 5 days × 90 ten-minute slots (7am-10pm) = 450 bits.
// Stored in 8 uint64s (512 bits) for simple indexing.
type TimeMask [8]uint64

// EmptyMask returns a TimeMask with no time slots set.
func EmptyMask() TimeMask {
	return TimeMask{}
}

// slotIndex computes the bit index for a given day (0-4) and slot (0-89).
func slotIndex(day, slot int) int {
	return day*slotsPerDay + slot
}

// SetSlot marks a specific time slot as occupied.
func (m *TimeMask) SetSlot(day, slot int) {
	idx := slotIndex(day, slot)
	m[idx/64] |= 1 << (idx % 64)
}

// Conflicts returns true if any time slot is set in both masks.
func (m TimeMask) Conflicts(other TimeMask) bool {
	for i := range 8 {
		if m[i]&other[i] != 0 {
			return true
		}
	}
	return false
}

// Merge returns a new mask with all slots from both masks set.
func (m TimeMask) Merge(other TimeMask) TimeMask {
	var result TimeMask
	for i := range 8 {
		result[i] = m[i] | other[i]
	}
	return result
}

// timeToSlot converts a time string ("1030" or "10:30") to a slot index.
// Returns -1 if the time is invalid or outside the 7am-10pm range.
func timeToSlot(t string) int {
	mins := parseTimeToMins(t)
	if mins < 0 {
		return -1
	}
	slot := (mins - dayStartMins) / 10
	if slot < 0 || slot >= slotsPerDay {
		return -1
	}
	return slot
}

// parseTimeToMins converts a time string to minutes from midnight.
// Accepts formats: "1030", "10:30", "0900", "09:00".
// Returns -1 for invalid or empty strings.
func parseTimeToMins(t string) int {
	if t == "" {
		return -1
	}

	// Remove colon if present
	t = strings.ReplaceAll(t, ":", "")

	if len(t) != 4 {
		return -1
	}

	hours, err := strconv.Atoi(t[:2])
	if err != nil || hours < 0 || hours > 23 {
		return -1
	}

	minutes, err := strconv.Atoi(t[2:])
	if err != nil || minutes < 0 || minutes > 59 {
		return -1
	}

	return hours*60 + minutes
}

// FromMeetingTimes builds a TimeMask from a slice of meeting times.
func FromMeetingTimes(meetings []cache.MeetingTime) TimeMask {
	var mask TimeMask
	for _, mt := range meetings {
		startSlot := timeToSlot(mt.StartTime)
		endSlot := timeToSlot(mt.EndTime)
		if startSlot < 0 || endSlot < 0 {
			continue // Skip TBD or invalid times
		}
		for day := range 5 {
			// Days[0]=Sun, Days[1]=Mon, etc. We want Mon-Fri (indices 1-5)
			if mt.Days[day+1] {
				for slot := startSlot; slot < endSlot; slot++ {
					mask.SetSlot(day, slot)
				}
			}
		}
	}
	return mask
}

// FromBlockedTimes builds a TimeMask from user-specified blocked times.
func FromBlockedTimes(blocked []BlockedTime) TimeMask {
	var mask TimeMask
	for _, bt := range blocked {
		if bt.Day < 0 || bt.Day > 4 {
			continue
		}
		startSlot := timeToSlot(bt.StartTime)
		endSlot := timeToSlot(bt.EndTime)
		if startSlot < 0 || endSlot < 0 {
			continue
		}
		for slot := startSlot; slot < endSlot; slot++ {
			mask.SetSlot(bt.Day, slot)
		}
	}
	return mask
}
