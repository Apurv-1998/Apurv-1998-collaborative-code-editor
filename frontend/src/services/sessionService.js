// src/services/sessionService.js
import api from './api';

export const saveSession = async (roomId, code) => {
  const response = await api.post('/session/save', { room_id: roomId, code });
  return response.data;
};

export const getSession = async (roomId) => {
  const response = await api.get(`/session/${roomId}`);
  return response.data;
};

export const exportSession = async (roomId) => {
  const response = await api.get(`/session/export/${roomId}`);
  return response.data;
};

export const logAudit = async (roomId, action, details) => {
  const response = await api.post('/session/audit', { room_id: roomId, action, details });
  return response.data;
};

export const getAuditLogs = async (roomId) => {
  const response = await api.get(`/session/audit/${roomId}`);
  return response.data;
};
