// src/components/RoomHistory.js
import React, { useState, useEffect } from 'react';
import { getRoomHistory } from '../services/roomService';
import { useNavigate } from 'react-router-dom';

const RoomHistory = () => {
  const [rooms, setRooms] = useState([]);
  const navigate = useNavigate();

  useEffect(() => {
    getRoomHistory()
      .then(data => setRooms(data))
      .catch(err => console.error(err));
  }, []);

  return (
    <div className="room-history">
      <h2>Past Rooms</h2>
      <ul>
        {rooms.map(room => (
          <li key={room.id}>
            {room.name} — Created: {new Date(room.created_at).toLocaleString()}
            <button onClick={() => navigate(`/editor/${room.id}?readOnly=true`)}>
              View
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
};

export default RoomHistory;
