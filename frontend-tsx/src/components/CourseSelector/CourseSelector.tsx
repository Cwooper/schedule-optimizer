import React, { useState } from 'react';
import { ChevronDown, X, Plus } from 'lucide-react';

interface Course {
  id: number;
  subject: string;
  code: string;
  force: boolean;
}

interface CourseSelectorProps {
  courses: Course[];
  onAddCourse: (subject: string, code: string) => void;
  onRemoveCourse: (id: number) => void;
  onToggleForce: (id: number) => void;
}

const CourseSelector: React.FC<CourseSelectorProps> = ({
  courses,
  onAddCourse,
  onRemoveCourse,
  onToggleForce,
}) => {
  const [selectedSubject, setSelectedSubject] = useState('');
  const [courseCode, setCourseCode] = useState('');
  const [error, setError] = useState('');

  // TODO: This would come from subjects.txt
  const subjects = ['CSCI', 'MATH', 'PHYS', 'CHEM'];

  const validateCourseCode = (code: string) => {
    const regex = /^\d{3}[A-Z]?$/;
    return regex.test(code);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedSubject) {
      setError('Please select a subject');
      return;
    }
    if (!validateCourseCode(courseCode)) {
      setError('Course code must be 3 digits optionally followed by a letter');
      return;
    }
    
    onAddCourse(selectedSubject, courseCode);
    setCourseCode('');
    setError('');
  };

  return (
    <div className="p-4">
      <h3 className="text-lg font-semibold mb-4">Course Selection</h3>
      
      <form onSubmit={handleSubmit} className="flex gap-4 mb-4">
        <div className="relative">
          <select
            value={selectedSubject}
            onChange={(e) => setSelectedSubject(e.target.value)}
            className="pl-3 pr-8 py-2 border rounded-md appearance-none bg-white"
          >
            <option value="">Select Subject</option>
            {subjects.map(subject => (
              <option key={subject} value={subject}>{subject}</option>
            ))}
          </select>
          <ChevronDown className="absolute right-2 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-500" />
        </div>

        <input
          type="text"
          value={courseCode}
          onChange={(e) => setCourseCode(e.target.value.toUpperCase())}
          placeholder="Course Code (e.g. 301A)"
          className="pl-3 pr-3 py-2 border rounded-md"
        />

        <button
          type="submit"
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          <Plus className="w-4 h-4" />
          Add
        </button>
      </form>

      {error && (
        <p className="text-red-500 text-sm mb-4">{error}</p>
      )}

      <div className="space-y-2">
        {courses.map((course) => (
          <div key={course.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-md">
            <span className="font-medium">{course.subject} {course.code}</span>
            <div className="flex items-center gap-2">
              <button
                onClick={() => onToggleForce(course.id)}
                className={`px-3 py-1 rounded-md text-sm ${
                  course.force 
                    ? 'bg-indigo-600 text-white' 
                    : 'bg-gray-200 text-gray-700'
                }`}
              >
                Force
              </button>
              <button
                onClick={() => onRemoveCourse(course.id)}
                className="p-1 text-gray-500 hover:text-red-500"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default CourseSelector;