// src/components/SchedulePreview/EventPopup.tsx
import React from "react";
import Popup from "../Popup/Popup";
import styles from "./EventPopup.module.css";

interface EventPopupProps {
  course: {
    Subject: string;
    Title: string;
    Credits: number;
    CRN: string;
    Instructor: string;
    GPA?: number;
    Capacity: number;
    Enrolled: number;
    AvailableSeats: number;
  };
  session: {
    Days: string;
    StartTime: number;
    EndTime: number;
    Location: string;
    IsAsync?: boolean;
    IsTimeTBD?: boolean;
  };
  isOpen: boolean;
  onClose: () => void;
}

const formatTime = (time: number): string => {
  const hours = Math.floor(time / 100);
  const minutes = time % 100;
  const period = hours >= 12 ? "pm" : "am";
  const formattedHours = hours > 12 ? hours - 12 : hours;
  return `${formattedHours}:${minutes.toString().padStart(2, "0")}${period}`;
};

const formatDays = (days: string): string => {
  if (!days || days === "TBD" || days === "N/A") {
    return days;
  }

  const dayMap: { [key: string]: string } = {
    M: "Monday",
    T: "Tuesday",
    W: "Wednesday",
    R: "Thursday",
    F: "Friday",
  };

  return Array.from(days)
    .map((day) => dayMap[day] || day)
    .join(", ");
};

const EventPopupContent: React.FC<EventPopupProps> = ({ course, session }) => {
  return (
    <div>
      <div className={styles.stats}>
        <div className={styles.statItem}>
          <div className={styles.statLabel}>CRN</div>
          <div className={styles.statValue}>{course.CRN}</div>
        </div>
        {course.GPA !== undefined && course.GPA !== 0.0 ? (
          <div className={styles.statItem}>
            <div className={styles.statLabel}>GPA</div>
            <div className={styles.statValue}>{course.GPA.toFixed(2)}</div>
          </div>
        ) : course.GPA === 0.0 ? (
          <div className={styles.statItem}>
            <div className={styles.statLabel}>GPA</div>
            <div className={styles.statValue}>N/A</div>
          </div>
        ) : null}
        <div className={styles.statItem}>
          <div className={styles.statLabel}>Credits</div>
          <div className={styles.statValue}>{course.Credits}</div>
        </div>
      </div>

      <div className={styles.section}>
        <h3 className={styles.sectionTitle}>Schedule</h3>
        <div className={styles.detailsList}>
          {session.Days && !session.IsAsync && (
            <>
              <span className={styles.label}>Days:</span>
              {session.Days === "TBD" || session.Days === "N/A" ? (
                <span className={styles.asyncBadge}>{session.Days}</span>
              ) : (
                <span className={styles.value}>{formatDays(session.Days)}</span>
              )}
            </>
          )}

          {!session.IsTimeTBD &&
            !session.IsAsync &&
            Boolean(session.StartTime || session.EndTime) && (
              <>
                <span className={styles.label}>Time:</span>
                <span className={styles.value}>
                  {formatTime(session.StartTime)} -{" "}
                  {formatTime(session.EndTime)}
                </span>
              </>
            )}

          {session.IsTimeTBD && (
            <>
              <span className={styles.label}>Time:</span>
              <span className={styles.asyncBadge}>TBD</span>
            </>
          )}

          {session.Location && (
            <>
              <span className={styles.label}>Location:</span>
              <span className={styles.highlight}>{session.Location}</span>
            </>
          )}

          <span className={styles.label}>Instructor:</span>
          <span className={styles.value}>{course.Instructor || "Staff"}</span>
        </div>
      </div>

      <div className={styles.section}>
        <h3 className={styles.sectionTitle}>Enrollment</h3>
        <div className={styles.detailsList}>
          <span className={styles.label}>Available Seats:</span>
          <span className={styles.value}>{course.AvailableSeats}</span>

          <span className={styles.label}>Max Students:</span>
          <span className={styles.value}>{course.Capacity}</span>

          <span className={styles.label}>Students Enrolled:</span>
          <span className={styles.value}>{course.Enrolled}</span>
        </div>
      </div>
    </div>
  );
};

const EventPopup: React.FC<EventPopupProps> = ({
  course,
  session,
  isOpen,
  onClose,
}) => {
  return (
    <Popup
      isOpen={isOpen}
      onClose={onClose}
      title={`${course.Subject} ${course.Title}`}
      width="600px"
    >
      <EventPopupContent
        course={course}
        session={session}
        isOpen={isOpen}
        onClose={onClose}
      />
    </Popup>
  );
};

export default EventPopup;
