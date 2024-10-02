import React, { useState } from 'react';
import { CourseSelector } from './components/CourseSelector';
import { CourseList } from './components/CourseList';
import { ScheduleDisplay } from './components/ScheduleDisplay';
import { AsyncCoursesDisplay } from './components/AsyncCoursesDisplay';
import { CourseSearch } from './components/CourseSearch';
import { ErrorBoundary } from './components/ErrorBoundary';
import { ScheduleContext } from './context/ScheduleContext';
import { Course, Schedule } from './types/scheduleTypes';

const App: React.FC = () => {
  const [courses, setCourses] = useState<Course[]>([]);
  const [schedules, setSchedules] = useState<Schedule[]>([]);
  const [currentScheduleIndex, setCurrentScheduleIndex] = useState(0);

  const addCourse = (course: Course) => {
    setCourses(prevCourses => [...prevCourses, course]);
  };

  const removeCourse = (courseId: string) => {
    setCourses(prevCourses => prevCourses.filter(course => course.id !== courseId));
  };

  const generateSchedules = () => {
    // This is a placeholder for the actual schedule generation logic
    // You would typically call an API or use a complex algorithm here
    const newSchedules: Schedule[] = [
      {
        id: '1',
        courses: courses.map(course => ({
          ...course,
          startTime: '09:00',
          endTime: '10:00',
          days: ['Monday', 'Wednesday', 'Friday']
        }))
      },
      {
        id: '2',
        courses: courses.map(course => ({
          ...course,
          startTime: '11:00',
          endTime: '12:00',
          days: ['Tuesday', 'Thursday']
        }))
      }
    ];
    setSchedules(newSchedules);
    setCurrentScheduleIndex(0);
  };

  return (
    <ErrorBoundary>
      <ScheduleContext.Provider value={{ schedules, currentScheduleIndex, setCurrentScheduleIndex }}>
        <div className="App">
          <header>
            <h1>WWU Schedule Optimizer</h1>
          </header>
          <main>
            <section>
              <h2>Course Selection</h2>
              <CourseSelector onAddCourse={addCourse} />
              <CourseList courses={courses} onRemoveCourse={removeCourse} />
              <button onClick={generateSchedules}>Generate Schedules</button>
            </section>
            <section>
              <h2>Schedule</h2>
              <ScheduleDisplay />
              <AsyncCoursesDisplay />
            </section>
            <section>
              <h2>Course Search</h2>
              <CourseSearch />
            </section>
          </main>
        </div>
      </ScheduleContext.Provider>
    </ErrorBoundary>
  );
};

export default App;