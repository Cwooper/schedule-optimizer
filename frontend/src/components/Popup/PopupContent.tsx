// src/components/Popup/PopupContent.tsx
import React from "react";
import styles from "./Popup.module.css";

export const AboutContent: React.FC = () => {
  return (
    <div className={styles.aboutContent}>
      <h2>About WWU Schedule Optimizer</h2>
      <p>
        The WWU Schedule Optimizer is a tool designed to help Western Washington
        University students plan their course schedules efficiently. It takes
        into account various factors such as course availability, time
        conflicts, and prerequisites to generate optimal schedule combinations.
      </p>

      <h3>Features</h3>
      <ul>
        <li>Course selection from multiple departments</li>
        <li>Automatic conflict detection</li>
        <li>Support for multiple schedule variations</li>
        <li>Course forcing capability</li>
        <li>Customizable course load</li>
        <li>Customizable weighting system</li>
      </ul>

      <p>
        <strong>Version:</strong> 1.0.0
        <br />
        <strong>Last Updated:</strong> {new Date().toLocaleDateString()}
      </p>

      <div className={styles.footer}>
        <p>Created by Cooper Morgan and Konnor Kooi.</p>
      </div>
      <div className={styles.footer}>
        <p>
          This project is not affiliated with Western Washington University. It
          is an independent initiative developed solely for educational and
          personal use. All data provided by this project is for informational
          purposes only and should not be considered official or binding. Use at
          your own discretion.
        </p>
      </div>
    </div>
  );
};

export const HelpContent: React.FC = () => {
  return (
    <div className={styles.helpContent}>
      <h2>How to Use the Schedule Optimizer</h2>

      <section className={styles.helpSection}>
        <h3>Getting Started</h3>
        <ol>
          <li>Select your desired quarter and year</li>
          <li>Choose minimum and maximum number of courses</li>
          <li>Add courses using the course selector</li>
        </ol>
      </section>

      <section className={styles.helpSection}>
        <h3>Adding Courses</h3>
        <ol>
          <li>Select a department from the dropdown (e.g., CSCI, MATH)</li>
          <li>Enter the course number (e.g., 141, 145)</li>
          <li>Click "Add" to include the course in your selection</li>
        </ol>
      </section>

      <section className={styles.helpSection}>
        <h3>Course Options</h3>
        <ul>
          <li>
            <strong>Force Course:</strong> Ensures a specific course is included
            in all generated schedules
          </li>
          <li>
            <strong>Remove Course:</strong> Removes a course from your selection
          </li>
        </ul>
      </section>

      <section className={styles.helpSection}>
        <h3>Generating Schedules</h3>
        <ol>
          <li>After adding your courses, click "Submit Schedule"</li>
          <li>
            Use the navigation buttons to view different schedule combinations
          </li>
          <li>Review any warnings or conflicts displayed</li>
        </ol>
      </section>

      <section className={styles.helpSection}>
        <h3>Tips</h3>
        <ul>
          <li>
            Add more courses than needed to see more possible combinations
          </li>
          <li>Use the force option for required courses</li>
          <li>Check for warnings about course conflicts or prerequisites</li>
        </ul>
      </section>
    </div>
  );
};
