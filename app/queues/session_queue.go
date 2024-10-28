package queues

import (
	"context"

	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
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
