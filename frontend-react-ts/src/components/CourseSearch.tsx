import React, { useState } from 'react';

export const CourseSearch: React.FC = () => {
  const [searchTerm, setSearchTerm] = useState('');
  const [searchResults, setSearchResults] = useState<any[]>([]);

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    // Here you would typically make an API call to search for courses
    // This is a placeholder implementation
    const results = await fetch(`/api/search?term=${searchTerm}`)
      .then(res => res.json())
      .catch(err => {
        console.error('Error searching courses:', err);
        return [];
      });
    setSearchResults(results);
  };

  return (
    <div>
      <form onSubmit={handleSearch}>
        <input
          type="text"
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          placeholder="Search for courses"
        />
        <button type="submit">Search</button>
      </form>
      <ul>
        {searchResults.map((course, index) => (
          <li key={index}>{course.name} - {course.description}</li>
        ))}
      </ul>
    </div>
  );
};