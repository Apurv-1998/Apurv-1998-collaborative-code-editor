// src/components/Invite.js
import React from 'react';

const Invite = ({ token }) => {
  const copyToClipboard = () => {
    navigator.clipboard.writeText(token);
  };

  return (
    <div className="invite-container">
      <p>
        <strong>Invite Token:</strong> {token}
      </p>
      <button onClick={copyToClipboard}>Copy Invite Token</button>
    </div>
  );
};

export default Invite;
