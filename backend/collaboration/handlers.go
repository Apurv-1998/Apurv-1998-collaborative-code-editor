package collaboration

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"example.com/collaborative-coding-editor/config"
	"example.com/collaborative-coding-editor/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Global mapping of rooms and their hubs
var (
	hubs     = make(map[string]*Hub)
	hubMutex sync.Mutex
)

// This upgrades the HTTP to websocket connection
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Return hub for the room
func GetHub(roomID string) *Hub {
	hubMutex.Lock()
	defer hubMutex.Unlock()

	if hub, exists := hubs[roomID]; exists {
		return hub
	}

	// if doent exists create a new hub
	hub := NewHub()
	hubs[roomID] = hub

	// run the hub
	go hub.Run()
	return hub
}

// Upgrades the connection and registers the client
// WebSocketHandler upgrades the connection and registers the client.
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Try to get claims from context.
	var claims jwt.MapClaims
	if ctxClaims := r.Context().Value(middleware.UserKey); ctxClaims != nil {
		var ok bool
		claims, ok = ctxClaims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
	} else {
		// If no claims in context, try to get token from query parameter.
		tokenString := r.URL.Query().Get("token")
		if tokenString == "" {
			http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
			return
		}
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(config.AppConfig.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}
		var ok bool
		claims, ok = token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Unauthorized: invalid token claims", http.StatusUnauthorized)
			return
		}
	}

	// Verify token expiration.
	if exp, ok := claims["exp"].(float64); ok {
		if int64(exp) < time.Now().Unix() {
			http.Error(w, "Token expired", http.StatusUnauthorized)
			return
		}
	}

	// Extract the user ID.
	userID, ok := claims["user_id"].(string)
	if !ok {
		http.Error(w, "Invalid user id", http.StatusUnauthorized)
		return
	}
	userName, ok := claims["username"].(string)
	if userName == "" {
		userName = userID
	}

	// Extract the room ID from URL variables.
	vars := mux.Vars(r)
	roomID, ok := vars["room_id"]
	if !ok || roomID == "" {
		http.Error(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	// Upgrade the connection to a WebSocket.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Obtain the hub for the room.
	hub := GetHub(roomID)

	// Create a new client.
	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan Message, 256),
		userID:   userID,
		userName: userName,
	}

	// Broadcast the join event
	joinMsg := Message{
		Type:       MessageTypeChat,
		SenderID:   client.userID,
		SenderName: client.userName,
		Content:    fmt.Sprintf("%s joined the room", client.userName),
		Timestamp:  time.Now(),
	}
	hub.Broadcast <- joinMsg

	// Register the client with the hub.
	hub.Register <- client

	// Start client read and write pumps.
	go client.writePump()
	client.readPump()
}
