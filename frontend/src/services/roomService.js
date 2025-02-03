// src/services/roomService.js
import api from './api';

export const createRoom = async (roomName) => {
  const response = await api.post('/rooms', { name: roomName });
  return response.data;
};

export const generateInvite = async (roomId) => {
  const response = await api.post(`/rooms/${roomId}/invite`);
  return response.data;
};

export const joinRoom = async (inviteToken) => {
  // The backend validates the invite token and returns room details.
  const response = await api.post('/rooms/join', { token: inviteToken });
  return response.data; // Expected response: { message, room_id }
};

export const getRoomDetails = async (roomId) => {
  const response = await api.get(`/rooms/${roomId}`);
  return response.data;
};

export const getActiveInvitations = async () => {
  const response = await api.get('/auth/invitations');
  return response.data;
};

export const getRoomHistory = async () => {
  const response = await api.get('/rooms/history');
  return response.data;
};

export const closeRoom = async (roomId) => {
  const response = await api.post(`/rooms/${roomId}/close`);
  return response.data;
};
