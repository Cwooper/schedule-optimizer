import React, { useContext } from 'react';
import { ScheduleContext } from '../context/ScheduleContext';

export const ScheduleDisplay: React.FC = () => {
  const { schedules, currentScheduleIndex, setCurrentScheduleIndex } = useContext(ScheduleContext);

  if (schedules.length === 0) {
    return <p>No schedules generated yet.</p>;
  }

  const currentSchedule = schedules[currentScheduleIndex];

  const nextSchedule = () => {
    if (currentScheduleIndex < schedules.length - 1) {
      setCurrentScheduleIndex(currentScheduleIndex + 1);
    }
  };

  const prevSchedule = () => {
    if (currentScheduleIndex > 0) {
      setCurrentScheduleIndex(currentScheduleIndex - 1);
    }
  };

  return (
    <div>
      <h2>Schedule {currentScheduleIndex + 1} of {schedules.length}</h2>
      <button onClick={prevSchedule} disabled={currentScheduleIndex === 0}>Previous</button>
      <button onClick={nextSchedule} disabled={currentScheduleIndex === schedules.length - 1}>Next</button>
      <table>
        <thead>
          <tr>
            <th>Time</th>
            <th>Monday</th>
            <th>Tuesday</th>
            <th>Wednesday</th>
            <th>Thursday</th>
            <th>Friday</th>
          </tr>
        </thead>
        <tbody>
          {/* Here you would generate rows for each time slot */}
          {/* This is a simplified example */}
          <tr>
            <td>8:00 AM</td>
            <td>{/* Course for Monday 8:00 AM */}</td>
            <td>{/* Course for Tuesday 8:00 AM */}</td>
            <td>{/* Course for Wednesday 8:00 AM */}</td>
            <td>{/* Course for Thursday 8:00 AM */}</td>
            <td>{/* Course for Friday 8:00 AM */}</td>
          </tr>
          {/* More rows... */}
        </tbody>
      </table>
    </div>
  );
};
