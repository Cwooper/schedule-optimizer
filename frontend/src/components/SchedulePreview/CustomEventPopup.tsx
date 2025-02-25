import React, { useState } from "react";
import Popup from "../Popup/Popup";
import styles from "./CustomEventPopup.module.css";
import type { ScheduleEvent } from "../../types/types";

// Available background colors for custom events
const EVENT_COLORS = [
  "#DCE8FF", // Light Blue
  "#D4FFD4", // Light Green
  "#FFD6D6", // Light Red
  "#EBD6FF", // Light Purple
  "#FFE4CC", // Light Orange
  "#D6FFF7", // Light Cyan
  "#FFD6E8", // Light Pink
  "#E8FFD6", // Light Yellow-Green
];

interface CustomEventPopupProps {
  isOpen: boolean;
  onClose: () => void;
  onAddEvent: (event: Omit<ScheduleEvent, "id">) => void;
}

const CustomEventPopup: React.FC<CustomEventPopupProps> = ({
  isOpen,
  onClose,
  onAddEvent,
}) => {
  // Helper function to convert HH:MM time to minutes
  const timeToMinutes = (time: string): number => {
    const [hours, minutes] = time.split(":").map(Number);
    return hours * 60 + minutes;
  };

  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [startTime, setStartTime] = useState("10:00");
  const [endTime, setEndTime] = useState("11:00");
  const [selectedDays, setSelectedDays] = useState<{
    [key: string]: boolean;
  }>({
    M: false,
    T: false,
    W: false,
    R: false,
    F: false,
  });
  const [selectedColor, setSelectedColor] = useState(EVENT_COLORS[0]);

  const handleDayToggle = (day: string) => {
    setSelectedDays((prev) => ({
      ...prev,
      [day]: !prev[day],
    }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    // Basic validation
    if (!title.trim()) {
      alert("Please enter an event title");
      return;
    }
    
    // Time validation
    const startMinutes = timeToMinutes(startTime);
    const endMinutes = timeToMinutes(endTime);
    
    if (endMinutes <= startMinutes) {
      alert("End time must be after start time. Please adjust your times.");
      return;
    }
    
    // Schedule typically shows 8:00 AM to 10:00 PM
    const minTime = 8 * 60; // 8:00 AM in minutes
    const maxTime = 22 * 60; // 10:00 PM in minutes
    
    if (startMinutes < minTime || endMinutes > maxTime) {
      const confirmAdd = window.confirm(
        "Some of your event may not be visible on the schedule (visible hours are typically 8:00 AM to 10:00 PM). Add anyway?"
      );
      if (!confirmAdd) {
        return;
      }
    }
    
    const days = Object.entries(selectedDays)
      .filter(([_, isSelected]) => isSelected)
      .map(([day]) => {
        // Convert day letter to day index (0-4)
        const dayMap: { [key: string]: number } = {
          M: 0,
          T: 1,
          W: 2,
          R: 3,
          F: 4,
        };
        return dayMap[day];
      });

    if (days.length === 0) {
      alert("Please select at least one day");
      return;
    }

    // Truncate title if too long (over 30 characters)
    const truncatedTitle = title.length > 30
      ? title.substring(0, 27) + "..."
      : title;

    onAddEvent({
      days,
      start: startTime,
      end: endTime,
      title: truncatedTitle,
      body,
      color: selectedColor,
    });

    // Reset form
    setTitle("");
    setBody("");
    setStartTime("10:00");
    setEndTime("11:00");
    setSelectedDays({
      M: false,
      T: false,
      W: false,
      R: false,
      F: false,
    });
    setSelectedColor(EVENT_COLORS[0]);
    
    onClose();
  };

  return (
    <Popup
      isOpen={isOpen}
      onClose={onClose}
      title="Add Custom Event"
      width="500px"
    >
      <form className={styles.form} onSubmit={handleSubmit}>
        <div className={styles.formGroup}>
          <label htmlFor="title" className={styles.label}>
            Event Title:
          </label>
          <input
            id="title"
            type="text"
            className={styles.input}
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="ENG 101"
            required
          />
        </div>

        <div className={styles.formGroup}>
          <label htmlFor="body" className={styles.label}>
            Description:
          </label>
          <textarea
            id="body"
            className={`${styles.input} ${styles.textarea}`}
            value={body}
            onChange={(e) => setBody(e.target.value)}
            placeholder="Smith, John&#10;OM 100"
            rows={3}
          />
        </div>

        <div className={styles.formGroup}>
          <label htmlFor="startTime" className={styles.label}>
            Start Time:
          </label>
          <input
            id="startTime"
            type="time"
            className={styles.input}
            value={startTime}
            onChange={(e) => setStartTime(e.target.value)}
            required
          />
        </div>

        <div className={styles.formGroup}>
          <label htmlFor="endTime" className={styles.label}>
            End Time:
          </label>
          <input
            id="endTime"
            type="time"
            className={styles.input}
            value={endTime}
            onChange={(e) => setEndTime(e.target.value)}
            required
          />
        </div>

        <div className={styles.formGroup}>
          <label className={styles.label}>Days:</label>
          <div className={styles.daysSelection}>
            {Object.entries(selectedDays).map(([day, isSelected]) => (
              <div key={day} className={styles.dayCheckGroup}>
                <input
                  type="checkbox"
                  id={`day-${day}`}
                  checked={isSelected}
                  onChange={() => handleDayToggle(day)}
                />
                <label htmlFor={`day-${day}`} className={styles.checkboxLabel}>
                  {day === "M"
                    ? "Monday"
                    : day === "T"
                    ? "Tuesday"
                    : day === "W"
                    ? "Wednesday"
                    : day === "R"
                    ? "Thursday"
                    : "Friday"}
                </label>
              </div>
            ))}
          </div>
        </div>

        <div className={styles.formGroup}>
          <label className={styles.label}>Event Color:</label>
          <div className={styles.colorSelection}>
            {EVENT_COLORS.map((color) => (
              <div
                key={color}
                className={`${styles.colorOption} ${
                  selectedColor === color ? styles.selected : ""
                }`}
                style={{ backgroundColor: color }}
                onClick={() => setSelectedColor(color)}
                title="Select color"
              />
            ))}
          </div>
        </div>

        <div className={styles.buttons}>
          <button
            type="button"
            className={styles.cancelButton}
            onClick={onClose}
          >
            Cancel
          </button>
          <button type="submit" className={styles.submitButton}>
            Add Event
          </button>
        </div>
      </form>
    </Popup>
  );
};

export default CustomEventPopup;
