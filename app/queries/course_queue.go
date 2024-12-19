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

// Reusable active filter function
func activeFilter(db *gorm.DB) *gorm.DB {
    return db.Where("active = ?", true)
}

func orderByPosition(db *gorm.DB) *gorm.DB {
    return db.Order("position")
}

func GetCourseWithDetails(courseID uint64) (models.Course, *gorm.DB) {
    var course models.Course

    result := db.GetDB().
        Preload("CourseStages", activeFilter, orderByPosition).
        Preload("CourseStages.CourseItems", activeFilter, orderByPosition).
        First(&course, courseID)

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

// Craete course revision
func CreateNewCourseRevision(revision models.CourseRevision) *gorm.DB {
	result := db.GetDB().Create(&revision)
	return result
}

// Get course revision with JOINs
func GetCourseRevision(courseID uint64, revisionID uint64) (models.CourseRevision, *gorm.DB) {
    var courseRevision models.CourseRevision

    // Use JOIN to fetch course revision with related course stages and course items
    result := db.GetDB().
        Preload("Course").
        Preload("Course.CourseStages", activeFilter).
        Preload("Course.CourseStages.CourseItems", activeFilter).
        Where("course_id = ?", courseID).
        Where("id = ?", revisionID).
        First(&courseRevision)

    return courseRevision, result
}

// Update course revision
func UpdateCourseRevision(revision models.CourseRevision) *gorm.DB {
	result := db.GetDB().Save(&revision)
	return result
}
