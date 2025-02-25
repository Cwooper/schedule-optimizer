import React, { useState, useRef } from "react";
import { Schedule } from "@konnorkooi/schedule-glance";
import { MoreVertical, Download, Plus, RotateCcw } from "lucide-react";
import "@konnorkooi/schedule-glance/dist/index.css";
import type { Schedule as ISchedule, ScheduleEvent } from "../../types/types";
import { generateScheduleEvents } from "../../services/schedule-service";
import { generateScheduleFilename } from "../../utils/schedule-utils";
import styles from "./SchedulePreview.module.css";
import EventPopup from "./EventPopup";
import CourseList from "../CourseList/CourseList";
import CustomEventPopup from "./CustomEventPopup";

interface SchedulePreviewProps {
  schedule?: ISchedule;
  warnings?: string[];
  errors?: string[];
  showMessages?: boolean;
  asyncCourses?: any[];
  quarter: string;
  year: string;
}

const SchedulePreview: React.FC<SchedulePreviewProps> = ({
  schedule,
  warnings = [],
  errors = [],
  showMessages = false,
  asyncCourses = [],
  quarter,
  year,
}) => {
  const scheduleRef = useRef<any>(null);
  const [showActionMenu, setShowActionMenu] = useState(false);
  const actionButtonRef = useRef<HTMLDivElement>(null);
  const [customEvents, setCustomEvents] = useState<ScheduleEvent[]>([]);
  const [isCustomEventPopupOpen, setIsCustomEventPopupOpen] = useState(false);

  const [popupState, setPopupState] = useState<{
    isOpen: boolean;
    courseData?: any;
    sessionData?: any;
  }>({
    isOpen: false,
    courseData: null,
    sessionData: null,
  });

  const events = schedule 
    ? [...generateScheduleEvents(schedule), ...customEvents] 
    : [...customEvents];

  const handleExportSchedule = async () => {
    try {
      const filename = generateScheduleFilename(quarter, year);
      await scheduleRef.current?.exportToPng(filename);
      setShowActionMenu(false);
    } catch (error) {
      console.error("Failed to export schedule:", error);
    }
  };

  const handleAddCustomEvent = (newEvent: Omit<ScheduleEvent, "id">) => {
    // Make sure newlines in the body are preserved
    // The Schedule component expects newlines to be actual \n characters
    const eventWithId: ScheduleEvent = {
      ...newEvent,
      id: `custom-${Date.now()}`,
      // Ensure body text displays newlines correctly
      body: newEvent.body || "",
    };
    setCustomEvents((prev) => [...prev, eventWithId]);
  };

  const handleResetCustomEvents = () => {
    setCustomEvents([]);
    setShowActionMenu(false);
  };

  const customHeaders = [
    { label: "Mon", dayIndex: 0 },
    { label: "Tue", dayIndex: 1 },
    { label: "Wed", dayIndex: 2 },
    { label: "Thu", dayIndex: 3 },
    { label: "Fri", dayIndex: 4 },
  ];

  const handlePopupClose = () => {
    setPopupState({ isOpen: false, courseData: null, sessionData: null });
  };

  const handleClickOutside = (event: MouseEvent) => {
    if (
      actionButtonRef.current &&
      !actionButtonRef.current.contains(event.target as Node)
    ) {
      setShowActionMenu(false);
    }
  };

  React.useEffect(() => {
    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  // Clear custom events when schedule changes
  React.useEffect(() => {
    setCustomEvents([]);
  }, [schedule]);

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

      <div className={styles.scheduleWithActions}>
        <div ref={actionButtonRef} className={styles.actionMenuContainer}>
          <button
            onClick={() => setShowActionMenu(!showActionMenu)}
            className={styles.actionMenuButton}
            title="Schedule Actions"
          >
            <MoreVertical className="w-5 h-5" />
          </button>

          {showActionMenu && (
            <div className={styles.actionMenuDropdown}>
              <button
                onClick={handleExportSchedule}
                className={styles.actionMenuItem}
                disabled={!events.length}
                title="Export Schedule"
              >
                <Download className="w-4 h-4" />
              </button>
              <button
                onClick={() => {
                  setIsCustomEventPopupOpen(true);
                  setShowActionMenu(false);
                }}
                className={styles.actionMenuItem}
                title="Add Custom Event"
              >
                <Plus className="w-4 h-4" />
              </button>
              <button
                onClick={handleResetCustomEvents}
                className={styles.actionMenuItem}
                disabled={customEvents.length === 0}
                title="Reset Custom Events"
              >
                <RotateCcw className="w-4 h-4" />
              </button>
            </div>
          )}
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
              // Skip custom events
              if (event.id.startsWith("custom-")) {
                return null;
              }
              
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
                quarter={quarter}
                year={year}
              />
            )}
        </div>
      </div>

      <CustomEventPopup
        isOpen={isCustomEventPopupOpen}
        onClose={() => setIsCustomEventPopupOpen(false)}
        onAddEvent={handleAddCustomEvent}
      />

      {asyncCourses && asyncCourses.length > 0 && (
        <CourseList
          courses={asyncCourses}
          title="Asynchronous Courses"
          emptyMessage="No asynchronous courses"
          quarter={quarter}
          year={year}
        />
      )}
    </div>
  );
};

export default SchedulePreview;
