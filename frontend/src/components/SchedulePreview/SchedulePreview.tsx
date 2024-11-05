// src/components/SchedulePreview/SchedulePreview.tsx
import React from "react";
import { Schedule } from "@konnorkooi/schedule-glance";
import "@konnorkooi/schedule-glance/dist/index.css";
import type { Schedule as ISchedule } from "../../types/types";
import { generateScheduleEvents } from "../../services/schedule-service";
import styles from "./SchedulePreview.module.css";

interface SchedulePreviewProps {
  schedule?: ISchedule;
  warnings?: string[];
  errors?: string[];
  showMessages?: boolean;
}

const SchedulePreview: React.FC<SchedulePreviewProps> = ({
  schedule,
  warnings = [],
  errors = [],
  showMessages = false,
}) => {
  const customHeaders = [
    { label: "Mon", dayIndex: 0 },
    { label: "Tue", dayIndex: 1 },
    { label: "Wed", dayIndex: 2 },
    { label: "Thu", dayIndex: 3 },
    { label: "Fri", dayIndex: 4 },
  ];

  const events = schedule ? generateScheduleEvents(schedule) : [];

  return (
    <div className={styles.container}>
      {showMessages && (errors.length > 0 || warnings.length > 0) && (
        <div
          className={`${styles.messageContainer} ${
            errors.length > 0 ? styles.errorContainer : styles.warningContainer
          }`}
        >
          {errors.map((error) => (
            <div key={`error-${error}`} className={styles.message}>
              {error}
            </div>
          ))}
          {warnings.map((warning) => (
            <div key={`warning-${warning}`} className={styles.message}>
              {warning}
            </div>
          ))}
        </div>
      )}
      <div className={styles.scheduleContainer}>
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
