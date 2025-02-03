// src/services/compilerService.js
import api from './api';

export const compileCode = async (script, language, versionIndex, stdin = "") => {
  const payload = {
    script,
    language,
    versionIndex,
    stdin,
    clientId: process.env.REACT_APP_JDOODLE_CLIENT_ID,
    clientSecret: process.env.REACT_APP_JDOODLE_CLIENT_SECRET
  };
  const response = await api.post('/compile', payload);
  return response.data;
};
