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
        <strong>Last Updated:</strong> {new Date().toLocaleDateString()}
      </p>

      <div className={styles.footer}>
        <p>Created by:</p>
        <div className={styles.creators}>
          <a
            href="https://cwooper.me"
            target="_blank"
            rel="noopener noreferrer"
            className={styles.creatorLink}
          >
            Cooper Morgan
          </a>
          <a
            href="https://konnorkooi.com"
            target="_blank"
            rel="noopener noreferrer"
            className={styles.creatorLink}
          >
            Konnor Kooi
          </a>
        </div>
      </div>

      <div className={styles.disclaimer}>
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
          <li>Select your quarter and year from the top dropdowns</li>
          <li>
            Choose your minimum and maximum number of desired courses that will
            appear in each schedule
          </li>
          <li>Add your courses using the course selector</li>
        </ol>
      </section>

      <section className={styles.helpSection}>
        <h3>Adding Courses</h3>
        <ol>
          <li>Select a department from the dropdown (e.g., CSCI, MATH)</li>
          <li>Enter the course number (e.g., 141, 145)</li>
          <li>Click "Add" to include the course</li>
          <li>
            Your course selections are automatically saved between sessions
          </li>
        </ol>
      </section>

      <section className={styles.helpSection}>
        <h3>Course Management</h3>
        <ul>
          <li>
            <strong>Force Course:</strong> Click "Force" to ensure a specific
            course appears in all generated schedules
          </li>
          <li>
            <strong>Remove Course:</strong> Click "Remove" to delete a course
            from your selection
          </li>
          <li>
            <strong>Course Details:</strong> Click on any course to view
            additional information including prerequisites, instructors, and
            available seats
          </li>
        </ul>
      </section>

      <section className={styles.helpSection}>
        <h3>Schedule Generation</h3>
        <ol>
          <li>Click "Submit Schedule" after adding your desired courses</li>
          <li>
            Use "Previous" and "Next" buttons to navigate through different
            schedule combinations
          </li>
          <li>
            Click on any course block in the schedule to view detailed
            information
          </li>
          <li>Click the download button to save your schedule as an image</li>
        </ol>
      </section>

      <section className={styles.helpSection}>
        <h3>Schedule Preferences</h3>
        <ul>
          <li>
            <strong>Weights Button:</strong> Customize your schedule
            preferences:
            <ul>
              <li>Preferred start times</li>
              <li>Preferred end times</li>
              <li>Gap time between classes</li>
              <li>GPA considerations</li>
            </ul>
          </li>
        </ul>
      </section>

      <section className={styles.helpSection}>
        <h3>Course Search</h3>
        <ul>
          <li>
            Use the search bar to find courses by subject, title, or instructor
          </li>
          <li>Click on search results to view detailed course information</li>
          <li>Search results will show available seats and course details</li>
        </ul>
      </section>

      <section className={styles.helpSection}>
        <h3>Tips for Best Results</h3>
        <ul>
          <li>
            Add more courses than needed to see more possible combinations
          </li>
          <li>Use the force option for required courses in your major</li>
          <li>
            Pay attention to warnings about course conflicts or prerequisites
          </li>
          <li>Adjust the weights to better match your schedule preferences</li>
          <li>
            Check asynchronous courses section for online/flexible options
          </li>
        </ul>
      </section>

      <section className={styles.helpSection}>
        <h3>Understanding the Schedule View</h3>
        <ul>
          <li>Each course is color-coded for easy identification</li>
          <li>Course blocks show the instructor and location</li>
          <li>Asynchronous courses are listed separately below the schedule</li>
          <li>Warnings and errors appear above the schedule when relevant</li>
        </ul>
      </section>
    </div>
  );
};
