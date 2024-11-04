import React, { useState, useEffect } from "react";
import styles from "./CourseSelector.module.css";
import { X } from "lucide-react";

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
  onCreditUpdate: (field: 'minCredits' | 'maxCredits', value: string) => void;
}

const CourseSelector: React.FC<CourseSelectorProps> = ({
  courses,
  onAddCourse,
  onRemoveCourse,
  onToggleForce,
  minCredits,
  maxCredits,
  onCreditUpdate,
}) => {
  const [selectedSubject, setSelectedSubject] = useState("");
  const [courseCode, setCourseCode] = useState("");
  const [error, setError] = useState("");

  // Set default credits on component mount
  useEffect(() => {
    if (!minCredits) {
      onCreditUpdate('minCredits', '3');
    }
    if (!maxCredits) {
      onCreditUpdate('maxCredits', '3');
    }
  }, []);

  const subjects = [
    "ACCT",
    "ANTH",
    "ART",
    "BIOL",
    "CHEM",
    "CSCI",
    "MATH",
    "PHYS",
  ];

  const validateCourseCode = (code: string) => {
    const regex = /^\d{3}[A-Z]?$/;
    return regex.test(code);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedSubject) {
      setError("Please select a subject");
      return;
    }
    if (!validateCourseCode(courseCode)) {
      setError("Course code must be 3 digits optionally followed by a letter");
      return;
    }

    // Check if we've reached the maximum number of courses
    if (courses.length >= 9) {
      setError("Maximum of 9 courses allowed");
      return;
    }
    
    onAddCourse(selectedSubject, courseCode);
    setCourseCode("");
    setError("");
  };

  return (
    <>
      <form onSubmit={handleSubmit} className={styles.container}>
        <select
          value={selectedSubject}
          onChange={(e) => setSelectedSubject(e.target.value)}
          className={styles.select}
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
          className={styles.input}
        />

        <button 
          type="submit" 
          className={styles.addButton}
          disabled={courses.length >= 9}
        >
          Add
        </button>
      </form>

      {error && <p className={styles.error}>{error}</p>}

      <div className={styles.courseList}>
        {courses.map((course) => (
          <div key={course.id} className={styles.courseItem}>
            <span>
              {course.subject} {course.code}
            </span>
            <div className={styles.courseActions}>
              <button
                onClick={() => onToggleForce(course.id)}
                className={`${styles.forceButton} ${
                  course.force ? styles.forced : ""
                }`}
              >
                Force
              </button>
              <button
                onClick={() => onRemoveCourse(course.id)}
                className={styles.removeButton}
              >
                <X className="w-4 h-4" />
              </button>
            </div>
          </div>
        ))}
      </div>
    </>
  );
};

export default CourseSelector;