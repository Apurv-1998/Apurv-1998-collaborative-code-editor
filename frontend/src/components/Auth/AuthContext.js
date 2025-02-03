import React, { createContext, useState, useEffect } from 'react';
import { getLocalAccessToken,clearLocalTokens } from '../../services/authService';
import { jwtDecode } from 'jwt-decode'; // Use named import if needed

export const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
  const [accessToken, setAccessToken] = useState(getLocalAccessToken());
  const [user, setUser] = useState(null);

  useEffect(() => {
    if (accessToken) {
      try {
        const decoded = jwtDecode(accessToken); // Ensure token includes role & username
        setUser(decoded);
      } catch (error) {
        console.error("Token decode error:", error);
      }
    }
  }, [accessToken]);

  const logout = () => {
    // Clear tokens from storage.
    clearLocalTokens();
    // Clear user state.
    setAccessToken(null);
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ accessToken, setAccessToken, user, setUser, logout }}>
      {children}
    </AuthContext.Provider>
  );
};
