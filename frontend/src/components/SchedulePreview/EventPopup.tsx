// src/components/SchedulePreview/EventPopup.tsx
import React from "react";
import Popup from "../Popup/Popup";
import styles from "./EventPopup.module.css";

interface EventPopupProps {
  course: {
    Subject: string;
    Title: string;
    CRN: string;
    GPA?: number;
    Credits: number;
    Prerequisites?: string;
    Attributes?: string;
    AdditionalFees?: string;
    Restrictions?: string;
    AvailableSeats: number;
    Capacity: number;
    Enrolled: number;
  };
  session: {
    Days: string;
    StartTime: number;
    EndTime: number;
    Instructor: string;
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
  // Find a non-"Staff" instructor from other sessions if current instructor is "Staff"
  const getInstructor = (currentSession: any, course: any): string => {
    if (
      currentSession.Instructor &&
      currentSession.Instructor.toLowerCase() !== "staff"
    ) {
      return currentSession.Instructor;
    }

    // Look for non-"Staff" instructor in other sessions
    const otherInstructor = course.Sessions.find(
      (s: any) => s.Instructor && s.Instructor.toLowerCase() !== "staff"
    );

    return otherInstructor
      ? otherInstructor.Instructor
      : currentSession.Instructor;
  };

  const displayInstructor = getInstructor(session, course);
  // Removed unused isAsyncOrTBD variable

  return (
    <div>
      <div className={styles.stats}>
        <div className={styles.statItem}>
          <div className={styles.statLabel}>CRN</div>
          <div className={styles.statValue}>{course.CRN}</div>
        </div>
        {course.GPA !== undefined && (
          <div className={styles.statItem}>
            <div className={styles.statLabel}>GPA</div>
            <div className={styles.statValue}>{course.GPA.toFixed(2)}</div>
          </div>
        )}
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
          <span className={styles.value}>{displayInstructor}</span>
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

      {(course.Prerequisites ||
        course.Attributes ||
        course.AdditionalFees ||
        course.Restrictions) && (
        <div className={styles.additionalInfo}>
          <h3 className={styles.sectionTitle}>Additional Information</h3>
          <div className={styles.detailsList}>
            {course.Prerequisites && (
              <>
                <span className={styles.label}>Prerequisites:</span>
                <span className={styles.value}>{course.Prerequisites}</span>
              </>
            )}
            {course.Attributes && (
              <>
                <span className={styles.label}>Attributes:</span>
                <div>
                  {course.Attributes.split(" ").map((attr) => (
                    <span key={attr} className={styles.infoTag}>
                      {attr}
                    </span>
                  ))}
                </div>
              </>
            )}
            {course.AdditionalFees && (
              <>
                <span className={styles.label}>Additional Fees:</span>
                <span className={styles.value}>{course.AdditionalFees}</span>
              </>
            )}
            {course.Restrictions && (
              <>
                <span className={styles.label}>Restrictions:</span>
                <div>
                  {course.Restrictions.split(" ").map((restriction) => (
                    <span key={restriction} className={styles.infoTag}>
                      {restriction}
                    </span>
                  ))}
                </div>
              </>
            )}
          </div>
        </div>
      )}
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
