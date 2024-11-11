import React, { useState, useEffect } from "react";
import {
  Clock,
  Sun,
  Moon,
  ArrowDownUp,
  GraduationCap,
  Check,
} from "lucide-react";
import styles from "./WeightsPopup.module.css";
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
  weights: WeightsState;
  onApplyWeights: (sortedSchedules: Schedule[], weights: WeightsState) => void;
  onClose: () => void;
}

interface DailySession {
  startTime: number;
  endTime: number;
}

const extractScheduleTimes = (courses: Course[]) => {
  let startTime = 2359;
  let endTime = 0;
  let totalGPA = 0;
  let gpaCount = 0;

  const sessionsByDay: { [key: number]: DailySession[] } = {};

  courses.forEach((course) => {
    course.Sessions.forEach((session) => {
      if (!session.IsAsync && !session.IsTimeTBD) {
        const dayMap: { [key: string]: number } = {
          M: 0,
          T: 1,
          W: 2,
          R: 3,
          F: 4,
        };

        session.Days.split("").forEach((day) => {
          const dayIndex = dayMap[day];
          if (dayIndex !== undefined) {
            if (!sessionsByDay[dayIndex]) {
              sessionsByDay[dayIndex] = [];
            }
            sessionsByDay[dayIndex].push({
              startTime: session.StartTime,
              endTime: session.EndTime,
            });
          }
        });

        if (session.StartTime < startTime) startTime = session.StartTime;
        if (session.EndTime > endTime) endTime = session.EndTime;
      }
    });

    if (course.GPA) {
      totalGPA += course.GPA;
      gpaCount++;
    }
  });

  let totalGapTime = 0;
  let daysWithClasses = 0;

  Object.values(sessionsByDay).forEach((daySessions) => {
    if (daySessions.length > 0) {
      daysWithClasses++;
      daySessions.sort((a, b) => a.startTime - b.startTime);

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
  weights: initialWeights,
  onApplyWeights,
  onClose,
}) => {
  const [weights, setWeights] = useState<WeightsState>(initialWeights);

  // When initialWeights changes, update local state
  useEffect(() => {
    setWeights(initialWeights);
  }, [initialWeights]);

  const timeToMinutes = (timeStr: string): number => {
    const [hours, minutes] = timeStr.split(":").map(Number);
    return hours * 60 + Math.round(minutes / 10) * 10;
  };

  const minutesToTime = (totalMinutes: number): string => {
    const hours = Math.floor(totalMinutes / 60);
    const minutes = Math.round((totalMinutes % 60) / 10) * 10;
    return `${hours.toString().padStart(2, "0")}:${minutes
      .toString()
      .padStart(2, "0")}`;
  };

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
    // Calculate difference as before
    const difference = Math.abs(actual - ideal);

    // Use exponential decay instead of linear
    // This will approach but never quite reach 0
    const score = Math.exp(-difference / (range / 2));

    // Scale to ensure score is between 0.1 and 1.0
    return 0.1 + 0.9 * score;
  };

  const calculateScheduleScore = (schedule: Schedule): number => {
    let totalScore = 0;
    let totalImportance = 0;

    const times = extractScheduleTimes(schedule.Courses);
    const startTimeMinutes = militaryToMinutes(times.startTime);
    const endTimeMinutes = militaryToMinutes(times.endTime);

    console.log(
      `\n[Schedule] Start: ${times.startTime} (${minutesToTime(
        startTimeMinutes
      )}), End: ${times.endTime} (${minutesToTime(endTimeMinutes)}), Gap: ${
        times.gapTime
      }min, GPA: ${times.averageGPA.toFixed(2)}`
    );

    if (weights["Start Time"].importance > 0) {
      const score = calculateLinearScore(
        startTimeMinutes,
        weights["Start Time"].idealValue!,
        120
      );
      const weightedScore = score * weights["Start Time"].importance;
      console.log(
        `[Start Time] ${minutesToTime(startTimeMinutes)} | Importance: ${
          weights["Start Time"].importance
        }, Ideal: ${minutesToTime(
          weights["Start Time"].idealValue!
        )}, Score: ${score.toFixed(3)} → ${weightedScore.toFixed(3)}`
      );
      totalScore += weightedScore;
      totalImportance += weights["Start Time"].importance;
    }

    if (weights["End Time"].importance > 0) {
      const score = calculateLinearScore(
        endTimeMinutes,
        weights["End Time"].idealValue!,
        120
      );
      const weightedScore = score * weights["End Time"].importance;
      console.log(
        `[End Time] ${minutesToTime(endTimeMinutes)} | Importance: ${
          weights["End Time"].importance
        }, Ideal: ${minutesToTime(
          weights["End Time"].idealValue!
        )}, Score: ${score.toFixed(3)} → ${weightedScore.toFixed(3)}`
      );
      totalScore += weightedScore;
      totalImportance += weights["End Time"].importance;
    }

    if (weights["Gap Time"].importance > 0) {
      const score = calculateLinearScore(
        times.gapTime,
        weights["Gap Time"].idealValue!,
        60
      );
      const weightedScore = score * weights["Gap Time"].importance;
      console.log(
        `[Gap Time] ${times.gapTime} mins | Importance: ${
          weights["Gap Time"].importance
        }, Ideal: ${weights["Gap Time"].idealValue}min, Score: ${score.toFixed(
          3
        )} → ${weightedScore.toFixed(3)}`
      );
      totalScore += weightedScore;
      totalImportance += weights["Gap Time"].importance;
    }

    if (weights["GPA"].importance > 0 && times.averageGPA > 0) {
      const score = times.averageGPA / 4.0;
      const weightedScore = score * weights["GPA"].importance;
      console.log(
        `[GPA] ${times.averageGPA} | Importance: ${
          weights["GPA"].importance
        }, Value: ${times.averageGPA.toFixed(2)}, Score: ${score.toFixed(
          3
        )} → ${weightedScore.toFixed(3)}`
      );
      totalScore += weightedScore;
      totalImportance += weights["GPA"].importance;
    }

    const finalScore = totalImportance > 0 ? totalScore / totalImportance : 0;
    console.log(
      `[Final] Score: ${finalScore.toFixed(3)} = ${totalScore.toFixed(
        3
      )}/${totalImportance}\n`
    );

    return finalScore;
  };

  const handleApply = () => {
    const sortedSchedules = [...schedules]
      .map((schedule) => ({
        ...schedule,
        Score: calculateScheduleScore(schedule),
      }))
      .sort((a, b) => (b.Score || 0) - (a.Score || 0));

    onApplyWeights(sortedSchedules, weights);
    onClose();
  };

  const getWeightIcon = (category: string) => {
    switch (category) {
      case "Start Time":
        return <Sun className="w-5 h-5" />;
      case "End Time":
        return <Moon className="w-5 h-5" />;
      case "Gap Time":
        return <ArrowDownUp className="w-5 h-5" />;
      case "GPA":
        return <GraduationCap className="w-5 h-5" />;
      default:
        return <Clock className="w-5 h-5" />;
    }
  };

  return (
    <div className={styles.container}>
      <p className={styles.description}>
        Adjust importance (0-10) and ideal values for each factor to customize
        your schedule preferences.
      </p>

      <div className={styles.weightsList}>
        {Object.entries(weights).map(([category, config]) => (
          <div key={category} className={styles.weightCard}>
            <div className={styles.cardHeader}>
              <div className={styles.cardIcon}>{getWeightIcon(category)}</div>
              <h3 className={styles.cardTitle}>{category}</h3>
            </div>

            <div className={styles.controls}>
              <div className={styles.controlGroup}>
                <label className={styles.label}>Importance:</label>
                <input
                  type="number"
                  min="0"
                  max="10"
                  step="0.5"
                  value={config.importance}
                  onChange={(e) =>
                    handleWeightChange(category, "importance", e.target.value)
                  }
                  className={styles.input}
                />
              </div>

              {config.idealValue !== undefined && (
                <div className={styles.controlGroup}>
                  <label className={styles.label}>
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
                      className={styles.input}
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
                      className={styles.timeInput}
                    />
                  )}
                </div>
              )}
            </div>
          </div>
        ))}
      </div>

      <button onClick={handleApply} className={styles.applyButton}>
        <Check size={16} />
        Apply Weights
      </button>
    </div>
  );
};

export default WeightsPopup;
