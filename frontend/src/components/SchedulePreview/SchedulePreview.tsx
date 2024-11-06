import React, { useState, useRef } from "react";
import { Schedule } from "@konnorkooi/schedule-glance";
import { Download, Map } from "lucide-react";
import "@konnorkooi/schedule-glance/dist/index.css";
import type { Schedule as ISchedule } from "../../types/types";
import { generateScheduleEvents } from "../../services/schedule-service";
import styles from "./SchedulePreview.module.css";
import EventPopup from "./EventPopup";
import CourseList from "../CourseList/CourseList";

interface SchedulePreviewProps {
  schedule?: ISchedule;
  warnings?: string[];
  errors?: string[];
  showMessages?: boolean;
  asyncCourses?: any[];
}

const SchedulePreview: React.FC<SchedulePreviewProps> = ({
  schedule,
  warnings = [],
  errors = [],
  showMessages = false,
  asyncCourses = [],
}) => {
  const scheduleRef = useRef<any>(null);
  const [showMap, setShowMap] = useState(false);
  const [popupState, setPopupState] = useState<{
    isOpen: boolean;
    courseData?: any;
    sessionData?: any;
  }>({
    isOpen: false,
    courseData: null,
    sessionData: null,
  });

  const handleExportSchedule = async () => {
    try {
      await scheduleRef.current?.exportToPng(
        `schedule-${new Date().toISOString()}.png`
      );
    } catch (error) {
      console.error("Failed to export schedule:", error);
    }
  };

  const handleShowMap = () => {
    setShowMap(!showMap);
    // TODO: map logic here when added
  };

  const customHeaders = [
    { label: "Mon", dayIndex: 0 },
    { label: "Tue", dayIndex: 1 },
    { label: "Wed", dayIndex: 2 },
    { label: "Thu", dayIndex: 3 },
    { label: "Fri", dayIndex: 4 },
  ];

  const events = schedule ? generateScheduleEvents(schedule) : [];

  const handlePopupClose = () => {
    setPopupState({ isOpen: false, courseData: null, sessionData: null });
  };

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

      <div className={styles.scheduleHeader}>
        <div className={styles.scheduleTitle}>Schedule Preview</div>
        <div className={styles.actionButtons}>
          <button
            onClick={handleExportSchedule}
            className={styles.actionButton}
            disabled={!schedule || !events.length}
            title="Export Schedule as PNG"
          >
            <Download className="w-4 h-4 mr-2" />
            Export
          </button>
          <button
            onClick={handleShowMap}
            className={styles.actionButton}
            disabled={!schedule || !events.length}
            title="View Walking Map"
          >
            <Map className="w-4 h-4 mr-2" />
            Map
          </button>
        </div>
      </div>

      <div className={styles.scheduleContainer}>
        <Schedule
          ref={scheduleRef}
          events={events}
          headers={customHeaders}
          width={800}
          height={600}
          useDefaultPopup={false}
          emptyStateMessage="No courses selected for this schedule"
          customPopupHandler={(event) => {
            const [crn, days] = event.id.split("-");
            const courseData = schedule?.Courses.find(
              (course) => String(course.CRN) === String(crn)
            );
            const sessionData = courseData?.Sessions.find(
              (session) => session.Days === days
            );

            if (courseData && sessionData) {
              setPopupState({
                isOpen: true,
                courseData,
                sessionData,
              });
            }
            return null;
          }}
        />

        {popupState.isOpen &&
          popupState.courseData &&
          popupState.sessionData && (
            <EventPopup
              course={popupState.courseData}
              session={popupState.sessionData}
              isOpen={popupState.isOpen}
              onClose={handlePopupClose}
            />
          )}
      </div>

      {asyncCourses && asyncCourses.length > 0 && (
        <CourseList
          courses={asyncCourses}
          title="Asynchronous Courses"
          emptyMessage="No asynchronous courses"
        />
      )}
    </div>
  );
};

export default SchedulePreview;
