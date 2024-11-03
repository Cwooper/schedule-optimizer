// src/App.tsx
import React, { useState } from 'react';
import './styles/variables.css';
import './styles/global.css';
import TopBar from './components/TopBar/TopBar';

interface Course {
  id: number;
  name: string;
  force: boolean;
  crn?: string;
}

interface ScheduleState {
  quarter: string;
  year: string;
  minCredits: string;
  maxCredits: string;
  courses: Course[];
  currentScheduleIndex: number;
  totalSchedules: number;
}

const App: React.FC = () => {
  // Initialize state with default values
  const [scheduleState, setScheduleState] = useState<ScheduleState>({
    quarter: '',
    year: '',
    minCredits: '',
    maxCredits: '',
    courses: [
      { id: 1, name: '', force: false },
      { id: 2, name: '', force: false }
    ],
    currentScheduleIndex: 0,
    totalSchedules: 0
  });

  const handleAddCourse = () => {
    setScheduleState(prev => ({
      ...prev,
      courses: [
        ...prev.courses,
        { 
          id: prev.courses.length + 1, 
          name: '', 
          force: false 
        }
      ]
    }));
  };

  const handleForceToggle = (courseId: number) => {
    setScheduleState(prev => ({
      ...prev,
      courses: prev.courses.map(course =>
        course.id === courseId 
          ? { ...course, force: !course.force }
          : course
      )
    }));
  };

  const handleInputChange = (field: keyof ScheduleState, value: string) => {
    setScheduleState(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const handleCourseChange = (courseId: number, value: string) => {
    setScheduleState(prev => ({
      ...prev,
      courses: prev.courses.map(course =>
        course.id === courseId
          ? { ...course, name: value }
          : course
      )
    }));
  };

  const handleNavigateSchedule = (direction: 'prev' | 'next') => {
    setScheduleState(prev => ({
      ...prev,
      currentScheduleIndex: direction === 'next'
        ? Math.min(prev.currentScheduleIndex + 1, prev.totalSchedules - 1)
        : Math.max(prev.currentScheduleIndex - 1, 0)
    }));
  };

  return (
    <div className="app">
      <TopBar />
      <main className="main-content">
        {/* Form Controls Section */}
        <div className="card">
          <div className="flex gap-md">
            <div className="form-group">
              <label className="form-label">Quarter:</label>
              <select 
                value={scheduleState.quarter}
                onChange={(e) => handleInputChange('quarter', e.target.value)}
              >
                <option value="">Select</option>
                <option value="Fall">Fall</option>
                <option value="Winter">Winter</option>
                <option value="Spring">Spring</option>
                <option value="Summer">Summer</option>
              </select>
            </div>

            <div className="form-group">
              <label className="form-label">Year:</label>
              <select 
                value={scheduleState.year}
                onChange={(e) => handleInputChange('year', e.target.value)}
              >
                <option value="">Select</option>
                {[2024, 2025, 2026, 2027].map(year => (
                  <option key={year} value={year}>{year}</option>
                ))}
              </select>
            </div>

            <div className="form-group">
              <label className="form-label">Min:</label>
              <select 
                value={scheduleState.minCredits}
                onChange={(e) => handleInputChange('minCredits', e.target.value)}
              >
                <option value="">Select</option>
                {Array.from({ length: 20 }, (_, i) => i + 1).map(num => (
                  <option key={num} value={num}>{num}</option>
                ))}
              </select>
            </div>

            <div className="form-group">
              <label className="form-label">Max:</label>
              <select 
                value={scheduleState.maxCredits}
                onChange={(e) => handleInputChange('maxCredits', e.target.value)}
              >
                <option value="">Select</option>
                {Array.from({ length: 20 }, (_, i) => i + 1).map(num => (
                  <option key={num} value={num}>{num}</option>
                ))}
              </select>
            </div>
          </div>
        </div>

        {/* Course Selection Section */}
        <div className="card mt-4">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">Subject/CRN</h2>
            <button onClick={handleAddCourse} className="button-primary">
              Add
            </button>
          </div>

          <div className="flex flex-col gap-md">
            {scheduleState.courses.map((course) => (
              <div key={course.id} className="flex items-center justify-between">
                <input
                  type="text"
                  value={course.name}
                  onChange={(e) => handleCourseChange(course.id, e.target.value)}
                  placeholder="Enter course or CRN"
                  className="form-input"
                />
                <button
                  onClick={() => handleForceToggle(course.id)}
                  className={`button-secondary ${course.force ? 'button-forced' : ''}`}
                >
                  Force
                </button>
              </div>
            ))}
          </div>
        </div>

        {/* Navigation and Schedule Preview */}
        <div className="flex justify-between gap-md mt-4">
          <button 
            className="button-secondary"
            onClick={() => handleNavigateSchedule('prev')}
            disabled={scheduleState.currentScheduleIndex === 0}
          >
            Prev
          </button>
          
          <button className="button-secondary">
            Weights/Sort by
          </button>
          
          <button 
            className="button-secondary"
            onClick={() => handleNavigateSchedule('next')}
            disabled={scheduleState.currentScheduleIndex === scheduleState.totalSchedules - 1}
          >
            Next
          </button>
        </div>

        {/* Schedule Preview */}
        <div className="card mt-4">
          <h2 className="text-xl font-semibold mb-4">Schedule at a glance...</h2>
          <div className="schedule-preview">
            {/* Schedule preview content will go here */}
          </div>
        </div>
      </main>
    </div>
  );
};

export default App;