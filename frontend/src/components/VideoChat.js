// src/components/VideoChat.js
import React, { useEffect, useRef, useState, useContext } from 'react';
import { AuthContext } from './Auth/AuthContext';

const VideoChat = ({ roomId, ws, videoEnabled }) => {
  const { user } = useContext(AuthContext);
  const localVideoRef = useRef(null);
  const [localStream, setLocalStream] = useState(null);
  // Define remoteStreams and its setter here:
  const [remoteStreams, setRemoteStreams] = useState([]);
  const peerConnections = useRef({}); // Map: peerId -> RTCPeerConnection

  // ICE servers configuration from environment variables.
  const iceServers = {
    iceServers: [
      { urls: process.env.REACT_APP_STUN_SERVER || "stun:stun.l.google.com:19302" }
    ]
  };

  // Start local video/audio stream
  useEffect(() => {
    if (!videoEnabled) return;
    async function startLocalStream() {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
        setLocalStream(stream);
        if (localVideoRef.current) {
          localVideoRef.current.srcObject = stream;
        }
        console.log("Local stream started.");
        // Send a video-introduce signal so that other peers know you are online.
        const message = {
          videoSignal: {
            type: "video-introduce",
            userId: user.user_id,
            username: user.username
          }
        };
        ws.send(JSON.stringify(message));
        console.log("Sent video-introduce message:", message);
      } catch (error) {
        console.error("Error accessing media devices:", error);
      }
    }
    startLocalStream();
  }, [videoEnabled, ws, user]);

  // Handle incoming signaling messages for video.
  useEffect(() => {
    if (!videoEnabled || !ws) return;
    
    const handleSignal = (event) => {
      let message;
      try {
        message = JSON.parse(event.data);
      } catch (err) {
        console.error("Error parsing signaling message:", err);
        return;
      }
      if (!message.videoSignal) return;
      const { type, from, sdp, candidate } = message.videoSignal;
      console.log("Received video signal:", message.videoSignal);
      if (from === user.user_id) return;
      
      if (type === "video-introduce") {
        // A new peer has introduced itself; if we don't have a connection with them, create one.
        if (!peerConnections.current[from]) {
          console.log("Handling video-introduce from:", from);
          const pc = createPeerConnection(from);
          peerConnections.current[from] = pc;
          // Create an offer to initiate connection.
          pc.createOffer()
            .then((offer) => pc.setLocalDescription(offer))
            .then(() => {
              const msg = {
                videoSignal: {
                  type: "video-offer",
                  sdp: pc.localDescription,
                  from: user.user_id,
                  to: from,
                }
              };
              ws.send(JSON.stringify(msg));
              console.log("Sent video-offer to:", from);
            })
            .catch((error) => {
              console.error("Error creating offer for new peer:", error);
            });
        }
      } else if (type === "video-offer") {
        handleVideoOffer(message.videoSignal);
      } else if (type === "video-answer") {
        handleVideoAnswer(message.videoSignal);
      } else if (type === "ice-candidate") {
        handleNewICECandidate(message.videoSignal);
      } else {
        console.warn("Unknown video signal type:", type);
      }
    };
    
    ws.addEventListener("message", handleSignal);
    return () => {
      ws.removeEventListener("message", handleSignal);
    };
  }, [videoEnabled, ws, user]);

  const createPeerConnection = (peerId) => {
    const pc = new RTCPeerConnection(iceServers);
    if (localStream) {
      localStream.getTracks().forEach((track) => {
        pc.addTrack(track, localStream);
      });
    }
    // ICE candidate handler
    pc.onicecandidate = (event) => {
      if (event.candidate) {
        const msg = {
          videoSignal: {
            type: "ice-candidate",
            candidate: event.candidate,
            from: user.user_id,
            to: peerId,
          }
        };
        ws.send(JSON.stringify(msg));
        console.log("Sent ICE candidate to:", peerId, event.candidate);
      }
    };
    // When a remote track is added, update remote streams.
    pc.ontrack = (event) => {
      console.log("ontrack event for peer", peerId, ":", event);
      const remoteStream = event.streams[0];
      if (remoteStream) {
        setRemoteStreams((prev) => {
          // Check if stream already exists by its id.
          if (!prev.find((s) => s.id === remoteStream.id)) {
            console.log("Adding remote stream from peer:", peerId);
            return [...prev, remoteStream];
          }
          return prev;
        });
      }
    };
    return pc;
  };

  const handleVideoOffer = async (signal) => {
    const { from, sdp } = signal;
    const pc = createPeerConnection(from);
    peerConnections.current[from] = pc;
    try {
      await pc.setRemoteDescription(new RTCSessionDescription(sdp));
      const answer = await pc.createAnswer();
      await pc.setLocalDescription(answer);
      const msg = {
        videoSignal: {
          type: "video-answer",
          sdp: pc.localDescription,
          from: user.user_id,
          to: from,
        }
      };
      ws.send(JSON.stringify(msg));
      console.log("Sent video-answer to:", from);
    } catch (error) {
      console.error("Error handling video offer:", error);
    }
  };

  const handleVideoAnswer = async (signal) => {
    const { from, sdp } = signal;
    const pc = peerConnections.current[from];
    if (pc) {
      try {
        await pc.setRemoteDescription(new RTCSessionDescription(sdp));
        console.log("Set remote description for peer:", from);
      } catch (error) {
        console.error("Error handling video answer:", error);
      }
    }
  };

  const handleNewICECandidate = async (signal) => {
    const { from, candidate } = signal;
    const pc = peerConnections.current[from];
    if (pc) {
      try {
        await pc.addIceCandidate(new RTCIceCandidate(candidate));
        console.log("Added ICE candidate from peer:", from);
      } catch (error) {
        console.error("Error adding ICE candidate:", error);
      }
    }
  };

  // When a new peer announces presence, send them an offer.
  useEffect(() => {
    if (!videoEnabled || !ws) return;

    const handleIntroduce = (event) => {
      let message;
      try {
        message = JSON.parse(event.data);
      } catch (err) {
        console.error("Error parsing introduction message:", err);
        return;
      }
      if (message.videoSignal && message.videoSignal.type === "video-introduce") {
        const newPeerId = message.videoSignal.userId;
        console.log("Handling video-introduce from", newPeerId);
        if (newPeerId === user.user_id) return;
        if (peerConnections.current[newPeerId]) return;
        const pc = createPeerConnection(newPeerId);
        peerConnections.current[newPeerId] = pc;
        pc.createOffer()
          .then((offer) => pc.setLocalDescription(offer))
          .then(() => {
            const msg = {
              videoSignal: {
                type: "video-offer",
                sdp: pc.localDescription,
                from: user.user_id,
                to: newPeerId,
              }
            };
            ws.send(JSON.stringify(msg));
            console.log("Sent video-offer to", newPeerId);
          })
          .catch((error) => {
            console.error("Error creating offer for new peer:", error);
          });
      }
    };

    ws.addEventListener("message", handleIntroduce);
    return () => {
      ws.removeEventListener("message", handleIntroduce);
    };
  }, [videoEnabled, ws, user]);

  return (
    <div className="video-chat">
      <h3>Video Chat</h3>
      <div className="local-video">
        <video ref={localVideoRef} autoPlay muted style={{ width: "200px" }} />
      </div>
      <div className="remote-videos">
        {remoteStreams.map((stream) => (
          <video
            key={stream.id}
            autoPlay
            playsInline
            style={{ width: "200px" }}
            ref={(el) => {
              if (el) {
                el.srcObject = stream;
                el.play().catch((err) => {
                  console.error("Error playing remote video:", err);
                });
              }
            }}
          />
        ))}
      </div>
    </div>
  );
};

export default VideoChat;
