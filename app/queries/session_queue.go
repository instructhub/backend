package queries

import (
	"context"

	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Create new session
func CreateSessionQueue(session models.Session) error {
	_, err := database.GetCollection("sessions").InsertOne(context.Background(), session)
	return err
}

// Get session
func GetSessionQueue(secretKey string) (models.Session, error) {
	session := models.Session{}
	err := database.GetCollection("sessions").FindOne(context.Background(), bson.M{"secret_key": secretKey}).Decode(&session)
	return session, err
}

// Get session
func DeleteSessionQueue(secretKey string) *mongo.SingleResult {
	err := database.GetCollection("sessions").FindOneAndDelete(context.Background(), bson.M{"secret_key": secretKey})
	return err
}

// Find and update session
func FindOneAndUpdateSession(secretKey string, used bool) (session models.Session, err error) {
	filter := bson.M{"secret_key": secretKey}
	update := bson.M{"$set": bson.M{"used": used}}

	result := database.GetCollection("sessions").FindOneAndUpdate(context.Background(), filter, update).Decode(&session)

	return session, result
}