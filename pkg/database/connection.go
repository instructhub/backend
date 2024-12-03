package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/instructhub/backend/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB Client
var MongoClient *mongo.Client

// InitMongoDB Init MongoDB connection
func init() {
	// Set connection options
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))

	// Set a context to avoid long blocking
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.Log.Sugar().Fatalf("Failed to connect to MongoDB:", err)
	}

	// Check connection
	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Log.Sugar().Fatal("Failed to connect to MongoDB:", err)
	}

	MongoClient = client

	// Setup ttl index to auto delete
	setupTTLIndex()
}

// GetCollection Get MongoDB collection
func GetCollection(collectionName string) *mongo.Collection {
	return MongoClient.Database(os.Getenv("DB_NAME")).Collection(collectionName)
}

// CloseMongoDB Close MongoDB connection
func CloseMongoDB() {
	if err := MongoClient.Disconnect(context.TODO()); err != nil {
		logger.Log.Sugar().Fatal("Failed to disconnect from MongoDB:", err)
	}
	fmt.Println("Successfully disconnected from MongoDB")
}
