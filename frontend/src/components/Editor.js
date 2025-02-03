// src/components/Editor.js
import React, { useEffect, useState, useRef, useContext } from 'react';
import MonacoEditor from "react-monaco-editor";
import VideoChat from './VideoChat';
import Chat from './Chat';
import { getSession, saveSession, getRoomDetails } from '../services/sessionService';
import { compileCode } from '../services/compilerService';
import { closeRoom } from '../services/roomService';
import { AuthContext } from './Auth/AuthContext';
import { useParams, useNavigate } from 'react-router-dom';
import SettingsPanel from './SettingsPanel'; // Optional settings panel

const Editor = () => {
  const { roomId } = useParams();
  const navigate = useNavigate();
  const { user } = useContext(AuthContext);
  const [code, setCode] = useState('');
  const [compileResult, setCompileResult] = useState('');
  const [ws, setWs] = useState(null);
  const [language, setLanguage] = useState('python3');
  const [versionIndex, setVersionIndex] = useState('3');
  const [editorOptions, setEditorOptions] = useState({
    automaticLayout: true,
    readOnly: false,
    fontSize: 14,
    minimap: { enabled: true },
  });
  const [roomDetails, setRoomDetails] = useState(null);
  const autoSaveRef = useRef(null);
  const [toast, setToast] = useState("");

  // Load session and room details on mount.
  useEffect(() => {
    (async () => {
      try {
        const sessionData = await getSession(roomId);
        setCode(sessionData.code || '');
      } catch (err) {
        console.error('No session found, starting new session.');
      }
      try {
        const details = await getRoomDetails(roomId);
        setRoomDetails(details);
      } catch (err) {
        console.error("Error fetching room details:", err);
      }
    })();
  }, [roomId]);

  // Establish WebSocket connection (token in query parameter)
  useEffect(() => {
    const token = localStorage.getItem("access_token");
    const wsUrl = `${process.env.REACT_APP_WS_URL || 'ws://localhost:8080'}/collaboration/${roomId}?token=${encodeURIComponent(token)}`;
    const socket = new WebSocket(wsUrl);
    socket.onopen = () => console.log('WebSocket connected');
    socket.onmessage = (event) => {
      const message = JSON.parse(event.data);
      if (message.type === "edit") {
        setCode(message.content);
      } else if (message.content && message.content.includes("Room has been closed")) {
        alert(message.content);
        navigate('/dashboard');
      }
      // Chat and other messages are handled by the Chat component.
    };
    socket.onerror = (err) => console.error('WebSocket error:', err);
    socket.onclose = () => console.log('WebSocket disconnected');
    setWs(socket);
    return () => socket.close();
  }, [roomId, navigate]);

  // Auto-save code every 30 seconds and on tab close.
  useEffect(() => {
    autoSaveRef.current = setInterval(() => {
      saveSession(roomId, code)
        .then(() => showToast("Code saved"))
        .catch((err) => console.error("Auto-save failed", err));
    }, 30000);
    const handleBeforeUnload = async () => {
      await saveSession(roomId, code);
      showToast("Code saved");
    };
    window.addEventListener("beforeunload", handleBeforeUnload);
    return () => {
      clearInterval(autoSaveRef.current);
      window.removeEventListener("beforeunload", handleBeforeUnload);
    };
  }, [roomId, code]);

  const handleEditorChange = (newValue) => {
    setCode(newValue);
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: "edit", content: newValue }));
    }
  };

  const handleCompile = async () => {
    try {
      const result = await compileCode(code, language, versionIndex);
      setCompileResult(result.output || "No output");
    } catch (err) {
      setCompileResult("Compilation failed");
    }
  };

  const showToast = (msg) => {
    setToast(msg);
    setTimeout(() => setToast(""), 3000);
  };

  const isAdmin = user && user.role === "admin";

  return (
    <div className="workspace-container">
      {/* Left side: Code editor */}
      <div className="editor-left">
        <SettingsPanel onSettingsChange={setEditorOptions} />
        <div className="editor-controls">
          <label htmlFor="language-select">Language:</label>
          <select
            id="language-select"
            value={language}
            onChange={(e) => setLanguage(e.target.value)}
            disabled={!isAdmin}  // If only admins can change language
          >
            <option value="python3">Python 3</option>
            <option value="javascript">JavaScript</option>
            <option value="go">Go</option>
          </select>
        </div>
        <MonacoEditor
          width="100%"
          height="600"
          language={language}
          theme={editorOptions.theme || "vs-dark"}
          value={code}
          options={editorOptions}
          onChange={handleEditorChange}
        />
        <div className="editor-actions">
          {<button onClick={handleCompile}>Compile & Run</button>}
          {compileResult && (
            <div className="compile-result">
              <h3>Output:</h3>
              <pre>{compileResult}</pre>
            </div>
          )}
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
        {toast && <div className="toast">{toast}</div>}
      </div>
      {/* Right side: Top half: Video Chat; Bottom half: Chat */}
      <div className="editor-right">
        <div className="video-container">
          {roomDetails && roomDetails.video_enabled && (
            <VideoChat roomId={roomId} ws={ws} videoEnabled={true} />
          )}
        </div>
        <div className="chat-container">
          <Chat roomId={roomId} websocket={ws} />
        </div>
      </div>
    </div>
  );
};

export default Editor;
