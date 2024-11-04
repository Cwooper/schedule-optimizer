// src/components/TopBar/TopBar.tsx
import React from "react";
import styles from "./TopBar.module.css";

// src/components/TopBar/TopBar.tsx
interface TopBarProps {
  onHelpClick?: () => void;
  onAboutClick?: () => void; // Add this line
}

const TopBar: React.FC<TopBarProps> = ({ onHelpClick, onAboutClick }) => {
  const handleHelpClick = () => {
    alert("TODO Help menu.");
  };

  const handleAboutClick = () => {
    alert("TODO About this application.");
  };

  return (
    <header className={styles.topbar}>
      <nav className={styles.container}>
        <div className={styles.content}>
          <h1 className={styles.title}>WWU Schedule Optimizer</h1>
          <div className={styles.buttonGroup}>
            <button
              className={`btn btn-primary ${styles.actionButton}`}
              onClick={onAboutClick || handleAboutClick}
              title="About this application"
            >
              About
            </button>
            <button
              className={`btn btn-primary ${styles.helpButton}`}
              onClick={onHelpClick || handleHelpClick}
              title="Display Help Menu"
            >
              Help
            </button>
          </div>
        </div>
      </nav>
    </header>
  );
};

export default TopBar;
