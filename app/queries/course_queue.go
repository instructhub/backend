package queries

import (
	"github.com/instructhub/backend/app/models"
	db "github.com/instructhub/backend/pkg/database"
	"gorm.io/gorm"
)

// Create new course
func CreateNewCourse(course models.Course) *gorm.DB {
	// Insert a new course into the database
	result := db.GetDB().Create(&course)
	return result
}

// Get course information by courseID
func GetCourseInformation(courseID uint64) (models.Course, *gorm.DB) {
	var course models.Course
	// Query the course based on courseID
	result := db.GetDB().Where("id = ?", courseID).First(&course)
	return course, result
}

// Create course image
func CreateCourseImage(image models.CourseImage) *gorm.DB {
	// Insert a new course image into the database
	result := db.GetDB().Create(&image)
	return result
}