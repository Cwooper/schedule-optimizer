// src/components/SchedulePlanner/SchedulePlanner.tsx
import React from 'react';
import styles from './SchedulePlanner.module.css';
import { ChevronDown } from 'lucide-react';

interface Course {
  id: number;
  name: string;
  force: boolean;
}

interface ScheduleData {
  quarter: string;
  year: string;
  minCredits: string;
  maxCredits: string;
  courses: Course[];
}

interface SchedulePlannerProps {
  scheduleData: ScheduleData;
  setScheduleData: React.Dispatch<React.SetStateAction<ScheduleData>>;
}

const SchedulePlanner: React.FC<SchedulePlannerProps> = ({ scheduleData, setScheduleData }) => {
  const addCourse = () => {
    setScheduleData(prev => ({
      ...prev,
      courses: [...prev.courses, { id: prev.courses.length + 1, name: `Course ${prev.courses.length + 1}`, force: false }]
    }));
  };

  const toggleForce = (courseId: number) => {
    setScheduleData(prev => ({
      ...prev,
      courses: prev.courses.map(course =>
        course.id === courseId ? { ...course, force: !course.force } : course
      )
    }));
  };

  return (
    <div className={styles.container}>
      <div className={styles.formControls}>
        <div className={styles.dropdowns}>
          <div className={styles.dropdownGroup}>
            <label>Quarter:</label>
            <div className={styles.selectWrapper}>
              <select
                value={scheduleData.quarter}
                onChange={(e) => setScheduleData(prev => ({ ...prev, quarter: e.target.value }))}
              >
                <option value="">Select Quarter</option>
                <option value="fall">Fall</option>
                <option value="winter">Winter</option>
                <option value="spring">Spring</option>
                <option value="summer">Summer</option>
              </select>
              <ChevronDown className={styles.dropdownIcon} />
            </div>
          </div>

          <div className={styles.dropdownGroup}>
            <label>Year:</label>
            <div className={styles.selectWrapper}>
              <select
                value={scheduleData.year}
                onChange={(e) => setScheduleData(prev => ({ ...prev, year: e.target.value }))}
              >
                <option value="">Select Year</option>
                {[2024, 2025, 2026, 2027, 2028].map(year => (
                  <option key={year} value={year}>{year}</option>
                ))}
              </select>
              <ChevronDown className={styles.dropdownIcon} />
            </div>
          </div>

          <div className={styles.dropdownGroup}>
            <label>Min:</label>
            <div className={styles.selectWrapper}>
              <select
                value={scheduleData.minCredits}
                onChange={(e) => setScheduleData(prev => ({ ...prev, minCredits: e.target.value }))}
              >
                <option value="">Min Credits</option>
                {Array.from({ length: 20 }, (_, i) => i + 1).map(num => (
                  <option key={num} value={num}>{num}</option>
                ))}
              </select>
              <ChevronDown className={styles.dropdownIcon} />
            </div>
          </div>

          <div className={styles.dropdownGroup}>
            <label>Max:</label>
            <div className={styles.selectWrapper}>
              <select
                value={scheduleData.maxCredits}
                onChange={(e) => setScheduleData(prev => ({ ...prev, maxCredits: e.target.value }))}
              >
                <option value="">Max Credits</option>
                {Array.from({ length: 20 }, (_, i) => i + 1).map(num => (
                  <option key={num} value={num}>{num}</option>
                ))}
              </select>
              <ChevronDown className={styles.dropdownIcon} />
            </div>
          </div>
        </div>

        <div className={styles.courseSection}>
          <div className={styles.courseHeader}>
            <h3>Subject/CRN</h3>
            <button onClick={addCourse} className={styles.addButton}>Add</button>
          </div>
          
          <div className={styles.courseList}>
            {scheduleData.courses.map((course) => (
              <div key={course.id} className={styles.courseItem}>
                <span>{course.name}</span>
                <button
                  onClick={() => toggleForce(course.id)}
                  className={`${styles.forceButton} ${course.force ? styles.forced : ''}`}
                >
                  Force
                </button>
              </div>
            ))}
          </div>
        </div>

        <div className={styles.scheduleActions}>
          <button className={styles.actionButton}>Prev</button>
          <button className={styles.actionButton}>Weights/Sort by</button>
          <button className={styles.actionButton}>Next</button>
        </div>

        <div className={styles.scheduleGlance}>
          <h3>Schedule at a glance...</h3>
          <div className={styles.schedulePreview}>
            {/* Schedule preview content will go here */}
          </div>
        </div>
      </div>
    </div>
  );
};

export default SchedulePlanner;