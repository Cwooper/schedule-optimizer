import React, { useState } from "react";
import styles from "./SchedulePlanner.module.css";
import type { Schedule, Course } from "../../types/types";

interface WeightConfig {
  importance: number;
  idealValue?: number;
}

interface WeightsState {
  [key: string]: WeightConfig;
}

interface WeightsPopupProps {
  schedules: Schedule[];
  onApplyWeights: (sortedSchedules: Schedule[], weights: WeightsState) => void;
  onClose: () => void;
}

interface DailySession {
  startTime: number;
  endTime: number;
}

// Helper function to extract course times from schedule
const extractScheduleTimes = (courses: Course[]) => {
  let startTime = 2359;
  let endTime = 0;
  let totalGPA = 0;
  let gpaCount = 0;

  // Create a map of sessions by day
  const sessionsByDay: { [key: number]: DailySession[] } = {};

  // First collect all sessions by day
  courses.forEach((course) => {
    course.Sessions.forEach((session) => {
      if (!session.IsAsync && !session.IsTimeTBD) {
        // Convert day string (e.g., 'MWF') to array of day indices
        const daysArray = session.Days.split("").map((day) => {
          // Convert MTWRF to numbers (0-4)
          const dayMap: { [key: string]: number } = {
            M: 0,
            T: 1,
            W: 2,
            R: 3,
            F: 4,
          };
          return dayMap[day];
        });

        daysArray.forEach((day) => {
          if (day !== undefined) {
            // Skip any invalid day characters
            if (!sessionsByDay[day]) {
              sessionsByDay[day] = [];
            }
            sessionsByDay[day].push({
              startTime: session.StartTime,
              endTime: session.EndTime,
            });
          }
        });

        // Update overall start and end times
        if (session.StartTime < startTime) startTime = session.StartTime;
        if (session.EndTime > endTime) endTime = session.EndTime;
      }
    });

    if (course.GPA) {
      totalGPA += course.GPA;
      gpaCount++;
    }
  });

  // Calculate average gap time across all days
  let totalGapTime = 0;
  let daysWithClasses = 0;

  Object.values(sessionsByDay).forEach((daySessions) => {
    if (daySessions.length > 0) {
      daysWithClasses++;

      // Sort sessions by start time
      daySessions.sort((a, b) => a.startTime - b.startTime);

      // Calculate gaps between sessions for this day
      let dayGapTime = 0;
      for (let i = 1; i < daySessions.length; i++) {
        const gap = daySessions[i].startTime - daySessions[i - 1].endTime;
        if (gap > 0) {
          dayGapTime += gap;
        }
      }

      totalGapTime += dayGapTime;
    }
  });

  // Calculate average gap time per day
  const averageGapTime =
    daysWithClasses > 0 ? totalGapTime / daysWithClasses : 0;

  return {
    startTime,
    endTime,
    gapTime: averageGapTime,
    averageGPA: gpaCount > 0 ? totalGPA / gpaCount : 0,
  };
};

