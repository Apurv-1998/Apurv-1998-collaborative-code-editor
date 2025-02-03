// src/services/authService.js
import api from './api';

// Token storage helpers.
export const getLocalAccessToken = () => localStorage.getItem('access_token');
export const getLocalRefreshToken = () => localStorage.getItem('refresh_token');
export const setLocalAccessToken = (token) => localStorage.setItem('access_token', token);
export const setLocalTokens = ({ access_token, refresh_token }) => {
  localStorage.setItem('access_token', access_token);
  localStorage.setItem('refresh_token', refresh_token);
};
export const clearLocalTokens = () => {
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
};

export const login = async (email, password) => {
  const response = await api.post('/auth/login', { email, password });
  setLocalTokens(response.data);
  return response.data;
};

export const register = async (username, email, password) => {
  const response = await api.post('/auth/register', { username, email, password });
  return response.data;
};


export const invitations = async () => {
  const response = await api.get('/auth/invitations');
  return response.data;
};
