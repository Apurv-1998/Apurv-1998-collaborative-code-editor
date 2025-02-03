// src/components/Header.js
import React, { useContext } from "react";
import { Link, useNavigate } from "react-router-dom";
import { AuthContext } from "./Auth/AuthContext";

const Header = () => {
  const { user,logout } = useContext(AuthContext);
  const navigate = useNavigate();

  const isAdmin = user && user.role === "admin";

  const handleLogout = () => {
    logout(); // This clears tokens and state.
    navigate('/login'); // Navigate after logout.
  };

  return (
    <header className="app-header">
      <h1>Collaborative Coding Editor</h1>
      <nav>
        <Link to="/dashboard">Dashboard</Link>
        {isAdmin ? (
          <>
            <Link to="/room/create">Create Room</Link>
            <Link to="/auth/invitations">View Active Invitations</Link>
            {/* When in a room (Editor view), the Close Room button is rendered in Editor */}
          </>
        ) : (
          <>
            <Link to="/room/join">Join Room</Link>
            <Link to="/auth/invitations">View Active Invitations</Link>
          </>
        )}
       <Link to="/rooms/history">Room History</Link>
        <button onClick={logout}>Logout</button>
      </nav>
    </header>
  );
};

export default Header;
