// src/components/Dashboard.js
import React from 'react';
import { Link } from 'react-router-dom';

const Dashboard = () => {
  return (
    <div className="dashboard">
      <h2>Dashboard</h2>
      <div className="dashboard-actions">
        <Link to="/room/create">
          <button>Create Room (Admin)</button>
        </Link>
        <Link to="/room/join">
          <button>Join Room</button>
        </Link>
      </div>
    </div>
  );
};

export default Dashboard;
