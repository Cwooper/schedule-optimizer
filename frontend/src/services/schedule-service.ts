// src/services/schedule-service.ts
import type {
  Schedule,
  ScheduleEvent,
  ScheduleRequest,
  ScheduleResponse,
} from "../types/types";

const BASE_COLORS = [
  "#E6F3FF", // Light Blue
  "#E6FFE6", // Light Green
  "#FFE6E6", // Light Red
  "#F2E6FF", // Light Purple
  "#FFF2E6", // Light Orange
  "#E6FFF2", // Light Cyan
  "#FFE6F2", // Light Pink
  "#F2FFE6", // Light Yellow-Green
];

export const generateScheduleEvents = (schedule: Schedule): ScheduleEvent[] => {
  const events: ScheduleEvent[] = [];
  let colorIndex = 0;

  schedule.Courses.forEach((course) => {
    course.Sessions.forEach((session) => {
      if (session.IsAsync || session.IsTimeTBD) {
        return;
      }

      // Convert days string to array of day indices (0 = Monday, 4 = Friday)
      const dayMap: { [key: string]: number } = {
        M: 0,
        T: 1,
        W: 2,
        R: 3,
        F: 4,
      };
      const days = Array.from(session.Days).map((day) => dayMap[day]);

      // Format time
      const formatTime = (time: number): string => {
        const hours = Math.floor(time / 100);
        const minutes = time % 100;
        return `${hours.toString().padStart(2, "0")}:${minutes
          .toString()
          .padStart(2, "0")}`;
      };

      events.push({
        id: `${course.CRN}-${session.Days}`,
        days,
        start: formatTime(session.StartTime),
        end: formatTime(session.EndTime),
        color: BASE_COLORS[colorIndex % BASE_COLORS.length],
        title: `${course.Subject}`,
        body: `${session.Instructor}\n${session.Location}`,
      });
    });
    colorIndex++;
  });

  return events;
};

export const submitSchedule = async (
  request: ScheduleRequest
): Promise<ScheduleResponse> => {
  try {
    console.log("Request: ", request);

    const response = await fetch("/schedule-optimizer/", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data: ScheduleResponse = await response.json();
    console.log("Response: ", data);
    return data;
  } catch (error) {
    console.error("Error submitting schedule:", error);
    throw error;
  }
};
