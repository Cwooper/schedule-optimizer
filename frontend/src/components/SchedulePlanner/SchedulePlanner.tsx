import React, { useState, useRef } from "react";
import type {
  Course,
  Schedule,
  ScheduleRequest,
  ScheduleResponse,
  WeightsState,
} from "../../types/types";
import CourseSelector from "../CourseSelector/CourseSelector";
import QuarterSelector from "../QuarterSelector/QuarterSelector";
import SchedulePreview from "../SchedulePreview/SchedulePreview";
import CourseList from "../CourseList/CourseList";
import Popup from "../Popup/Popup";
import WeightsPopup from "./WeightsPopup";
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
  customWeights: WeightsState | null;
  weights: WeightsState;
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
    customWeights: null,
    weights: {
      "Start Time": { importance: 1, idealValue: 600 },
      "End Time": { importance: 1, idealValue: 780 },
      "Gap Time": { importance: 2, idealValue: 0 },
      GPA: { importance: 2 },
    },
  });
  const [isWeightsPopupOpen, setIsWeightsPopupOpen] = useState(false);
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

      // If we have custom weights, apply them to the new schedules
      let sortedSchedules = response.Schedules || [];
      if (scheduleData.weights) {
        sortedSchedules = applyCustomWeights(
          sortedSchedules,
          scheduleData.weights
        );
      }

      setScheduleData((prev) => ({
        ...prev,
        schedules: sortedSchedules,
        totalSchedules: sortedSchedules.length,
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

  const applyCustomWeights = (
    schedules: Schedule[],
    weights: WeightsState
  ): Schedule[] => {
    const extractTimes = (courses: Course[]) => {
      let startTime = 2359;
      let endTime = 0;
      let totalGapTime = 0;
      let totalGPA = 0;
      let gpaCount = 0;

      courses.forEach((course) => {
        course.Sessions.forEach((session) => {
          if (!session.IsAsync && !session.IsTimeTBD) {
            if (session.StartTime < startTime) startTime = session.StartTime;
            if (session.EndTime > endTime) endTime = session.EndTime;
          }
        });
        if (course.GPA) {
          totalGPA += course.GPA;
          gpaCount++;
        }
      });

      totalGapTime = endTime - startTime - courses.length * 50;

      return {
        startTime,
        endTime,
        gapTime: Math.max(0, totalGapTime),
        averageGPA: gpaCount > 0 ? totalGPA / gpaCount : 0,
      };
    };

    const calculateScore = (schedule: Schedule): number => {
      let totalScore = 0;
      let totalImportance = 0;

      const times = extractTimes(schedule.Courses);

      if (weights["Start Time"].importance > 0) {
        const score =
          1 -
          Math.abs(times.startTime - weights["Start Time"].idealValue!) / 300;
        totalScore += Math.max(0, score) * weights["Start Time"].importance;
        totalImportance += weights["Start Time"].importance;
      }

      if (weights["End Time"].importance > 0) {
        const score =
          1 - Math.abs(times.endTime - weights["End Time"].idealValue!) / 300;
        totalScore += Math.max(0, score) * weights["End Time"].importance;
        totalImportance += weights["End Time"].importance;
      }

      if (weights["Gap Time"].importance > 0) {
        const score =
          1 - Math.abs(times.gapTime - weights["Gap Time"].idealValue!) / 60;
        totalScore += Math.max(0, score) * weights["Gap Time"].importance;
        totalImportance += weights["Gap Time"].importance;
      }

      if (weights["GPA"].importance > 0 && times.averageGPA > 0) {
        const score = times.averageGPA / 4.0;
        totalScore += score * weights["GPA"].importance;
        totalImportance += weights["GPA"].importance;
      }

      return totalImportance > 0 ? totalScore / totalImportance : 0;
    };

    return [...schedules]
      .map((schedule) => ({
        ...schedule,
        Score: calculateScore(schedule),
      }))
      .sort((a, b) => (b.Score || 0) - (a.Score || 0));
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

            <button
              onClick={() => setIsWeightsPopupOpen(true)}
              className={styles.actionButton}
              disabled={scheduleData.schedules.length === 0}
            >
              Weights
            </button>

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

        <Popup
          isOpen={isWeightsPopupOpen}
          onClose={() => setIsWeightsPopupOpen(false)}
          title="Schedule Weights"
          width="600px"
        >
          <WeightsPopup
            schedules={scheduleData.schedules}
            weights={scheduleData.weights}
            onApplyWeights={(sortedSchedules, weights) => {
              setScheduleData((prev) => ({
                ...prev,
                schedules: sortedSchedules,
                currentScheduleIndex: 0,
                totalSchedules: sortedSchedules.length,
                weights: weights,
              }));
              setIsWeightsPopupOpen(false);
            }}
            onClose={() => setIsWeightsPopupOpen(false)}
          />
        </Popup>

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
