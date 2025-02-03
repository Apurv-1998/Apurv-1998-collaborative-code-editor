// src/App.js
import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './components/Auth/AuthContext';
import Login from './components/Auth/Login';
import Register from './components/Auth/Register';
import Dashboard from './components/Dashboard';
import RoomCreation from './components/RoomCreation';
import RoomJoin from './components/RoomJoin';
import Editor from './components/Editor';
import ProtectedRoute from './components/ProtectedRoute';
import Header from './components/Header';  // <-- Header is imported here
import ActiveInvitations from './components/ActiveInvitations';
import RoomHistory from './components/RoomHistory';

const App = () => {
  return (
    <AuthProvider>
      <Router>
        <Header />
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/dashboard" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
          <Route path="/room/create" element={<ProtectedRoute><RoomCreation /></ProtectedRoute>} />
          <Route path="/room/join" element={<ProtectedRoute><RoomJoin /></ProtectedRoute>} />
          <Route path="/editor/:roomId" element={<ProtectedRoute><Editor /></ProtectedRoute>} />
          <Route path="/auth/invitations" element={<ProtectedRoute><ActiveInvitations /></ProtectedRoute>} />
          <Route path="/rooms/history" element={<ProtectedRoute><RoomHistory /></ProtectedRoute>} />
          <Route path="*" element={<Login />} />
        </Routes>
      </Router>
    </AuthProvider>
  );
};
export default App;
