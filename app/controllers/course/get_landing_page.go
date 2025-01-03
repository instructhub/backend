package courses

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/app/queries"
	"github.com/instructhub/backend/pkg/utils"
)

// GetCourseLandingPageData handles fetching the landing page details of a course
func GetCourseLandingPageData(c *gin.Context) {
	courseID, err := utils.StrToUint64(c.Param("courseID"))
	if err != nil {
		utils.FullyResponse(c, http.StatusBadRequest, "Invalid course ID", utils.ErrBadRequest, nil)
		return
	}

	// Retrieve the landing page data for the course
	landingPage := models.CourseLandingPage{CourseID: courseID}
	landingPage, result := queries.GetCourseLandingPage(landingPage.CourseID)
	if result.Error != nil {
		if result.RowsAffected == 0 {
			utils.FullyResponse(c, http.StatusNotFound, "Landing page not found for the course", utils.ErrCourseNotExist, nil)
		} else {
			utils.ServerErrorResponse(c, http.StatusInternalServerError, "Error fetching landing page data", utils.ErrGetData, result.Error)
		}
		return
	}

	// Respond with the landing page data
	utils.FullyResponse(c, http.StatusOK, "Landing page data fetched successfully", nil, landingPage)
}
