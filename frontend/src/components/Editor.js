// src/components/Editor.js
import React, { useEffect, useState, useRef, useContext } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import MonacoEditor from "react-monaco-editor";
import { getSession, saveSession } from '../services/sessionService';
import { compileCode } from '../services/compilerService';
import Chat from './Chat';
import { AuthContext } from './Auth/AuthContext';
import { closeRoom } from '../services/roomService';

const Editor = () => {
  const { roomId } = useParams();
  const navigate = useNavigate();
  const { user } = useContext(AuthContext); // user should contain role and username
  const [code, setCode] = useState('');
  const [compileResult, setCompileResult] = useState('');
  const [ws, setWs] = useState(null);
  const [language, setLanguage] = useState('python3');
  const [versionIndex, setVersionIndex] = useState('3');
  const autoSaveRef = useRef(null);
  const [toast, setToast] = useState("");

  // Load session state on mount.
  useEffect(() => {
    const loadSession = async () => {
      try {
        const sessionData = await getSession(roomId);
        setCode(sessionData.code || '');
      } catch (err) {
        console.error('No session found, starting new session.');
      }
    };
    loadSession();
  }, [roomId]);

  // Establish WebSocket connection.
  useEffect(() => {
    const token = localStorage.getItem("access_token");
    const wsUrl = `${process.env.REACT_APP_WS_URL || 'ws://localhost:8080'}/collaboration/${roomId}?token=${encodeURIComponent(token)}`;
    const socket = new WebSocket(wsUrl);
    socket.onopen = () => console.log('WebSocket connected');
    socket.onmessage = (event) => {
      const message = JSON.parse(event.data);
      if (message.type === "edit") {
        setCode(message.content);
      } else if (message.content && message.content.includes("joined the room")) {
        // Optionally update presence UI.
      } else if (message.content && message.content.includes("Room has been closed")) {
        alert(message.content);
        navigate('/dashboard');
      }
      // Chat messages are handled in Chat component.
    };
    socket.onerror = (err) => console.error('WebSocket error:', err);
    socket.onclose = () => console.log('WebSocket disconnected');
    setWs(socket);
    return () => socket.close();
  }, [roomId, navigate]);

  // Auto-save every 30 seconds and on tab close.
  useEffect(() => {
    autoSaveRef.current = setInterval(() => {
      saveSession(roomId, code)
        .then(() => showToast("Code saved"))
        .catch((err) => console.error("Auto-save failed", err));
    }, 30000);

    const handleBeforeUnload = async (e) => {
      await saveSession(roomId, code);
      showToast("Code saved");
    };
    window.addEventListener("beforeunload", handleBeforeUnload);
    return () => {
      clearInterval(autoSaveRef.current);
      window.removeEventListener("beforeunload", handleBeforeUnload);
    };
  }, [roomId, code]);

  const handleEditorChange = (newValue, e) => {
    setCode(newValue);
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: "edit", content: newValue }));
    }
  };

  const handleCompile = async () => {
    try {
      const result = await compileCode(code, language, versionIndex);
      console.log("Compile result:", result); // Log the result
      setCompileResult(result.output || "No output");
    } catch (err) {
      console.error("Compilation error:", err);
      setCompileResult("Compilation failed");
    }
  };

  // Language drop-down options.
  const languageOptions = [
    { value: "python3", label: "Python" },
    { value: "javascript", label: "JavaScript" },
    { value: "go", label: "Go" },
    // Add more as needed.
  ];

  const showToast = (msg) => {
    setToast(msg);
    setTimeout(() => setToast(""), 3000);
  };

  // Determine if user is admin.
  const isAdmin = user && user.role === "admin";

  return (
    <div className="editor-container">
      <h2>Editor - Room: {roomId}</h2>
      <div className="editor-controls">
        <label htmlFor="language-select">Language:</label>
        <select
          id="language-select"
          value={language}
          onChange={(e) => setLanguage(e.target.value)}
          disabled={!isAdmin} // non-admins see read-only view.
        >
          {languageOptions.map(opt => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
      </div>
      <MonacoEditor
        width="800"
        height="600"
        language={language}
        theme="vs-dark"
        value={code}
        options={{
          readOnly: false, // allow editing for all users
          automaticLayout: true,
        }}
        onChange={handleEditorChange}
      />

      <div className="editor-actions">
        {<button onClick={handleCompile}>Compile & Run</button>}
      </div>
      {compileResult && (
        <div className="compile-result">
          <h3>Output:</h3>
          <pre>{compileResult}</pre>
        </div>
      )}
      <Chat roomId={roomId} websocket={ws} />
      {toast && <div className="toast">{toast}</div>}
      {isAdmin && (
        <button
          onClick={async () => {
            try {
              await closeRoom(roomId);
              alert("Room closed, redirecting to dashboard.");
              navigate('/dashboard');
            } catch (err) {
              alert("Failed to close room.");
            }
          }}
        >
          Close Room
        </button>
      )}
    </div>
  );
};

export default Editor;
