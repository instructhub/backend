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

func GetCourseWithDetails(courseID uint64) (models.Course, *gorm.DB) {
	var course models.Course

	result := db.GetDB().Model(&models.Course{ID: courseID}).Preload("CourseStages.CourseItems").First(&course)
	return course, result
}

// Create course stages
func CreateCourseStages(stages []models.CourseStage) *gorm.DB {
	result := db.GetDB().Create(&stages)
	return result
}

// Create course items
func CreateCourseItems(items []models.CourseItem) *gorm.DB {
	result := db.GetDB().Create(&items)
	return result
}

// Update course stages
func UpdateCourseStage(stage models.CourseStage) *gorm.DB {
	result := db.GetDB().Model(&stage).Where("id = ?", stage.ID).Updates(&stage)
	return result
}

// Update course stages
func UpdateCourseItem(item models.CourseItem) *gorm.DB {
	result := db.GetDB().Model(&item).Where("id = ?", item.ID).Updates(&item)
	return result
}

// Craete course history
func CreateNewCourseHistory(revision models.CourseRevision) *gorm.DB {
	result := db.GetDB().Create(&revision)
	return result
}

func GetCourseRevision(courseID uint64, revisionID uint64) (models.CourseRevision, *gorm.DB) {
	var courseRevision models.CourseRevision

	result := db.GetDB().Model(&models.CourseRevision{ID: revisionID, CourseID: courseID}).
		Preload("Course").
		Preload("Course.CourseStages.CourseItems").
		First(&courseRevision)
	return courseRevision, result
}
