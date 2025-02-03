package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"example.com/collaborative-coding-editor/config"
	"example.com/collaborative-coding-editor/middleware"
	"example.com/collaborative-coding-editor/models"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// Register Request in the payload for registration
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register User
func Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}

	//Input validation
	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Missing Fields", http.StatusBadRequest)
		return
	}

	userCollection := GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check user if user with same email exists
	count, err := userCollection.CountDocuments(ctx, bson.M{"email": req.Email})
	if err != nil {
		log.Printf("Error checking existing user: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if count > 0 {
		http.Error(w, "User already exists", http.StatusBadRequest)
		return
	}

	// Hash the password and create the user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), config.AppConfig.BCryptCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	user := models.User{
		ID:        primitive.NewObjectID(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      "user",
		CreatedAt: time.Now(),
	}

	_, err = userCollection.InsertOne(ctx, user)
	if err != nil {
		log.Printf("Error inserting user: %v", err)
		http.Error(w, "Error inserting user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User successfully registered"})
}

// Login Request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Token Response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Login User
func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}

	//Input validation
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Missing Fields", http.StatusBadRequest)
		return
	}

	userCollection := GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Compare the provided and stored password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Create the access toke
	accessClaims := jwt.MapClaims{
		"user_id":  user.ID.Hex(),
		"role":     user.Role,
		"exp":      time.Now().Add(72 * time.Hour).Unix(),
		"email":    user.Email,
		"username": user.Username,
	}
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err := accessTokenObj.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		log.Printf("Error signing JWT: %v", err)
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Create a refresh token
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID.Hex(),
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
	}
	refereshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := refereshTokenObj.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		log.Printf("Error signing refresh token: %v", err)
		http.Error(w, "Error generating refresh token", http.StatusInternalServerError)
		return
	}

	// Store the refresh token in DB
	refreshCollection := GetCollection("refresh_tokens")
	refreshTokenRecord := models.RefreshToken{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}
	_, err = refreshCollection.InsertOne(ctx, refreshTokenRecord)
	if err != nil {
		log.Printf("Error storing refresh token: %v", err)
		http.Error(w, "Error processing login", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})

}

// Validate the refresh token and generate a new one
func Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}

	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(config.AppConfig.RefreshTokenSecret), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["user_id"] == nil {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}
	userID := claims["user_id"].(string)
	userRole := claims["role"].(string)

	refreshCollection := GetCollection("refresh_tokens")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var storedToken models.RefreshToken
	err = refreshCollection.FindOne(ctx, bson.M{"token": req.RefreshToken}).Decode(&storedToken)
	if err != nil {
		http.Error(w, "Refresh token not found", http.StatusUnauthorized)
		return
	}

	if time.Now().After(storedToken.ExpiresAt) {
		http.Error(w, "Refresh token expired", http.StatusUnauthorized)
		return
	}

	// TODO: delete the refresh token for one time usage

	newAccessClaims := jwt.MapClaims{
		"user_id": userID,
		"role":    userRole,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	}
	newAccessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, newAccessClaims)
	newAccessToken, err := newAccessTokenObj.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		log.Printf("Error signing new access token: %v", err)
		http.Error(w, "Error signing new access token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"access_token": newAccessToken,
	})
}

// Returns the active invitation for logged_in user
func GetActiveInvitations(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserKey).(jwt.MapClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userEmail, ok := claims["email"].(string)
	if !ok || userEmail == "" {
		http.Error(w, "Email not found in token", http.StatusUnauthorized)
		return
	}

	invitationCollection := GetCollection("invitations")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"invited_email": userEmail,
		"used":          false,
		"expires_at":    bson.M{"$gt": time.Now()},
	}

	cursor, err := invitationCollection.Find(ctx, filter)
	if err != nil {
		http.Error(w, "Error fetching invitations", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var invitations []models.Invitation
	if err = cursor.All(ctx, &invitations); err != nil {
		http.Error(w, "Error decoding invitations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invitations)
}

// Profile -> protected route that returns the JWT claims
func Profile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserKey).(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid user content", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(claims)
}