const WeightsPopup: React.FC<WeightsPopupProps> = ({
  schedules,
  onApplyWeights,
  onClose,
}) => {
  // Initialize weights with default values (using minutes since midnight)
  const [weights, setWeights] = useState<WeightsState>({
    "Start Time": { importance: 1, idealValue: 600 }, // 10:00 AM in minutes (9 * 60)
    "End Time": { importance: 1, idealValue: 780 }, // 1:00 PM in minutes (15 * 60)
    "Gap Time": { importance: 2, idealValue: 0 },
    GPA: { importance: 2 },
  });

  // Convert HH:MM time string to minutes since midnight
  const timeToMinutes = (timeStr: string): number => {
    const [hours, minutes] = timeStr.split(":").map(Number);
    return hours * 60 + Math.round(minutes / 10) * 10;
  };

  // Convert minutes since midnight to HH:MM format
  const minutesToTime = (totalMinutes: number): string => {
    const hours = Math.floor(totalMinutes / 60);
    const minutes = Math.round((totalMinutes % 60) / 10) * 10;
    return `${hours.toString().padStart(2, "0")}:${minutes
      .toString()
      .padStart(2, "0")}`;
  };

  // Convert military time (HHMM) to minutes since midnight
  const militaryToMinutes = (militaryTime: number): number => {
    const hours = Math.floor(militaryTime / 100);
    const minutes = militaryTime % 100;
    return hours * 60 + minutes;
  };

  const handleWeightChange = (
    category: string,
    field: "importance" | "idealValue",
    value: string
  ) => {
    setWeights((prev) => ({
      ...prev,
      [category]: {
        ...prev[category],
        [field]:
          field === "idealValue" &&
          (category === "Start Time" || category === "End Time")
            ? timeToMinutes(value)
            : parseFloat(value) || 0,
      },
    }));
  };

  const calculateLinearScore = (
    actual: number,
    ideal: number,
    range: number
  ): number => {
    const difference = Math.abs(actual - ideal);
    return Math.max(0, 1 - difference / range);
  };

  const calculateScheduleScore = (schedule: Schedule): number => {
    let totalScore = 0;
    let totalImportance = 0;

    const times = extractScheduleTimes(schedule.Courses);

    // Convert military time to minutes for comparison
    const startTimeMinutes = militaryToMinutes(times.startTime);
    const endTimeMinutes = militaryToMinutes(times.endTime);

    // Start Time (using 120 minutes as range - 2 hour window)
    if (weights["Start Time"].importance > 0) {
      const score = calculateLinearScore(
        startTimeMinutes,
        weights["Start Time"].idealValue!,
        120
      );
      totalScore += score * weights["Start Time"].importance;
      totalImportance += weights["Start Time"].importance;
    }

    // End Time (using 120 minutes as range - 2 hour window)
    if (weights["End Time"].importance > 0) {
      const score = calculateLinearScore(
        endTimeMinutes,
        weights["End Time"].idealValue!,
        120
      );
      totalScore += score * weights["End Time"].importance;
      totalImportance += weights["End Time"].importance;
    }

    // Gap Time (using 60 minutes as range)
    if (weights["Gap Time"].importance > 0) {
      const score = calculateLinearScore(
        times.gapTime,
        weights["Gap Time"].idealValue!,
        60
      );
      totalScore += score * weights["Gap Time"].importance;
      totalImportance += weights["Gap Time"].importance;
    }

    // GPA
    if (weights["GPA"].importance > 0 && times.averageGPA > 0) {
      const score = times.averageGPA / 4.0;
      totalScore += score * weights["GPA"].importance;
      totalImportance += weights["GPA"].importance;
    }

    const finalScore = totalImportance > 0 ? totalScore / totalImportance : 0;
    return finalScore;
  };

  const handleApply = () => {
    const sortedSchedules = [...schedules].map((schedule) => ({
      ...schedule,
      Score: calculateScheduleScore(schedule),
    }));

    // Sort schedules by score in descending order
    sortedSchedules.sort((a, b) => (b.Score || 0) - (a.Score || 0));

    onApplyWeights(sortedSchedules, weights);
    onClose();
  };

  return (
    <div className={styles.weightsContainer}>
      <p className={styles.weightsDescription}>
        Adjust the importance and ideal values for each scheduling factor.
        Higher importance values give more weight to that factor.
      </p>

      <div className={styles.weightsList}>
        {Object.entries(weights).map(([category, config]) => (
          <div key={category} className={styles.weightItem}>
            <div className={styles.weightHeader}>{category}</div>

            <div className={styles.weightControls}>
              <div className={styles.controlGroup}>
                <label className={styles.weightLabel}>Importance:</label>
                <input
                  type="number"
                  min="0"
                  max="10"
                  step="0.5"
                  value={config.importance}
                  onChange={(e) =>
                    handleWeightChange(category, "importance", e.target.value)
                  }
                  className={styles.weightInput}
                />
              </div>

              {config.idealValue !== undefined && (
                <div className={styles.controlGroup}>
                  <label className={styles.weightLabel}>
                    {category === "Gap Time"
                      ? "Ideal Gap (mins/day):"
                      : "Ideal Time:"}
                  </label>
                  {category === "Gap Time" ? (
                    <input
                      type="number"
                      min="0"
                      max="240"
                      step="10"
                      value={config.idealValue}
                      onChange={(e) =>
                        handleWeightChange(
                          category,
                          "idealValue",
                          e.target.value
                        )
                      }
                      className={styles.weightInput}
                    />
                  ) : (
                    <input
                      type="time"
                      value={minutesToTime(config.idealValue)}
                      onChange={(e) =>
                        handleWeightChange(
                          category,
                          "idealValue",
                          e.target.value
                        )
                      }
                      step="600"
                      className={styles.weightInput}
                    />
                  )}
                </div>
              )}
            </div>
          </div>
        ))}
      </div>

      <button
        onClick={handleApply}
        className={`btn btn-primary ${styles.applyButton}`}
      >
        Apply Weights
      </button>
    </div>
  );
};

export default WeightsPopup;
