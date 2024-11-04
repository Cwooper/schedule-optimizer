import React, { useState } from 'react';
import styles from './SchedulePlanner.module.css';
import QuarterSelector from '../QuarterSelector/QuarterSelector';
import CourseSelector from '../CourseSelector/CourseSelector';
import SchedulePreview from '../SchedulePreview/SchedulePreview';

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
}

const SchedulePlanner: React.FC = () => {
  const [scheduleData, setScheduleData] = useState<ScheduleData>({
    quarter: '',
    year: '',
    minCredits: '',
    maxCredits: '',
    courses: [],
    currentScheduleIndex: 0,
    totalSchedules: 0
  });

  const handleQuarterUpdate = (field: string, value: string) => {
    setScheduleData(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const handleAddCourse = (subject: string, code: string) => {
    setScheduleData(prev => ({
      ...prev,
      courses: [
        ...prev.courses,
        {
          id: Date.now(),
          subject,
          code,
          force: false
        }
      ]
    }));
  };

  const handleRemoveCourse = (id: number) => {
    setScheduleData(prev => ({
      ...prev,
      courses: prev.courses.filter(course => course.id !== id)
    }));
  };

  const handleToggleForce = (id: number) => {
    setScheduleData(prev => ({
      ...prev,
      courses: prev.courses.map(course =>
        course.id === id 
          ? { ...course, force: !course.force }
          : course
      )
    }));
  };

  const handleNavigateSchedule = (direction: 'prev' | 'next') => {
    setScheduleData(prev => ({
      ...prev,
      currentScheduleIndex: direction === 'next'
        ? Math.min(prev.currentScheduleIndex + 1, prev.totalSchedules - 1)
        : Math.max(prev.currentScheduleIndex - 1, 0)
    }));
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
        />

        <div className={styles.scheduleGlance}>
          <div className={styles.scheduleActions}>
          <button
            onClick={() => handleNavigateSchedule('prev')}
            disabled={scheduleData.currentScheduleIndex === 0}
            className={styles.actionButton}
          >
            Prev
          </button>
          
          <button className={styles.actionButton}>
            Weights/Sort by
          </button>
          
          <button
            onClick={() => handleNavigateSchedule('next')}
            disabled={scheduleData.currentScheduleIndex === scheduleData.totalSchedules - 1}
            className={styles.actionButton}
          >
            Next
          </button>
        </div>
          <div className={styles.schedulePreview}>
            <SchedulePreview />
          </div>
        </div>
      </div>
    </div>
  );
};

export default SchedulePlanner;