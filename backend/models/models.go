package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User model
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string             `bson:"username" json:"username"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"password"`
	Role      string             `bson:"role" json:"role"` // default "user"
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// Refresh Token
type RefreshToken struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id,omitempty" json:"user_id"`
	Token     string             `bson:"token" json:"token"`
	ExpiresAt time.Time          `bson:"expires_at" json:"expires_at"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// Room model
type Room struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	AdminID      string             `bson:"admin_id" json:"admin_id"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	InviteLimit  int                `bson:"invite_limit" json:"invite_limit"`
	Participants []string           `bson:"participants" json:"participants"`
	Status       string             `bson:"status" json:"status"` // open or closed
}

// Invitation Model
type Invitation struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoomID       primitive.ObjectID `bson:"room_id,omitempty" json:"room_id"`
	Token        string             `bson:"token" json:"token"`
	ExpiresAt    time.Time          `bson:"expires_at" json:"expires_at"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	Used         bool               `bson:"used" json:"used"`
	InvitedEmail string             `bson:"invited_email,omitempty" json:"invited_email,omitempty"`
}

// Session Model
type Session struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoomID     primitive.ObjectID `bson:"room_id" json:"room_id"` // Reference to the room
	Code       string             `bson:"code" json:"code"`       // The current code in the editor
	LastSaved  time.Time          `bson:"last_saved" json:"last_saved"`
	UpdatedBy  string             `bson:"updated_by" json:"updated_by"` // Last user (by user_id) who saved/updated
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	ModifiedAt time.Time          `bson:"modified_at" json:"modified_at"`
}

// Auditlog Model
type AuditLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoomID    primitive.ObjectID `bson:"room_id" json:"room_id"` // Associated room/session
	UserID    string             `bson:"user_id" json:"user_id"` // The user who performed the action
	Action    string             `bson:"action" json:"action"`   // e.g., "auto-save", "edit", "join", "export"
	Details   string             `bson:"details" json:"details"` // Additional information (could be a diff, etc.)
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
}
