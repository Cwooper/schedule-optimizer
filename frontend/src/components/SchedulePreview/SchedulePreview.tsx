import React from "react";
import { Schedule } from "@konnorkooi/schedule-glance";
import "@konnorkooi/schedule-glance/dist/index.css";
import type { Schedule as ISchedule } from "../../types/types";
import { generateScheduleEvents } from "../../services/schedule-service";

interface SchedulePreviewProps {
  schedule?: ISchedule;
  warnings?: string[];
  errors?: string[];
}

const SchedulePreview: React.FC<SchedulePreviewProps> = ({
  schedule,
  warnings = [],
  errors = [],
}) => {
  // Custom headers for Monday-Friday schedule
  const customHeaders = [
    { label: "Mon", dayIndex: 0 },
    { label: "Tue", dayIndex: 1 },
    { label: "Wed", dayIndex: 2 },
    { label: "Thu", dayIndex: 3 },
    { label: "Fri", dayIndex: 4 },
  ];

  const events = schedule ? generateScheduleEvents(schedule) : [];

  const getMessageType = () => {
    if (errors.length > 0) return "error";
    if (warnings.length > 0) return "warning";
    return null;
  };

  const getMessage = () => {
    if (errors.length > 0) return errors.join(". ");
    if (warnings.length > 0) return warnings.join(". ");
    return null;
  };

  const messageType = getMessageType();
  const message = getMessage();

  return (
    <div className="w-full max-w-4xl mx-auto p-4">
      {messageType && message && (
        <div
          className={`mb-4 p-4 rounded-lg ${
            messageType === "error"
              ? "bg-red-100 text-red-700 border border-red-300"
              : "bg-yellow-100 text-yellow-700 border border-yellow-300"
          }`}
        >
          {message}
        </div>
      )}
      <div className="bg-white rounded-lg shadow-md">
        <Schedule
          events={events}
          headers={customHeaders}
          width={800}
          height={600}
        />
      </div>
    </div>
  );
};

export default SchedulePreview;
