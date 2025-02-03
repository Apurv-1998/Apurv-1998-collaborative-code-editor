// src/components/ProtectedRoute.js
import React, { useContext } from 'react';
import { AuthContext } from './Auth/AuthContext';
import { Navigate } from 'react-router-dom';

const ProtectedRoute = ({ children }) => {
  const { accessToken } = useContext(AuthContext);
  return accessToken ? children : <Navigate to="/login" />;
};

export default ProtectedRoute;
