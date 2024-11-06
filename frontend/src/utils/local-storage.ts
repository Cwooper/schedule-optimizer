export interface StoredCourse {
  id: number;
  subject: string;
  code: string;
  force: boolean;
}

const STORAGE_KEY = "wwu-schedule-courses";

export const saveCourseState = (courses: StoredCourse[]): void => {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(courses));
  } catch (error) {
    console.error("Error saving courses to local storage:", error);
  }
};

export const loadCourseState = (): StoredCourse[] => {
  try {
    const storedCourses = localStorage.getItem(STORAGE_KEY);
    if (!storedCourses) {
      return [];
    }
    return JSON.parse(storedCourses);
  } catch (error) {
    console.error("Error loading courses from local storage:", error);
    return [];
  }
};
