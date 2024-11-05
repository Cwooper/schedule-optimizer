// src/components/SchedulePlanner/SchedulePlanner.tsx
import React, { useState, useRef } from "react";
import styles from "./SchedulePlanner.module.css";
import QuarterSelector from "../QuarterSelector/QuarterSelector";
import CourseSelector from "../CourseSelector/CourseSelector";
import SchedulePreview from "../SchedulePreview/SchedulePreview";
import {
  Schedule,
  Course as BackendCourse,
  ScheduleRequest,
} from "../../types/types";
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

  // Keep track of the last submitted request
  const lastRequest = useRef<ScheduleRequest | null>(null);

  const createScheduleRequest = (data: ScheduleData): ScheduleRequest => ({
    Courses: data.courses.map((course) => `${course.subject} ${course.code}`),
    Forced: data.courses
      .filter((course) => course.force)
      .map((course) => `${course.subject} ${course.code}`),
    Min: parseInt(data.minCredits),
    Max: parseInt(data.maxCredits),
    Term: `${data.year}${data.quarter}`,
    SearchTerm: "",
  });

  const areRequestsEqual = (
    req1: ScheduleRequest,
    req2: ScheduleRequest
  ): boolean => {
    return (
      JSON.stringify(req1.Courses.sort()) ===
        JSON.stringify(req2.Courses.sort()) &&
      JSON.stringify(req1.Forced.sort()) ===
        JSON.stringify(req2.Forced.sort()) &&
      req1.Min === req2.Min &&
      req1.Max === req2.Max &&
      req1.Term === req2.Term &&
      req1.SearchTerm === req2.SearchTerm
    );
  };

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
      const newRequest = createScheduleRequest(scheduleData);

      // Check if this request is identical to the last one
      if (
        lastRequest.current &&
        areRequestsEqual(lastRequest.current, newRequest)
      ) {
        // Skip submission if nothing has changed
        return;
      }

      const response = await submitSchedule(newRequest);

      // Update the last request reference
      lastRequest.current = newRequest;

      setScheduleData((prev) => ({
        ...prev,
        schedules: response.Schedules || [],
        totalSchedules: response.Schedules?.length || 0,
        currentScheduleIndex: 0,
        warnings: response.Warnings?.length ? response.Warnings : [],
        errors: response.Errors?.length ? response.Errors : [],
        asyncCourses: response.Asyncs || [],
      }));
    } catch (error) {
      console.error("Failed to submit schedule:", error);
      setScheduleData((prev) => ({
        ...prev,
        errors: ["Failed to generate schedules. Please try again."],
      }));
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
              showMessages={true}
              asyncCourses={scheduleData.asyncCourses}
            />
          </div>
        </div>
      </div>
    </div>
  );
};

export default SchedulePlanner;
