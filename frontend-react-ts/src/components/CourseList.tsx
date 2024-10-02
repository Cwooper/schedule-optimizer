import React from 'react';
import { Course } from '../types/scheduleTypes';

interface CourseListProps {
  courses: Course[];
  onRemoveCourse: (id: string) => void;
}

export const CourseList: React.FC<CourseListProps> = ({ courses, onRemoveCourse }) => {
  return (
    <ul>
      {courses.map((course) => (
        <li key={course.id}>
          {course.name} {course.section}
          <button onClick={() => onRemoveCourse(course.id)}>Remove</button>
        </li>
      ))}
    </ul>
  );
};