import React, { useEffect } from "react";
import styles from "./QuarterSelector.module.css";

interface QuarterSelectorProps {
  quarter: string;
  year: string;
  minCredits: string;
  maxCredits: string;
  onUpdate: (field: string, value: string) => void;
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
    const month = now.getMonth(); // 0-11
    const year = now.getFullYear();
  
    // Determine quarter and its academic year based on current date
    if (month >= 9 || month < 1) { // Oct-Jan = Winter of next year
      return {
        quarter: "10",
        year: (month >= 9 ? year + 1 : year).toString()
      };
    } else if (month >= 1 && month < 4) { // Feb-Apr = Spring
      return {
        quarter: "20",
        year: year.toString()
      };
    } else if (month >= 4 && month < 6) { // May-Jun = Summer
      return {
        quarter: "30",
        year: year.toString()
      };
    } else { // Jul-Sep = Fall
      return {
        quarter: "40",
        year: year.toString()
      };
    }
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
