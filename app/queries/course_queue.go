package queries

import (
	"context"

	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/pkg/database"
)

func CraeteNewCourse(course models.Course) error {
	_, err := database.GetCollection("course").InsertOne(context.Background(), course)
	return err
}
