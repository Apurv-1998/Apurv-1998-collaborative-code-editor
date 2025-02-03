// src/services/chatService.js
import api from './api';

export const getChatHistory = async (roomId) => {
  const response = await api.get(`/chat/${roomId}`); // Ensure your backend has this endpoint.
  return response.data;
};