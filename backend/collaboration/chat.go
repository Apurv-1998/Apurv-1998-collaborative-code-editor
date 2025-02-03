package collaboration

import (
	"context"
	"log"
	"time"

	"example.com/collaborative-coding-editor/auth"
	"go.mongodb.org/mongo-driver/bson"
)

func SaveChatMessage(msg Message, roomID string) {
	collection := auth.GetCollection("chat_logs")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, bson.M{
		"room_id":     roomID,
		"sender_id":   msg.SenderID,
		"sender_name": msg.SenderName,
		"content":     msg.Content,
		"timestamp":   msg.Timestamp,
	})

	if err != nil {
		log.Printf("Error saving chat message: %v", err)
	}
}
