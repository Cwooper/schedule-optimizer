// src/utils/schedule-utils.ts

/**
 * Convert quarter code to readable name
 */
export const getQuarterName = (quarterCode: string): string => {
  const quarterMap: { [key: string]: string } = {
    "10": "Winter",
    "20": "Spring",
    "30": "Summer",
    "40": "Fall",
  };
  return quarterMap[quarterCode] || "Unknown";
};

/**
 * Generate schedule filename based on quarter and year
 */
export const generateScheduleFilename = (
  quarter: string,
  year: string
): string => {
  const quarterName = getQuarterName(quarter);
  return `${quarterName}${year}-Schedule.png`;
};
