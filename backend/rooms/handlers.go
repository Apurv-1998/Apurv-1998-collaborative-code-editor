package rooms

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"example.com/collaborative-coding-editor/auth"
	"example.com/collaborative-coding-editor/collaboration"
	"example.com/collaborative-coding-editor/middleware"
	"example.com/collaborative-coding-editor/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	DefaultInviteLimit       = 3
	InvitationExpiryDuration = 24 * time.Hour
)

// Create Room Request
type CreateRoomRequest struct {
	Name string `json:"name"`
}

// Create Room
func CreateRoom(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserKey).(jwt.MapClaims)
	fmt.Printf("claims: %v", claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fmt.Printf("claims: %v", claims)

	// only admins can create room
	if claims["role"] != "admin" {
		http.Error(w, "Only admins can create rooms", http.StatusForbidden)
		return
	}

	adminID, ok := claims["user_id"].(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Room name is required", http.StatusBadRequest)
		return
	}

	newRoom := models.Room{
		ID:           primitive.NewObjectID(),
		Name:         req.Name,
		AdminID:      adminID,
		CreatedAt:    time.Now(),
		InviteLimit:  DefaultInviteLimit,
		Participants: []string{},
		Status:       "active",
	}

	roomCollection := auth.GetCollection("rooms")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := roomCollection.InsertOne(ctx, newRoom)
	if err != nil {
		http.Error(w, "Error crerating room", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newRoom)
}

type GenerateInviteRequest struct {
	InvitedEmail string `json:"invited_email,omitempty"`
}

// Generate Invite Response
type GenerateInviteResponse struct {
	Token string `json:"token"`
}

// Generate Invite
func GenerateInvite(w http.ResponseWriter, r *http.Request) {
	// extract roomId from the URL
	vars := mux.Vars(r)
	roomIdStr, ok := vars["room_id"]
	if !ok {
		http.Error(w, "Room Id is required", http.StatusBadRequest)
		return
	}

	roomID, err := primitive.ObjectIDFromHex(roomIdStr)
	if err != nil {
		http.Error(w, "Invalid Room Id", http.StatusBadRequest)
		return
	}

	claims, ok := r.Context().Value(middleware.UserKey).(jwt.MapClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// only admin can generate invite
	if claims["role"] != "admin" {
		http.Error(w, "Only admins can generate invited", http.StatusForbidden)
		return
	}

	adminID, ok := claims["user_id"].(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Verify the room exists and the requster is room admin
	roomCollection := auth.GetCollection("rooms")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var room models.Room
	err = roomCollection.FindOne(ctx, bson.M{"_id": roomID}).Decode(&room)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}
	if room.AdminID != adminID {
		http.Error(w, "You are not the admin of the room", http.StatusForbidden)
		return
	}

	// Check the number of invites already generated
	inviteCollection := auth.GetCollection("invitations")
	count, err := inviteCollection.CountDocuments(ctx, bson.M{"room_id": roomID})

	if err != nil {
		http.Error(w, "Error counting invites", http.StatusInternalServerError)
		return
	}
	if count > int64(DefaultInviteLimit) {
		http.Error(w, "Invite limit reached for this room", http.StatusForbidden)
		return
	}

	// Decode optional invited email
	var req GenerateInviteRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	// Generate a secure random token
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		log.Printf("Error generating token: %v", err)
		http.Error(w, "Error generating invite token", http.StatusInternalServerError)
		return
	}
	token := hex.EncodeToString(tokenBytes)

	// Create an invitation record
	newInvitation := models.Invitation{
		ID:           primitive.NewObjectID(),
		RoomID:       roomID,
		Token:        token,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(InvitationExpiryDuration),
		Used:         false,
		InvitedEmail: req.InvitedEmail,
	}

	_, err = inviteCollection.InsertOne(ctx, newInvitation)
	if err != nil {
		log.Printf("Error saving invitation: %v", err)
		http.Error(w, "Error saving invitation", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(GenerateInviteResponse{
		Token: token,
	})

}

// Join Room Request
type JoinRoomRequest struct {
	Token string `json:"token"`
}

// Join Room -> allows participant to join the room
func JoinRoom(w http.ResponseWriter, r *http.Request) {
	var req JoinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// Retrieve users identity
	claims, ok := r.Context().Value(middleware.UserKey).(jwt.MapClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		http.Error(w, "Invalid User Id", http.StatusUnauthorized)
		return
	}

	// look up the invitations by token
	inviteCollection := auth.GetCollection("invitations")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var invitation models.Invitation
	err := inviteCollection.FindOne(ctx, bson.M{"token": req.Token}).Decode(&invitation)
	if err != nil {
		http.Error(w, "Invlaid Invite Token", http.StatusBadRequest)
		return
	}

	// Validate invitation is not expired
	if time.Now().After(invitation.ExpiresAt) {
		http.Error(w, "Invitation has expired", http.StatusBadRequest)
		return
	}
	if invitation.Used {
		http.Error(w, "Inviation token already used", http.StatusBadRequest)
		return
	}

	// Add the participant to the room
	roomCollection := auth.GetCollection("rooms")
	filter := bson.M{"_id": invitation.RoomID}
	update := bson.M{
		"$addToSet": bson.M{"participants": userID},
	}

	_, err = roomCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("Error adding participant to room: %v", err)
		http.Error(w, "Error joining room", http.StatusInternalServerError)
		return
	}

	// Mark the invitation used
	_, err = inviteCollection.UpdateOne(ctx, bson.M{"_id": invitation.ID}, bson.M{"$set": bson.M{"used": true}})

	if err != nil {
		log.Printf("Error aupdating invitation: %v", err)
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Room joined successfully", "room_id": invitation.RoomID.Hex()})
}

// Get the room history
func GetRoomHistory(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserKey).(jwt.MapClaims)

	// Debug log the claims
	log.Printf("GetRoomHistory - Claims: %#v", claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := claims["user_id"].(string)
	if !ok {
		http.Error(w, "Invalid user id", http.StatusUnauthorized)
		return
	}

	roomCollection := auth.GetCollection("rooms")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"admin_id": userID},
			{"participants": userID},
		},
	}

	cursor, err := roomCollection.Find(ctx, filter)
	if err != nil {
		http.Error(w, "Error fetching room history", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var rooms []models.Room
	if err := cursor.All(ctx, &rooms); err != nil {
		http.Error(w, "Error decoding rooms", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

// Admin can close the room
func CloseRoom(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserKey).(jwt.MapClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, ok := claims["user_id"].(string)
	if !ok {
		http.Error(w, "Invalid user id", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	roomIDStr, ok := vars["room_id"]
	if !ok || roomIDStr == "" {
		http.Error(w, "Room ID is required", http.StatusBadRequest)
		return
	}
	roomID, err := primitive.ObjectIDFromHex(roomIDStr)
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	roomCollection := auth.GetCollection("rooms")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var room models.Room
	if err := roomCollection.FindOne(ctx, bson.M{"_id": roomID}).Decode(&room); err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}
	if room.AdminID != userID {
		http.Error(w, "Only admin can close the room", http.StatusForbidden)
		return
	}

	// Update room status to closed.
	_, err = roomCollection.UpdateOne(ctx, bson.M{"_id": roomID}, bson.M{"$set": bson.M{"status": "closed"}})
	if err != nil {
		http.Error(w, "Error closing room", http.StatusInternalServerError)
		return
	}

	//Broadcast a room closure message
	hub := collaboration.GetHub(roomIDStr)
	closeMsg := collaboration.Message{
		Type:       collaboration.MessageTypeRoomClosed,
		SenderID:   userID,
		SenderName: "System",
		Content:    "Room has been closed by admin. You will be logged out.",
		Timestamp:  time.Now(),
	}

	hub.Broadcast <- closeMsg

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Room Closed"})
}

// Request Room Details for both admin and participant
func GetRoomDetails(w http.ResponseWriter, r *http.Request) {
	// Get room id
	vars := mux.Vars(r)
	roomIdStr, ok := vars["room_id"]
	if !ok {
		http.Error(w, "Room Id is required", http.StatusBadRequest)
		return
	}

	roomID, err := primitive.ObjectIDFromHex(roomIdStr)
	if err != nil {
		http.Error(w, "Invalid Room Id", http.StatusBadRequest)
		return
	}

	// Retrieve users identity
	claims, ok := r.Context().Value(middleware.UserKey).(jwt.MapClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		http.Error(w, "Invalid User Id", http.StatusUnauthorized)
		return
	}

	//Lookup room
	roomCollection := auth.GetCollection("rooms")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var room models.Room
	err = roomCollection.FindOne(ctx, bson.M{"_id": roomID}).Decode(&room)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	// Allow access only if admin/participant
	if room.AdminID != userID {
		participant := false
		for _, pid := range room.Participants {
			if pid == userID {
				participant = true
				break
			}
		}

		if !participant {
			http.Error(w, "You are not participant of this room", http.StatusForbidden)
			return
		}
	}

	json.NewEncoder(w).Encode(room)

}
