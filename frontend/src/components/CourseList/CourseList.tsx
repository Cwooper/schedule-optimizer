import React, { useState } from "react";
import styles from "./CourseList.module.css";
import EventPopup from "../SchedulePreview/EventPopup";
import type { Course } from "../../types/types";

interface CourseListProps {
  courses: Course[];
  title?: string;
  emptyMessage?: string;
}

interface CourseItemProps {
  course: Course;
  onClick: () => void;
}

const CourseItem: React.FC<CourseItemProps> = ({ course, onClick }) => {
  return (
    <div className={styles.courseItem} onClick={onClick}>
      <div className={styles.courseMain}>
        <div className={styles.courseHeader}>
          <span className={styles.courseCode}>{course.Subject}</span>
          <span className={styles.seats}>
            {course.AvailableSeats} seats available
          </span>
        </div>
        <div className={styles.courseTitle}>{course.Title}</div>
      </div>
      <div className={styles.courseDetails}>
        <span className={styles.credits}>{course.Credits} credits</span>
        {course.Instructor || "Staff"}
      </div>
    </div>
  );
};

const CourseList: React.FC<CourseListProps> = ({
  courses,
  title = "Courses",
  emptyMessage = "No courses to display",
}) => {
  const [selectedCourse, setSelectedCourse] = useState<{
    course: Course;
    session: Course["Sessions"][0];
  } | null>(null);

  const handleCourseClick = (course: Course) => {
    // For async courses, we'll use the first session
    setSelectedCourse({
      course,
      session: course.Sessions[0],
    });
  };

  return (
    <div className={styles.container}>
      {title && <h3 className={styles.title}>{title}</h3>}
      {courses.length === 0 ? (
        <div className={styles.emptyMessage}>{emptyMessage}</div>
      ) : (
        <div className={styles.courseList}>
          {courses.map((course) => (
            <CourseItem
              key={course.CRN}
              course={course}
              onClick={() => handleCourseClick(course)}
            />
          ))}
        </div>
      )}

      {selectedCourse && (
        <EventPopup
          course={selectedCourse.course}
          session={selectedCourse.session}
          isOpen={selectedCourse !== null}
          onClose={() => setSelectedCourse(null)}
        />
      )}
    </div>
  );
};

export default CourseList;
