// src/services/schedule-service.ts
import type {
  Schedule,
  ScheduleEvent,
  ScheduleRequest,
  ScheduleResponse,
} from "../types/types";

const BASE_COLORS = [
  "#DCE8FF", // Light Blue
  "#D4FFD4", // Light Green
  "#FFD6D6", // Light Red
  "#EBD6FF", // Light Purple
  "#FFE4CC", // Light Orange
  "#D6FFF7", // Light Cyan
  "#FFD6E8", // Light Pink
  "#E8FFD6", // Light Yellow-Green
];

const getInstructor = (currentSession: any, allSessions: any[]): string => {
  if (
    currentSession.Instructor &&
    currentSession.Instructor.toLowerCase() !== "staff"
  ) {
    return currentSession.Instructor;
  }

  // Look for non-"Staff" instructor in other sessions
  const otherInstructor = allSessions.find(
    (s: any) => s.Instructor && s.Instructor.toLowerCase() !== "staff"
  );

  return otherInstructor
    ? otherInstructor.Instructor
    : currentSession.Instructor;
};

export const generateScheduleEvents = (schedule: Schedule): ScheduleEvent[] => {
  const events: ScheduleEvent[] = [];
  let colorIndex = 0;

  schedule.Courses.forEach((course) => {
    course.Sessions.forEach((session) => {
      if (session.IsAsync || session.IsTimeTBD) {
        return;
      }

      const dayMap: { [key: string]: number } = {
        M: 0,
        T: 1,
        W: 2,
        R: 3,
        F: 4,
      };
      const days = Array.from(session.Days).map((day) => dayMap[day]);

      const formatTime = (time: number): string => {
        const hours = Math.floor(time / 100);
        const minutes = time % 100;
        return `${hours.toString().padStart(2, "0")}:${minutes
          .toString()
          .padStart(2, "0")}`;
      };

      // Get the best instructor to display
      const displayInstructor = getInstructor(session, course.Sessions);

      events.push({
        id: `${String(course.CRN)}-${session.Days}`,
        days,
        start: formatTime(session.StartTime),
        end: formatTime(session.EndTime),
        color: BASE_COLORS[colorIndex % BASE_COLORS.length],
        title: `${course.Subject}`,
        body: `${displayInstructor}\n${session.Location}`,
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
