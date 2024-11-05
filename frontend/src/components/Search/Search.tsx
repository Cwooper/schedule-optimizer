import { FC, FormEvent, useState } from "react";
import CourseList from "../CourseList/CourseList";
import styles from "./Search.module.css";
import type { Course } from "../../types/types";

interface SearchProps {
  quarter: string;
  year: string;
}

const Search: FC<SearchProps> = ({ quarter, year }) => {
  const [searchText, setSearchText] = useState("");
  const [searchResults, setSearchResults] = useState<Course[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");

  const handleSearch = async (e: FormEvent) => {
    e.preventDefault();
    if (!searchText.trim()) return;

    setIsLoading(true);
    setError("");

    try {
      const searchRequest = {
        Courses: [],
        Forced: [],
        Min: 0,
        Max: 0,
        Term: `${year}${quarter}`,
        SearchTerm: searchText,
      };

      const response = await fetch("/schedule-optimizer/", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(searchRequest),
      });

      const data = await response.json();

      if (data.Errors?.length > 0) {
        setError(data.Errors.join(", "));
        setSearchResults([]);
      } else if (data.Courses?.length > 0) {
        setSearchResults(data.Courses);
      } else {
        setSearchResults([]);
      }
    } catch (err) {
      setError("Failed to search courses. Please try again.");
      setSearchResults([]);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className={styles.searchContainer}>
      <form onSubmit={handleSearch} className={styles.searchForm}>
        <input
          type="text"
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          placeholder="Search for courses..."
          className={styles.searchInput}
          disabled={isLoading}
        />
        <button
          type="submit"
          className={styles.searchButton}
          disabled={isLoading || !quarter || !year}
        >
          {isLoading ? "Searching..." : "Search"}
        </button>
      </form>

      {error && <div className={styles.error}>{error}</div>}

      {searchResults.length > 0 && (
        <CourseList
          courses={searchResults}
          title="Search Results"
          emptyMessage="No courses found matching your search."
        />
      )}
    </div>
  );
};

export default Search;
