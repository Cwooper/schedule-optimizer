import React, { useState } from "react";
import styles from "./SchedulePlanner.module.css";
import QuarterSelector from "../QuarterSelector/QuarterSelector";
import CourseSelector from "../CourseSelector/CourseSelector";
import SchedulePreview from "../SchedulePreview/SchedulePreview";
import { Schedule, Course as BackendCourse } from "../../types/types";
import { submitSchedule } from "../../services/schedule-service";

interface Course {
  id: number;
  subject: string;
  code: string;
  force: boolean;
}

interface ScheduleData {
  quarter: string;
  year: string;
  minCredits: string;
  maxCredits: string;
  courses: Course[];
  currentScheduleIndex: number;
  totalSchedules: number;
  schedules: Schedule[];
  warnings: string[];
  errors: string[];
  asyncCourses: BackendCourse[];
}

const SchedulePlanner: React.FC = () => {
  const [scheduleData, setScheduleData] = useState<ScheduleData>({
    quarter: "",
    year: "",
    minCredits: "3",
    maxCredits: "3",
    courses: [],
    currentScheduleIndex: 0,
    totalSchedules: 0,
    schedules: [],
    warnings: [],
    errors: [],
    asyncCourses: [],
  });

  const [error, setError] = useState<string | null>(null);

  const handleQuarterUpdate = (field: string, value: string) => {
    setScheduleData((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleAddCourse = (subject: string, code: string) => {
    setScheduleData((prev) => ({
      ...prev,
      courses: [
        ...prev.courses,
        {
          id: Date.now(),
          subject,
          code,
          force: false,
        },
      ],
    }));
  };

  const handleRemoveCourse = (id: number) => {
    setScheduleData((prev) => ({
      ...prev,
      courses: prev.courses.filter((course) => course.id !== id),
    }));
  };

  const handleToggleForce = (id: number) => {
    setScheduleData((prev) => ({
      ...prev,
      courses: prev.courses.map((course) =>
        course.id === id ? { ...course, force: !course.force } : course
      ),
    }));
  };

  const handleNavigateSchedule = (direction: "prev" | "next") => {
    setScheduleData((prev) => ({
      ...prev,
      currentScheduleIndex:
        direction === "next"
          ? Math.min(prev.currentScheduleIndex + 1, prev.totalSchedules - 1)
          : Math.max(prev.currentScheduleIndex - 1, 0),
    }));
  };

  const handleCreditUpdate = (
    field: "minCredits" | "maxCredits",
    value: string
  ) => {
    setScheduleData((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleScheduleSubmit = async () => {
    try {
      setError(null);
      const request = {
        Courses: scheduleData.courses.map(
          (course) => `${course.subject}${course.code}`
        ),
        Forced: scheduleData.courses
          .filter((course) => course.force)
          .map((course) => `${course.subject}${course.code}`),
        Min: parseInt(scheduleData.minCredits),
        Max: parseInt(scheduleData.maxCredits),
        Term: `${scheduleData.year}${scheduleData.quarter}`,
        SearchTerm: "",
      };

      const response = await submitSchedule(request);

      setScheduleData((prev) => ({
        ...prev,
        schedules: response.Schedules || [],
        totalSchedules: response.Schedules?.length || 0,
        currentScheduleIndex: 0,
        warnings: response.Warnings || [],
        errors: response.Errors || [],
        asyncCourses: response.Asyncs || [],
      }));

      // Display warnings if any
      if (response.Warnings?.length) {
        setError(response.Warnings.join(". "));
      }

      // Display errors if any
      if (response.Errors?.length) {
        setError(response.Errors.join(". "));
      }
    } catch (error) {
      console.error("Failed to submit schedule:", error);
      setError("Failed to generate schedules. Please try again.");
    }
  };

  return (
    <div className={styles.container}>
      <div className={styles.formControls}>
        <QuarterSelector
          quarter={scheduleData.quarter}
          year={scheduleData.year}
          minCredits={scheduleData.minCredits}
          maxCredits={scheduleData.maxCredits}
          onUpdate={handleQuarterUpdate}
        />
        <CourseSelector
          courses={scheduleData.courses}
          onAddCourse={handleAddCourse}
          onRemoveCourse={handleRemoveCourse}
          onToggleForce={handleToggleForce}
          minCredits={scheduleData.minCredits}
          maxCredits={scheduleData.maxCredits}
          onCreditUpdate={handleCreditUpdate}
          onSubmitSchedule={handleScheduleSubmit}
        />

        <div className={styles.scheduleGlance}>
          {error && (
            <div className="mx-4 mb-4 p-4 rounded-lg bg-red-100 text-red-700 border border-red-300">
              {error}
            </div>
          )}
          <div className={styles.scheduleActions}>
            <button
              onClick={() => handleNavigateSchedule("prev")}
              disabled={scheduleData.currentScheduleIndex === 0}
              className={styles.actionButton}
            >
              Prev
            </button>

            <button className={styles.actionButton}>Weights & Sort</button>

            <button
              onClick={() => handleNavigateSchedule("next")}
              disabled={
                scheduleData.currentScheduleIndex ===
                scheduleData.totalSchedules - 1
              }
              className={styles.actionButton}
            >
              Next
            </button>
          </div>
          <div className={styles.schedulePreview}>
            <SchedulePreview
              schedule={
                scheduleData.schedules[scheduleData.currentScheduleIndex]
              }
              warnings={scheduleData.warnings}
              errors={scheduleData.errors}
            />
          </div>
        </div>
      </div>
    </div>
  );
};

export default SchedulePlanner;
