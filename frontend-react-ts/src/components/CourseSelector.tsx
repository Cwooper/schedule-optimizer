import React, { useState, useEffect } from 'react';

interface CourseSelectorProps {
  onAddCourse: (course: { id: string; name: string; section: string }) => void;
}

export const CourseSelector: React.FC<CourseSelectorProps> = ({ onAddCourse }) => {
  const [courses, setCourses] = useState<string[]>([]);
  const [selectedCourse, setSelectedCourse] = useState('');
  const [section, setSection] = useState('');

  useEffect(() => {
    // Fetch courses from your API
    fetch('subjects.txt')
      .then(response => response.text())
      .then(text => {
        const courseList = text.trim().split('\n');
        setCourses(courseList);
      })
      .catch(error => console.error('Error fetching courses:', error));
  }, []);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (selectedCourse && section) {
      onAddCourse({
        id: `${selectedCourse}-${section}`,
        name: selectedCourse,
        section: section
      });
      setSection('');
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <select
        value={selectedCourse}
        onChange={(e) => setSelectedCourse(e.target.value)}
        required
      >
        <option value="">Select a course</option>
        {courses.map((course, index) => (
          <option key={index} value={course}>
            {course}
          </option>
        ))}
      </select>
      <input
        type="text"
        value={section}
        onChange={(e) => setSection(e.target.value)}
        placeholder="Section number"
        required
      />
      <button type="submit">Add Course</button>
    </form>
  );
};