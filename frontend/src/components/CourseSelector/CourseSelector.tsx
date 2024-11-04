import React, { useState, useEffect } from "react";
import styles from "./CourseSelector.module.css";
import SubmitButton from './SubmitButton';

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
  onSubmitSchedule: () => Promise<void>;  // Add this prop for schedule submission
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

  useEffect(() => {
    if (!minCredits) {
      onCreditUpdate("minCredits", "3");
    }
    if (!maxCredits) {
      onCreditUpdate("maxCredits", "3");
    }
  }, []);

  const subjects = [
    "ACCT", "ANTH", "ART", "BIOL", "CHEM", "CSCI", "MATH", "PHYS",
  ];

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

    if (courses.length >= 9) {
      setError("Maximum of 9 courses allowed");
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
          disabled={courses.length >= 9}
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
              className={`${styles.courseItem} ${course.force ? styles.forced : ""}`}
            >
              <span className={styles.courseText}>
                {course.subject} {course.code}
              </span>
              <div className={styles.courseActions}>
                <button
                  className={`${styles.forceButton} ${course.force ? styles.forceButtonActive : ""}`}
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