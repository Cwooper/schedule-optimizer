// src/components/TopBar/TopBar.tsx
import React from 'react';
import styles from './TopBar.module.css';

interface TopBarProps {
  onHelpClick?: () => void;
}

const TopBar: React.FC<TopBarProps> = ({ onHelpClick }) => {
  const handleHelpClick = () => {
    const helpPopup = document.getElementById('help-popup');
    const backdrop = document.getElementById('backdrop');
    
    if (helpPopup && backdrop) {
      helpPopup.classList.remove('hidden');
      helpPopup.classList.add('block');
      backdrop.classList.remove('hidden');
      backdrop.classList.add('block');
    }
  };

  return (
    <header className={styles.topbar}>
      <nav className={styles.container}>
        <div className={styles.content}>
          <h1 className={styles.title}>
            WWU Schedule Optimizer
          </h1>
          <button
            className={styles.helpButton}
            onClick={onHelpClick || handleHelpClick}
            title="Display Help Menu"
          >
            Help
          </button>
        </div>
      </nav>
    </header>
  );
};

export default TopBar;