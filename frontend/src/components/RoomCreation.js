import React, { useState, useContext } from 'react';
import { createRoom, generateInvite } from '../services/roomService';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from './Auth/AuthContext';

const RoomCreation = () => {
  const navigate = useNavigate();
  const { user } = useContext(AuthContext);
  const [roomName, setRoomName] = useState('');
  const [inviteToken, setInviteToken] = useState('');
  const [inviteEmail, setInviteEmail] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [currentRoomId, setCurrentRoomId] = useState(null);
  const [error, setError] = useState('');

  console.log("RoomCreation - user:", user);

  // Verify admin role (for testing, log user)
  console.log("RoomCreation - user:", user);

  // Create the room and, if admin, open the modal for email input.
  const handleCreateRoom = async (e) => {
    e.preventDefault();
    setError('');
    try {
      const room = await createRoom(roomName);
      setCurrentRoomId(room.id);

      // If the user is admin, open a modal to let them add an email for the invite.
      if (user && user.role === "admin") {
        setShowModal(true);
      } else {
        navigate(`/editor/${room.id}`);
      }
    } catch (err) {
      console.error("Error creating room:", err);
      setError('Error creating room');
    }
  };

  // Call generateInvite with the provided email, then navigate to the editor.
  const handleSendInvite = async () => {
    try {
      const inviteResponse = await generateInvite(currentRoomId, inviteEmail);
      setInviteToken(inviteResponse.token);
      setShowModal(false);
      navigate(`/editor/${currentRoomId}`);
    } catch (err) {
      console.error("Error generating invite:", err);
      setError('Error generating invite');
    }
  };


  return (
    <div className="room-creation">
      <h2>Create a Room</h2>
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <form onSubmit={handleCreateRoom}>
        <label>Room Name:</label>
        <input
          value={roomName}
          onChange={(e) => setRoomName(e.target.value)}
          required
        />
        <button type="submit">Create Room</button>
      </form>

      {/* Modal for admin to enter an email for the invite */}
      {showModal && (
        <div className="modal-overlay" style={modalOverlayStyle}>
          <div className="modal-content" style={modalContentStyle}>
            <h3>Send Invite</h3>
            <label>Email Address:</label>
            <input
              type="email"
              value={inviteEmail}
              onChange={(e) => setInviteEmail(e.target.value)}
              required
            />
            <div style={{ marginTop: '1rem' }}>
              <button onClick={handleSendInvite}>Send Invite</button>
              <button onClick={() => setShowModal(false)} style={{ marginLeft: '0.5rem' }}>
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Optional: Display invite token as fallback confirmation */}
      {inviteToken && (
        <div className="invite-details">
          <p>
            <strong>Invite Token:</strong> {inviteToken}
          </p>
          <p>An invite has been sent to the specified email.</p>
          <button onClick={() => navigator.clipboard.writeText(inviteToken)}>
            Copy Invite
          </button>
        </div>
      )}
    </div>
  );
};

// Inline styles for the modal overlay and content (you can replace these with your CSS)
const modalOverlayStyle = {
  position: 'fixed',
  top: 0,
  left: 0,
  right: 0,
  bottom: 0,
  backgroundColor: 'rgba(0,0,0,0.5)',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
};

const modalContentStyle = {
  backgroundColor: '#fff',
  padding: '2rem',
  borderRadius: '4px',
  width: '300px',
  textAlign: 'center',
};


export default RoomCreation;
