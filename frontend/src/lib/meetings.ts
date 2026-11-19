import type { MeetingTime } from "@/lib/api"
import { DAY_LETTERS } from "@/lib/schedule-utils"

function formatDays(days: boolean[]): string {
  return days
    .map((active, i) => (active ? DAY_LETTERS[i] : null))
    .filter(Boolean)
    .join("")
}

export function formatTime(time: string): string {
  if (!time) return ""
  const clean = time.replace(":", "").padStart(4, "0")
  const hours = parseInt(clean.slice(0, 2), 10)
  const mins = clean.slice(2, 4)
  const suffix = hours >= 12 ? "pm" : "am"
  let displayHour = hours
  if (hours > 12) {
    displayHour = hours - 12
  } else if (hours === 0) {
    displayHour = 12
  }
  return `${displayHour}:${mins}${suffix}`
}

function isAsync(meeting: MeetingTime): boolean {
  return meeting.room === "ASNC"
}

export function formatMeetingTime(meeting: MeetingTime): string {
  const { startTime, endTime } = meeting
  if (meeting.days.some(Boolean) && startTime && endTime) {
    return `${formatDays(meeting.days)} ${formatTime(startTime)}-${formatTime(endTime)}`
  }
  if (isAsync(meeting)) return "Asynchronous, no set times"
  return "Times arranged with instructor"
}

export function formatLocation(meeting: MeetingTime): string | null {
  const building = meeting.building || null
  // ASNC and SYNC are schedule modifiers, not rooms.
  const room =
    meeting.room && meeting.room !== "ASNC" && meeting.room !== "SYNC"
      ? meeting.room
      : null
  if (!building && !room) return null
  if (!building) return room
  if (!room) return building
  return `${building} ${room}`
}

/** "4", or "1-3" for a variable-credit course. */
export function formatCredits(credits: number, creditsHigh?: number): string {
  if (creditsHigh && creditsHigh !== credits) {
    return `${credits}-${creditsHigh}`
  }
  return String(credits)
}
