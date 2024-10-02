import React, { useContext } from 'react';
import { ScheduleContext } from '../context/ScheduleContext';

export const AsyncCoursesDisplay: React.FC = () => {
  const { schedules, currentScheduleIndex } = useContext(ScheduleContext);

  if (schedules.length === 0) {
    return null;
  }

  const currentSchedule = schedules[currentScheduleIndex];
  const asyncCourses = currentSchedule.courses.filter(course => course.isAsync);

  if (asyncCourses.length === 0) {
    return null;
  }

  return (
    <div>
      <h3>Asynchronous Courses</h3>
      <ul>
        {asyncCourses.map((course) => (
          <li key={course.id}>
            {course.name} - Section {course.section}
          </li>
        ))}
      </ul>
    </div>
  );
};