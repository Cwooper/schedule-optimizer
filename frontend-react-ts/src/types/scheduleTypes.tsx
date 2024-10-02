export interface Course {
    id: string;
    name: string;
    section: string;
    isAsync?: boolean;
    startTime?: string;
    endTime?: string;
    days?: string[];
  }
  
  export interface Schedule {
    id: string;
    courses: Course[];
  }