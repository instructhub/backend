package courses

import (
	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/queries"
	"github.com/instructhub/backend/pkg/utils"
)

// GetCourse handles get ccourse data (only stage and item no item content)
func GetCourse(c *gin.Context) {
	// Parse course ID
	courseID, err := utils.StrToUint64(c.Param("courseID"))
	if err != nil {
		utils.FullyResponse(c, 400, "Error getting course ID", utils.ErrBadRequest, nil)
		return
	}

	// Fetch course data
	courseData, result := queries.GetCourseWithDetails(courseID)
	if result.RowsAffected == 0 {
		utils.FullyResponse(c, 404, "Course not exist", utils.ErrCourseNotExist, nil)
		return
	} else if result.Error != nil {
		c.Error(result.Error)
		utils.FullyResponse(c, 500, "Error fetching course", utils.ErrGetData, nil)
		return
	}

	utils.FullyResponse(c, 200, "Successfully get course data", nil, courseData)
}
