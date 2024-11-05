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
  CRN: string;
  Credits: number;
  Sessions: Session[];
  GPA?: number;
  Prerequisites?: string;
  Attributes?: string;
  AdditionalFees?: string;
  Restrictions?: string;
  AvailableSeats: number;
  Capacity: number;
  Enrolled: number;
}

export interface Session {
  Days: string;
  StartTime: number;
  EndTime: number;
  Instructor: string;
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
