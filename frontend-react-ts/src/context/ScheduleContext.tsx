import React from 'react';
import { Schedule } from '../types/scheduleTypes';

interface ScheduleContextType {
  schedules: Schedule[];
  currentScheduleIndex: number;
  setCurrentScheduleIndex: (index: number) => void;
}

export const ScheduleContext = React.createContext<ScheduleContextType>({
  schedules: [],
  currentScheduleIndex: 0,
  setCurrentScheduleIndex: () => {}
});