package session

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"example.com/collaborative-coding-editor/auth"
	"example.com/collaborative-coding-editor/middleware"
	"example.com/collaborative-coding-editor/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SaveSessionRequest represents the payload for auto-saving a session.
type SaveSessionRequest struct {
	RoomID string `json:"room_id"` // Room ID (as hex string) to which this session belongs.
	Code   string `json:"code"`    // The current code in the editor.
}

// SaveSession creates or updates the session state for a given room.
func SaveSession(w http.ResponseWriter, r *http.Request) {
	var req SaveSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if req.RoomID == "" {
		http.Error(w, "RoomID is required", http.StatusBadRequest)
		return
	}

	roomID, err := primitive.ObjectIDFromHex(req.RoomID)
	if err != nil {
		http.Error(w, "Invalid RoomID", http.StatusBadRequest)
		return
	}

	// Retrieve user info from JWT (set by middleware)
	claims, ok := r.Context().Value(middleware.UserKey).(jwt.MapClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, _ := claims["user_id"].(string)

	sessionCollection := auth.GetCollection("sessions")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to update an existing session for this room.
	filter := bson.M{"room_id": roomID}
	update := bson.M{
		"$set": bson.M{
			"code":        req.Code,
			"last_saved":  time.Now(),
			"updated_by":  userID,
			"modified_at": time.Now(),
		},
		"$setOnInsert": bson.M{
			"created_at": time.Now(),
		},
	}
	opts := true
	_, err = sessionCollection.UpdateOne(ctx, filter, update, &options.UpdateOptions{
		Upsert: &opts,
	})
	if err != nil {
		log.Printf("Error saving session: %v", err)
		http.Error(w, "Error saving session", http.StatusInternalServerError)
		return
	}

	// Optionally, add an audit log entry for auto-save.
	go addAuditLog(roomID, userID, "auto-save", "Session auto-saved.")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Session saved successfully"})
}

// GetSession retrieves the current session state for a given room.
func GetSession(w http.ResponseWriter, r *http.Request) {
	// Expect room_id in the URL.
	vars := mux.Vars(r)
	roomIDStr := vars["room_id"]
	if roomIDStr == "" {
		http.Error(w, "RoomID is required", http.StatusBadRequest)
		return
	}
	roomID, err := primitive.ObjectIDFromHex(roomIDStr)
	if err != nil {
		http.Error(w, "Invalid RoomID", http.StatusBadRequest)
		return
	}

	sessionCollection := auth.GetCollection("sessions")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var sess models.Session
	err = sessionCollection.FindOne(ctx, bson.M{"room_id": roomID}).Decode(&sess)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sess)
}

// ExportSession exports the session state along with its audit trail for a given room.
func ExportSession(w http.ResponseWriter, r *http.Request) {
	// Get the room_id from the URL.
	vars := mux.Vars(r)
	roomIDStr := vars["room_id"]
	if roomIDStr == "" {
		http.Error(w, "RoomID is required", http.StatusBadRequest)
		return
	}
	roomID, err := primitive.ObjectIDFromHex(roomIDStr)
	if err != nil {
		http.Error(w, "Invalid RoomID", http.StatusBadRequest)
		return
	}

	sessionCollection := auth.GetCollection("sessions")
	auditCollection := auth.GetCollection("audit_logs")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Retrieve session state.
	var sess models.Session
	err = sessionCollection.FindOne(ctx, bson.M{"room_id": roomID}).Decode(&sess)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Retrieve audit logs for this session.
	cursor, err := auditCollection.Find(ctx, bson.M{"room_id": roomID})
	if err != nil {
		http.Error(w, "Error retrieving audit logs", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var audits []models.AuditLog
	for cursor.Next(ctx) {
		var a models.AuditLog
		if err := cursor.Decode(&a); err != nil {
			log.Printf("Error decoding audit log: %v", err)
			continue
		}
		audits = append(audits, a)
	}

	// Build the export payload.
	exportPayload := struct {
		Session   models.Session    `json:"session"`
		AuditLogs []models.AuditLog `json:"audit_logs"`
	}{
		Session:   sess,
		AuditLogs: audits,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exportPayload)
}

// LogAuditRequest is the payload for manually logging an audit event.
type LogAuditRequest struct {
	RoomID  string `json:"room_id"`
	Action  string `json:"action"`
	Details string `json:"details"`
}

// LogAudit logs an audit event for a session.
func LogAudit(w http.ResponseWriter, r *http.Request) {
	var req LogAuditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if req.RoomID == "" || req.Action == "" {
		http.Error(w, "RoomID and Action are required", http.StatusBadRequest)
		return
	}
	roomID, err := primitive.ObjectIDFromHex(req.RoomID)
	if err != nil {
		http.Error(w, "Invalid RoomID", http.StatusBadRequest)
		return
	}

	// Retrieve user from JWT.
	claims, ok := r.Context().Value(middleware.UserKey).(jwt.MapClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, _ := claims["user_id"].(string)

	if err := addAuditLog(roomID, userID, req.Action, req.Details); err != nil {
		http.Error(w, "Error logging audit event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Audit log added"})
}

// GetAuditLogs retrieves audit logs for a given room.
func GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomIDStr := vars["room_id"]
	if roomIDStr == "" {
		http.Error(w, "RoomID is required", http.StatusBadRequest)
		return
	}
	roomID, err := primitive.ObjectIDFromHex(roomIDStr)
	if err != nil {
		http.Error(w, "Invalid RoomID", http.StatusBadRequest)
		return
	}

	auditCollection := auth.GetCollection("audit_logs")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := auditCollection.Find(ctx, bson.M{"room_id": roomID})
	if err != nil {
		http.Error(w, "Error retrieving audit logs", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var audits []models.AuditLog
	for cursor.Next(ctx) {
		var a models.AuditLog
		if err := cursor.Decode(&a); err != nil {
			log.Printf("Error decoding audit log: %v", err)
			continue
		}
		audits = append(audits, a)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(audits)
}

// addAuditLog is a helper function to insert an audit log entry.
// It is used both synchronously (via the LogAudit endpoint) and asynchronously (e.g., auto-save events).
func addAuditLog(roomID primitive.ObjectID, userID, action, details string) error {
	auditCollection := auth.GetCollection("audit_logs")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	audit := models.AuditLog{
		ID:        primitive.NewObjectID(),
		RoomID:    roomID,
		UserID:    userID,
		Action:    action,
		Details:   details,
		Timestamp: time.Now(),
	}
	_, err := auditCollection.InsertOne(ctx, audit)
	if err != nil {
		log.Printf("Error inserting audit log: %v", err)
	}
	return err
}
