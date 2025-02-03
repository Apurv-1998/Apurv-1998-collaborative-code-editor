// src/components/ActiveInvitations.js
import React, { useState, useEffect } from 'react';
import { getActiveInvitations } from '../services/roomService';

const ActiveInvitations = () => {
  const [invitations, setInvitations] = useState([]);

  useEffect(() => {
    getActiveInvitations()
      .then(data => setInvitations(data))
      .catch(err => console.error(err));
  }, []);

  return (
    <div className="active-invitations">
      <h2>Active Invitations</h2>
      <ul>
        {invitations?.map(inv => (
          <li key={inv.id}>
            Room ID: {inv.room_id} — Expires: {new Date(inv.expires_at).toLocaleString()}
          </li>
        ))}
      </ul>
    </div>
  );
};

export default ActiveInvitations;
