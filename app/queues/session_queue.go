package queues

import (
	"context"

	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/pkg/database"
)

func CreateSessionQueue(session models.Session) error {
	_, err := database.GetCollection("sessions").InsertOne(context.Background(), session)
	return err
}
