package queries

import (
	"context"

	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
)

// Create new course
func CraeteNewCourse(course models.Course) error {
	_, err := database.GetCollection("course").InsertOne(context.Background(), course)
	return err
}

// Get course information
func GetCourseInformation(courseID uint64) (models.Course, error) {
	course := models.Course{}
	err := database.GetCollection("course").FindOne(context.Background(), bson.M{"course_id": courseID}).Decode(&course)
	return course, err
}

// Create image
func CraeteCourseImage(image models.CourseImage) error {
	_, err := database.GetCollection("course_image").InsertOne(context.Background(), image)
	return err
}
