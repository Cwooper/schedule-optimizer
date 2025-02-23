// src/types/types.ts

export interface WeightConfig {
  importance: number;
  idealValue?: number;
}

export interface WeightsState {
  [key: string]: WeightConfig;
}

export interface Course {
  Subject: string;
  Title: string;
  Credits: number;
  CRN: string;
  Instructor: string;
  Sessions: Session[];
  GPA?: number;
  Capacity: number;
  Enrolled: number;
  AvailableSeats: number;
}

export interface Session {
  Days: string;
  StartTime: number;
  EndTime: number;
  Location: string;
  IsAsync: boolean;
  IsTimeTBD: boolean;
}

export interface ScheduleRequest {
  Courses: string[];
  Forced: string[];
  Min: number;
  Max: number;
  Term: string;
  SearchTerm: string;
}

export interface ScheduleResponse {
  Schedules: Schedule[];
  Errors?: string[];
  Warnings?: string[];
  Asyncs?: Course[];
  Courses?: Course[];
}

export interface Schedule {
  Score: number;
  Courses: Course[];
  Weights: Weight[];
  StartTime?: number;
  EndTime?: number;
  GapTime?: number;
  AverageGPA?: number;
}

export interface Weight {
  Name: string;
  Value: number;
}

export interface ScheduleEvent {
  id: string;
  days: number[];
  start: string;
  end: string;
  color: string;
  title: string;
  body: string;
}
