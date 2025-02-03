package auth

import (
	"context"
	"log"
	"time"

	"example.com/collaborative-coding-editor/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

// Connect to MongoDB Client
func Connect() *mongo.Client {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.AppConfig.MongoURI)
	Client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("MongoDB connection error: %v", err)
	}

	// Ping to verify connection
	err = Client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("MongoDB ping error: %v", err)
	}

	log.Println("Connected to MongoDB")

	return Client
}

// Get Collection -> return Mongo Collection
func GetCollection(collectionName string) *mongo.Collection {
	return Client.Database(config.AppConfig.DBName).Collection(collectionName)
}
