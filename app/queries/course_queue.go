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
		Preload("CourseModules", activeFilter, orderByPosition).
		Preload("CourseModules.CourseSteps", activeFilter, orderByPosition).
		First(&course, courseID)

	return course, result
}

// Create course modules
func CreateCourseModules(modules []models.CourseModule) *gorm.DB {
	result := db.GetDB().Create(&modules)
	return result
}

// Create course steps
func CreateCourseSteps(steps []models.CourseStep) *gorm.DB {
	result := db.GetDB().Create(&steps)
	return result
}

// Update course modules
func UpdateCourseModule(module models.CourseModule) *gorm.DB {
	result := db.GetDB().Model(&module).Where("id = ?", module.ID).Updates(&module)
	return result
}

// Update course modules
func UpdateCourseStep(step models.CourseStep) *gorm.DB {
	result := db.GetDB().Model(&step).Where("id = ?", step.ID).Updates(&step)
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

	// Use JOIN to fetch course revision with related course modules and course steps
	result := db.GetDB().
		Preload("Course").
		Preload("Course.CourseModules", activeFilter).
		Preload("Course.CourseModules.CourseSteps", activeFilter).
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

// Get course landing page data
func GetCourseLandingPage(courseID uint64) (models.CourseLandingPage, *gorm.DB) {
	var landingPage models.CourseLandingPage
	result := db.GetDB().Where("course_id = ?", courseID).First(&landingPage)
	return landingPage, result
}

// Create course landing page
func CreateCourseLandingPage(landingPage models.CourseLandingPage) *gorm.DB {
	result := db.GetDB().Create(&landingPage)
	return result
}

// Update course revision
func UpdateCourseLandingPage(landingPage models.CourseLandingPage) *gorm.DB {
	result := db.GetDB().Model(&landingPage).Where("course_id = ?", landingPage.CourseID).Updates(&landingPage)
	return result
}
