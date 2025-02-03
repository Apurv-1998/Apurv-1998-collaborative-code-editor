// src/components/Chat.js
import React, { useState, useEffect } from 'react';
import { getChatHistory } from '../services/chatService'; // A new service to fetch saved chat logs


const Chat = ({ roomId, websocket }) => {
  const [messages, setMessages] = useState([]);
  const [chatInput, setChatInput] = useState("");

  // Fetch persisted chat history on mount.
  useEffect(() => {
    getChatHistory(roomId)
      .then(data => setMessages(data))
      .catch(err => console.error("Error fetching chat history:", err));
  }, [roomId]);

  useEffect(() => {
    if (!websocket) return;
    const handleMessage = (event) => {
      const message = JSON.parse(event.data);
      if (message.type === "chat") {
        setMessages(prev => [...prev, message]);
      }
    };
    websocket.addEventListener("message", handleMessage);
    return () => {
      websocket.removeEventListener("message", handleMessage);
    };
  }, [websocket]);

  const sendChat = () => {
    if (websocket && websocket.readyState === WebSocket.OPEN) {
      websocket.send(JSON.stringify({ type: "chat", content: chatInput }));
      setChatInput("");
    }
  };

  return (
    <div className="chat-container">
      <h3>Chat</h3>
      <div className="chat-messages">
        {messages.map((msg, idx) => (
          <div key={idx}>
            <strong>{msg.sender_name || msg.sender_id}:</strong> {msg.content}
          </div>
        ))}
      </div>
      <div className="chat-input">
        <input
          value={chatInput}
          onChange={(e) => setChatInput(e.target.value)}
          placeholder="Type a message..."
        />
        <button onClick={sendChat}>Send</button>
      </div>
    </div>
  );
};

export default Chat;
