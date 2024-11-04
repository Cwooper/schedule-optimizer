import React from "react";
import { Schedule } from "@konnorkooi/schedule-glance";
import "@konnorkooi/schedule-glance/dist/index.css";

const SchedulePreview = () => {
  // Example events to demonstrate the schedule
  const sampleEvents = [
    {
      id: "1",
      days: [1, 3], // Tuesday, Thursday
      start: "09:30",
      end: "10:50",
      color: "#acddef", // Indigo color for WWU theme
      title: "CSCI 241",
      body: "Data Structures",
    },
    {
      id: "2",
      days: [0, 2, 4], // Monday, Wednesday, Friday
      start: "11:00",
      end: "11:50",
      color: "#caf1de", // Blue color for variety
      title: "MATH 204",
      body: "Linear Algebra",
    },
    {
      id: "3",
      days: [1, 3], // Tuesday, Thursday
      start: "13:00",
      end: "14:20",
      color: "#ffe7c7", // Purple color for contrast
      title: "CSCI 305",
      body: "Algorithm Analysis",
    },
  ];

  // Custom headers for Monday-Friday schedule
  const customHeaders = [
    { label: "Mon", dayIndex: 0 },
    { label: "Tue", dayIndex: 1 },
    { label: "Wed", dayIndex: 2 },
    { label: "Thu", dayIndex: 3 },
    { label: "Fri", dayIndex: 4 },
  ];

  return (
    <div className="w-full max-w-4xl mx-auto p-4">
      <div className="bg-white rounded-lg shadow-md">
        <Schedule
          events={sampleEvents}
          headers={customHeaders}
          width={800}
          height={600}
        />
      </div>
    </div>
  );
};

export default SchedulePreview;
