// src/components/TopBar/TopBar.tsx
import React, { useState } from "react";
import styles from "./TopBar.module.css";
import Popup from "../Popup/Popup";
import { AboutContent, HelpContent } from "../Popup/PopupContent";

const TopBar: React.FC = () => {
  const [isAboutOpen, setIsAboutOpen] = useState(false);
  const [isHelpOpen, setIsHelpOpen] = useState(false);

  return (
    <header className={styles.topbar}>
      <nav className={styles.container}>
        <div className={styles.content}>
          <h1 className={styles.title}>WWU Schedule Optimizer</h1>
          <div className={styles.buttonGroup}>
            <button
              className={`btn btn-primary ${styles.actionButton}`}
              onClick={() => setIsAboutOpen(true)}
              title="About this application"
            >
              About
            </button>
            <button
              className={`btn btn-primary ${styles.helpButton}`}
              onClick={() => setIsHelpOpen(true)}
              title="Display Help Menu"
            >
              Help
            </button>
          </div>
        </div>
      </nav>

      <Popup
        isOpen={isAboutOpen}
        onClose={() => setIsAboutOpen(false)}
        width="600px"
      >
        <AboutContent />
      </Popup>

      <Popup
        isOpen={isHelpOpen}
        onClose={() => setIsHelpOpen(false)}
        width="800px"
      >
        <HelpContent />
      </Popup>
    </header>
  );
};

export default TopBar;
