import React from 'react';
import './styles/variables.css';
import './styles/global.css';
import TopBar from './components/TopBar/TopBar';
import SchedulePlanner from './components/SchedulePlanner/SchedulePlanner';

const App: React.FC = () => {
  return (
    <div className="app">
      <TopBar />
      <main className="main-content">
        <SchedulePlanner />
      </main>
    </div>
  );
};

export default App;