// src/components/RoomJoin.js
import React, { useState } from 'react';
import { joinRoom } from '../services/roomService';
import { useNavigate } from 'react-router-dom';

const RoomJoin = () => {
  const navigate = useNavigate();
  const [inviteToken, setInviteToken] = useState('');
  const [error, setError] = useState('');

  const handleJoinRoom = async (e) => {
    e.preventDefault();
    setError('');
    try {
      // Call joinRoom and expect a response with room_id.
      const data = await joinRoom(inviteToken);
      if (data && data.room_id) {
        navigate(`/editor/${data.room_id}`);
      } else {
        setError('Room details not returned by server.');
      }
    } catch (err) {
      setError('Invalid invite token or server error.');
    }
  };

  return (
    <div className="room-join">
      <h2>Join a Room</h2>
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <form onSubmit={handleJoinRoom}>
        <label>Invite Token:</label>
        <input
          value={inviteToken}
          onChange={(e) => setInviteToken(e.target.value)}
          required
          placeholder="Enter your invitation token"
        />
        <button type="submit">Join Room</button>
      </form>
    </div>
  );
};

export default RoomJoin;
