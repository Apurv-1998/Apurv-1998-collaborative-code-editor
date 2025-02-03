package collaboration

import "time"

// Defines the type of message to sent over websocket
type MessageType string

const (
	// Code editing operation
	MessageTypeEdit MessageType = "edit"
	// Chat Message
	MessageTypeChat MessageType = "chat"
	// Room closed
	MessageTypeRoomClosed MessageType = "room_closed"
)

// Message to be sent over WebSocket
type Message struct {
	Type       MessageType `json:"type"`
	SenderID   string      `json:"sender_id"`
	SenderName string      `json:"sender_name"`
	Content    string      `json:"content"` // mesage content or edit delta
	Timestamp  time.Time   `json:"timestamp"`
	RoomID     string      `json:"room_id"`
}
