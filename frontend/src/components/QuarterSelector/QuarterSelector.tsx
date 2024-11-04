import React, { useEffect } from "react";
import styles from "./QuarterSelector.module.css";

interface QuarterSelectorProps {
  quarter: string;
  year: string;
  minCredits: string;
  maxCredits: string;
  onUpdate: (field: string, value: string) => void;
}

interface Quarter {
  name: string;
  year: number;
  registrationDate: Date; // Added registration date
}

const QuarterSelector: React.FC<QuarterSelectorProps> = ({
  quarter,
  year,
  minCredits,
  maxCredits,
  onUpdate,
}) => {
  const getCurrentQuarter = (): { quarter: string; year: string } => {
    const now = new Date();
    const quarters: Quarter[] = [];

    // Generate quarters for the next 3 years
    for (let yearOffset = 0; yearOffset < 3; yearOffset++) {
      const baseYear = now.getFullYear() + yearOffset;

      quarters.push(
        {
          name: "40",
          year: baseYear,
          registrationDate: new Date(baseYear, 6, 1), // July 1st - registration
        },
        {
          name: "10",
          year: baseYear + 1,
          registrationDate: new Date(baseYear, 9, 1), // Oct 1st - registration
        },
        {
          name: "20",
          year: baseYear + 1,
          registrationDate: new Date(baseYear + 1, 1, 1), // Feb 1st - registration
        },
        {
          name: "30",
          year: baseYear + 1,
          registrationDate: new Date(baseYear + 1, 4, 1), // May 1st - registration
        }
      );
    }

    // Sort quarters by registration date
    quarters.sort(
      (a, b) => a.registrationDate.getTime() - b.registrationDate.getTime()
    );

    // Find the current quarter based on registration dates
    // If we're between Oct 1 and Feb 1, we want Winter
    // If we're between Feb 1 and May 1, we want Spring
    // If we're between May 1 and July 1, we want Summer
    // If we're between July 1 and Oct 1, we want Fall
    const currentQuarter =
      quarters.find(
        (q) =>
          q.registrationDate <= now &&
          (quarters[quarters.indexOf(q) + 1]?.registrationDate > now ||
            !quarters[quarters.indexOf(q) + 1])
      ) || quarters[0];

    return {
      quarter: currentQuarter.name,
      year: currentQuarter.year.toString(),
    };
  };

  // Set initial values when component mounts
  useEffect(() => {
    if (!quarter || !year) {
      const { quarter: currentQuarter, year: currentYear } =
        getCurrentQuarter();
      onUpdate("quarter", currentQuarter);
      onUpdate("year", currentYear);
    }
  }, []);

  // Generate available years (current year and next year)
  const currentYear = new Date().getFullYear();
  const availableYears = Array.from({ length: 2 }, (_, i) => currentYear + i);

  return (
    <div className={styles.container}>
      <div className={styles.formGroup}>
        <label className={styles.label}>Quarter:</label>
        <select
          value={quarter}
          onChange={(e) => onUpdate("quarter", e.target.value)}
          className={styles.select}
        >
          <option value="">Select Quarter</option>
          <option value="40">Fall</option>
          <option value="10">Winter</option>
          <option value="20">Spring</option>
          <option value="30">Summer</option>
        </select>
      </div>

      <div className={styles.formGroup}>
        <label className={styles.label}>Year:</label>
        <select
          value={year}
          onChange={(e) => onUpdate("year", e.target.value)}
          className={styles.select}
        >
          <option value="">Select Year</option>
          {availableYears.map((year) => (
            <option key={year} value={year}>
              {year}
            </option>
          ))}
        </select>
      </div>

      <div className={styles.formGroup}>
        <label className={styles.label}>Minimum Courses:</label>
        <select
          value={minCredits}
          onChange={(e) => onUpdate("minCredits", e.target.value)}
          className={styles.select}
        >
          <option value="">Min Courses</option>
          {Array.from({ length: 8 }, (_, i) => i + 1).map((num) => (
            <option key={num} value={num}>
              {num}
            </option>
          ))}
        </select>
      </div>

      <div className={styles.formGroup}>
        <label className={styles.label}>Maximum Courses:</label>
        <select
          value={maxCredits}
          onChange={(e) => onUpdate("maxCredits", e.target.value)}
          className={styles.select}
        >
          <option value="">Max Courses</option>
          {Array.from({ length: 8 }, (_, i) => i + 1).map((num) => (
            <option key={num} value={num}>
              {num}
            </option>
          ))}
        </select>
      </div>
    </div>
  );
};

export default QuarterSelector;
