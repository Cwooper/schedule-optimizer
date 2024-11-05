import React, { useState, useRef } from "react";
import type {
  Course,
  Schedule,
  ScheduleRequest,
  ScheduleResponse,
} from "../../types/types";
import CourseSelector from "../CourseSelector/CourseSelector";
import QuarterSelector from "../QuarterSelector/QuarterSelector";
import SchedulePreview from "../SchedulePreview/SchedulePreview";
import CourseList from "../CourseList/CourseList";
import styles from "./SchedulePlanner.module.css";

interface CourseItem {
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
  courses: CourseItem[];
  currentScheduleIndex: number;
  totalSchedules: number;
  schedules: Schedule[];
  warnings: string[];
  errors: string[];
  asyncCourses: Course[];
  searchResults: Course[];
  isSearching: boolean;
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
    searchResults: [],
    isSearching: false,
  });

  const [searchText, setSearchText] = useState("");
  const lastRequest = useRef<ScheduleRequest | null>(null);

  const createScheduleRequest = (
    courses: CourseItem[],
    minCredits: string,
    maxCredits: string,
    year: string,
    quarter: string,
    searchTerm = ""
  ): ScheduleRequest => ({
    Courses: courses.map((course) => `${course.subject} ${course.code}`),
    Forced: courses
      .filter((course) => course.force)
      .map((course) => `${course.subject} ${course.code}`),
    Min: parseInt(minCredits),
    Max: parseInt(maxCredits),
    Term: `${year}${quarter}`,
    SearchTerm: searchTerm,
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

  const submitRequest = async (
    request: ScheduleRequest
  ): Promise<ScheduleResponse> => {
    const response = await fetch("/schedule-optimizer/", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return await response.json();
  };

  const handleScheduleSubmit = async () => {
    try {
      const newRequest = createScheduleRequest(
        scheduleData.courses,
        scheduleData.minCredits,
        scheduleData.maxCredits,
        scheduleData.year,
        scheduleData.quarter
      );

      if (
        lastRequest.current &&
        areRequestsEqual(lastRequest.current, newRequest)
      ) {
        return;
      }

      const response = await submitRequest(newRequest);
      lastRequest.current = newRequest;

      setScheduleData((prev) => ({
        ...prev,
        schedules: response.Schedules || [],
        totalSchedules: response.Schedules?.length || 0,
        currentScheduleIndex: 0,
        warnings: response.Warnings || [],
        errors: response.Errors || [],
        asyncCourses: response.Asyncs || [],
        searchResults: [],
      }));
    } catch (error) {
      console.error("Failed to submit schedule:", error);
      setScheduleData((prev) => ({
        ...prev,
        errors: ["Failed to generate schedules. Please try again."],
      }));
    }
  };

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!searchText.trim() || !scheduleData.quarter || !scheduleData.year)
      return;

    setScheduleData((prev) => ({ ...prev, isSearching: true }));

    try {
      const searchRequest = createScheduleRequest(
        [],
        "0",
        "0",
        scheduleData.year,
        scheduleData.quarter,
        searchText
      );

      const response = await submitRequest(searchRequest);

      setScheduleData((prev) => ({
        ...prev,
        searchResults: response.Courses || [],
        warnings: response.Warnings || [],
        errors: response.Errors || [],
        isSearching: false,
      }));
    } catch (error) {
      console.error("Search failed:", error);
      setScheduleData((prev) => ({
        ...prev,
        errors: ["Failed to search courses. Please try again."],
        searchResults: [],
        isSearching: false,
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
              Previous
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

        {/* Search Section */}
        <div className={styles.searchSection}>
          <div className={styles.searchHeader}>
            <h3 className={styles.searchTitle}>Course Search</h3>
            <p className={styles.searchDescription}>
              Search for courses by subject, title, or instructor. View detailed
              course information before adding to your schedule.
            </p>
          </div>
          <form onSubmit={handleSearch} className={styles.searchForm}>
            <input
              type="text"
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              placeholder="Search for courses..."
              className={styles.searchInput}
              disabled={
                scheduleData.isSearching ||
                !scheduleData.quarter ||
                !scheduleData.year
              }
            />
            <button
              type="submit"
              className={styles.searchButton}
              disabled={
                scheduleData.isSearching ||
                !scheduleData.quarter ||
                !scheduleData.year
              }
            >
              {scheduleData.isSearching ? "Searching..." : "Search"}
            </button>
          </form>

          {scheduleData.searchResults.length > 0 && (
            <div className={styles.searchResults}>
              <CourseList
                courses={scheduleData.searchResults}
                title="Search Results"
                emptyMessage="No courses found matching your search."
              />
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default SchedulePlanner;
