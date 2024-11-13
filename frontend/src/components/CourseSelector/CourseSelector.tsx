import React, { useState, useEffect } from "react";
import styles from "./CourseSelector.module.css";
import SubmitButton from "./SubmitButton";

interface Course {
  id: number;
  subject: string;
  code: string;
  force: boolean;
}

interface CourseSelectorProps {
  courses: Course[];
  onAddCourse: (subject: string, code: string) => void;
  onRemoveCourse: (id: number) => void;
  onToggleForce: (id: number) => void;
  minCredits: string;
  maxCredits: string;
  onCreditUpdate: (field: "minCredits" | "maxCredits", value: string) => void;
  onSubmitSchedule: () => Promise<void>;
}

const LOCAL_STORAGE_KEY = "wwu-schedule-courses";

async function fetchSubjects(): Promise<string[]> {
  try {
    const response = await fetch("/schedule-optimizer/subjects");
    if (!response.ok) {
      throw new Error("Failed to fetch subjects");
    }
    const text = await response.text();
    return text
      .split("\n")
      .slice(1)
      .filter((line: string) => line.trim())
      .map((line: string) => line.trim())
      .sort();
  } catch (error) {
    console.error("Error fetching subjects:", error);
    throw error;
  }
}

const CourseSelector: React.FC<CourseSelectorProps> = ({
  courses,
  onAddCourse,
  onRemoveCourse,
  onToggleForce,
  minCredits,
  maxCredits,
  onCreditUpdate,
  onSubmitSchedule,
}) => {
  const [selectedSubject, setSelectedSubject] = useState("");
  const [courseCode, setCourseCode] = useState("");
  const [error, setError] = useState("");
  const [subjects, setSubjects] = useState<string[]>([
    "ACCT",
    "ANTH",
    "ART",
    "BIOL",
    "CHEM",
    "CSCI",
    "MATH",
    "PHYS",
  ]);
  const [hasLoadedFromStorage, setHasLoadedFromStorage] = useState(false);

  // Load courses from localStorage on initial mount
  useEffect(() => {
    if (hasLoadedFromStorage || courses.length > 0) return;

    const savedCourses = localStorage.getItem(LOCAL_STORAGE_KEY);
    if (savedCourses) {
      try {
        const parsedCourses = JSON.parse(savedCourses) as Course[];

        // Clear existing courses
        courses.forEach((course) => onRemoveCourse(course.id));

        // Add each saved course with its original ID
        parsedCourses.forEach((course) => {
          // First add the course
          onAddCourse(course.subject, course.code);

          // Then update its force status if needed
          if (course.force) {
            // Use requestAnimationFrame to ensure the course is added first
            requestAnimationFrame(() => {
              const addedCourse = courses.find(
                (c) => c.subject === course.subject && c.code === course.code
              );
              if (addedCourse) {
                onToggleForce(addedCourse.id);
              }
            });
          }
        });

        setHasLoadedFromStorage(true);
      } catch (e) {
        console.error("Error loading saved courses:", e);
      }
    }
    setHasLoadedFromStorage(true);
  }, [courses, hasLoadedFromStorage]);

  // Save courses to localStorage whenever they change
  useEffect(() => {
    if (hasLoadedFromStorage) {
      localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(courses));
    }
  }, [courses, hasLoadedFromStorage]);

  useEffect(() => {
    if (!minCredits) {
      onCreditUpdate("minCredits", "3");
    }
    if (!maxCredits) {
      onCreditUpdate("maxCredits", "3");
    }
  }, []);

  // Load subjects from the server
  useEffect(() => {
    const loadSubjects = async () => {
      try {
        const subjectList = await fetchSubjects();
        setSubjects(subjectList);
      } catch (error) {
        console.error("Error loading subjects:", error);
        setError("Error loading course subjects. Using default subject list.");
      }
    };

    loadSubjects();
  }, []);

  const validateCourseCode = (code: string) => {
    const regex = /^\d{3}[A-Z]?$/;
    return regex.test(code);
  };

  const handleAddCourse = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedSubject) {
      setError("Please select a subject");
      return;
    }
    if (!validateCourseCode(courseCode)) {
      setError("Course code must be 3 digits optionally followed by a letter");
      return;
    }

    if (courses.length >= 13) {
      setError("Maximum of 13 courses allowed");
      return;
    }

    // Check for duplicate course
    const isDuplicate = courses.some(
      (course) =>
        course.subject === selectedSubject && course.code === courseCode
    );

    if (isDuplicate) {
      setError(`${selectedSubject} ${courseCode} has already been added`);
      return;
    }

    onAddCourse(selectedSubject, courseCode);
    setCourseCode("");
    setError("");
  };

  return (
    <div>
      <form onSubmit={handleAddCourse} className={styles.container}>
        <select
          value={selectedSubject}
          onChange={(e) => setSelectedSubject(e.target.value)}
          className={`select ${styles.select}`}
        >
          <option value="">Select Subject</option>
          {subjects.map((subject) => (
            <option key={subject} value={subject}>
              {subject}
            </option>
          ))}
        </select>

        <input
          type="text"
          value={courseCode}
          onChange={(e) => setCourseCode(e.target.value.toUpperCase())}
          placeholder="Course Number"
          className={`input ${styles.input}`}
        />

        <button
          type="submit"
          className={`btn btn-primary ${styles.addButton}`}
          disabled={courses.length >= 13}
        >
          Add
        </button>
      </form>
      {error && <div className={styles.error}>{error}</div>}

      <div className={styles.courseListContainer}>
        <div className={styles.courseList}>
          {courses.map((course) => (
            <div
              key={course.id}
              className={`${styles.courseItem} ${
                course.force ? styles.forced : ""
              }`}
            >
              <span className={styles.courseText}>
                {course.subject} {course.code}
              </span>
              <div className={styles.courseActions}>
                <button
                  className={`${styles.forceButton} ${
                    course.force ? styles.forceButtonActive : ""
                  }`}
                  onClick={() => onToggleForce(course.id)}
                  title={course.force ? "Unforce course" : "Force course"}
                >
                  Force
                </button>
                <button
                  className={styles.removeButton}
                  onClick={() => onRemoveCourse(course.id)}
                  title="Remove course"
                >
                  Remove
                </button>
              </div>
            </div>
          ))}
        </div>
        {courses.length > 0 && (
          <div className={styles.submitButtonContainer}>
            <SubmitButton
              onSubmit={onSubmitSchedule}
              disabled={courses.length === 0}
            />
          </div>
        )}
      </div>
    </div>
  );
};

export default CourseSelector;
