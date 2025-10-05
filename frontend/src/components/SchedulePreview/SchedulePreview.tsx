import React, { useState, useRef, useEffect } from "react";
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
  const [editingCustomEvent, setEditingCustomEvent] = useState<ScheduleEvent | null>(null);
  const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);
  const scheduleWrapperRef = useRef<HTMLDivElement>(null);

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

  // Calculate appropriate height based on schedule content
  const calculateScheduleHeight = () => {
    // Base heights
    const baseDesktopHeight = 600;
    const baseMobileHeight = 500; // Increased from 400 to provide more space by default

    if (events.length === 0) {
      return isMobile ? baseMobileHeight : baseDesktopHeight;
    }

    // Find earliest start time and latest end time
    let earliestStart = 24 * 60; // 24 hours in minutes
    let latestEnd = 0;

    events.forEach(event => {
      const [startHours, startMinutes] = event.start.split(':').map(Number);
      const [endHours, endMinutes] = event.end.split(':').map(Number);

      const startInMinutes = startHours * 60 + startMinutes;
      const endInMinutes = endHours * 60 + endMinutes;

      earliestStart = Math.min(earliestStart, startInMinutes);
      latestEnd = Math.max(latestEnd, endInMinutes);
    });

    // Calculate total hours (rounded up)
    const totalHours = Math.ceil((latestEnd - earliestStart) / 60);

    // Add 2 hours of padding (1 before and 1 after)
    const displayHours = totalHours + 2;

    // Calculate height based on hours
    // On desktop: ~80px per hour is good
    // On mobile: ~70px per hour
    const hourHeight = isMobile ? 70 : 80;
    const calculatedHeight = displayHours * hourHeight;

    // Enforce minimum heights
    const minHeight = isMobile ? baseMobileHeight : baseDesktopHeight;

    return Math.max(calculatedHeight, minHeight);
  };

  const scheduleHeight = calculateScheduleHeight();
  // Fixed width for schedule component - matches the container width
  const scheduleWidth = 800;

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

  const handleUpdateCustomEvent = (eventId: string, updatedEvent: Omit<ScheduleEvent, "id">) => {
    setCustomEvents((prev) =>
      prev.map((event) =>
        event.id === eventId
          ? { ...updatedEvent, id: eventId }
          : event
      )
    );
  };

  const handleDeleteCustomEvent = (eventId: string) => {
    setCustomEvents((prev) => prev.filter((event) => event.id !== eventId));
  };

  const handleEditCustomEvent = (event: ScheduleEvent) => {
    setEditingCustomEvent(event);
    setIsCustomEventPopupOpen(true);
  };

  const handleCloseCustomEventPopup = () => {
    setIsCustomEventPopupOpen(false);
    setEditingCustomEvent(null);
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

  useEffect(() => {
    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  // Handle window resize and update isMobile state
  useEffect(() => {
    const handleResize = () => {
      setIsMobile(window.innerWidth <= 768);
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // Clear custom events when schedule changes
  useEffect(() => {
    setCustomEvents([]);
  }, [schedule]);

  return (
    <div className={styles.container}>
      {showMessages && (errors.length > 0 || warnings.length > 0) && (
        <div
          className={`${styles.messageContainer} ${errors.length > 0 ? styles.errorContainer : styles.warningContainer
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

        {/* Added wrapper div for horizontal scrolling */}
        <div className={styles.scheduleWrapper} ref={scheduleWrapperRef}>
          <div
            className={styles.scheduleContainer}
            style={{ height: `${scheduleHeight}px` }}
          >
            <Schedule
              ref={scheduleRef}
              events={events}
              headers={customHeaders}
              width={scheduleWidth}
              height={scheduleHeight}
              useDefaultPopup={false}
              emptyStateMessage="No courses selected for this schedule"
              customPopupHandler={(event) => {
                // Handle custom events
                if (event.id.startsWith("custom-")) {
                  const customEvent = customEvents.find((e) => e.id === event.id);
                  if (customEvent) {
                    handleEditCustomEvent(customEvent);
                  }
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
          </div>
        </div>
      </div>

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

      <CustomEventPopup
        isOpen={isCustomEventPopupOpen}
        onClose={handleCloseCustomEventPopup}
        onAddEvent={handleAddCustomEvent}
        editingEvent={editingCustomEvent}
        onUpdateEvent={handleUpdateCustomEvent}
        onDeleteEvent={handleDeleteCustomEvent}
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
