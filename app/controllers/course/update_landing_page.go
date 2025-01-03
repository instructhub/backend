package courses

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/app/queries"
	"github.com/instructhub/backend/pkg/utils"
	"github.com/lib/pq"
)

type courseLandingPageRequest struct {
	Description    *string         `json:"description,omitempty" binding:"required"`
	ImageURL       *string         `json:"image_url,omitempty" binding:"omitempty,http_url,max=256"`
	VideoURL       *string         `json:"video_url,omitempty" binding:"omitempty,http_url,max=256"`
	SEOKeywords    *pq.StringArray `json:"seo_keywords,omitempty" binding:"omitempty,dive,max=10"`
	Outcomes       *pq.StringArray `json:"outcomes,omitempty" binding:"omitempty,dive,max=10"`
	Prerequisites  *pq.StringArray `json:"prerequisites,omitempty" binding:"omitempty,dive,max=10"`
	TargetAudience *pq.StringArray `json:"target_audience,omitempty" binding:"omitempty,dive,max=10"`
}

// UpdateCourseLandingPage handles updating course landing page details
func UpdateCourseLandingPage(c *gin.Context) {
	courseID, err := utils.StrToUint64(c.Param("courseID"))
	if err != nil {
		utils.FullyResponse(c, http.StatusBadRequest, "Invalid course ID", utils.ErrBadRequest, nil)
		return
	}

	if _, result := queries.GetCourseWithDetails(courseID); result.Error != nil {
		if result.RowsAffected == 0 {
			utils.FullyResponse(c, http.StatusNotFound, "Course not found", utils.ErrCourseNotExist, nil)
		} else {
			utils.ServerErrorResponse(c, http.StatusInternalServerError, "Error fetching course", utils.ErrGetData, result.Error)
		}
		return
	}

	var request courseLandingPageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.FullyResponse(c, http.StatusBadRequest, "Invalid request body", utils.ErrBadRequest, err.Error())
		return
	}

	landingPage := models.CourseLandingPage{CourseID: courseID}
	if _, result := queries.GetCourseLandingPage(landingPage.CourseID); result.Error != nil {
		createLandingPage(c, courseID, request)
		return
	}

	updateLandingPage(c, landingPage, request)
}

func createLandingPage(c *gin.Context, courseID uint64, request courseLandingPageRequest) {
	landingPage := models.CourseLandingPage{
		CourseID:       courseID,
		Description:    request.Description,
		ImageURL:       request.ImageURL,
		VideoURL:       request.VideoURL,
		SEOKeywords:    request.SEOKeywords,
		Outcomes:       request.Outcomes,
		Prerequisites:  request.Prerequisites,
		TargetAudience: request.TargetAudience,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := queries.CreateCourseLandingPage(landingPage).Error; err != nil {
		utils.ServerErrorResponse(c, http.StatusInternalServerError, "Error creating landing page", utils.ErrSaveData, err)
		return
	}
	utils.FullyResponse(c, http.StatusCreated, "Landing page created successfully", nil, landingPage)
}

func updateLandingPage(c *gin.Context, landingPage models.CourseLandingPage, request courseLandingPageRequest) {
	landingPage.Description = request.Description
	landingPage.ImageURL = request.ImageURL
	landingPage.VideoURL = request.VideoURL
	landingPage.SEOKeywords = request.SEOKeywords
	landingPage.Outcomes = request.Outcomes
	landingPage.Prerequisites = request.Prerequisites
	landingPage.TargetAudience = request.TargetAudience
	landingPage.UpdatedAt = time.Now()

	if err := queries.UpdateCourseLandingPage(landingPage).Error; err != nil {
		utils.ServerErrorResponse(c, http.StatusInternalServerError, "Error updating landing page", utils.ErrSaveData, err)
		return
	}
	utils.FullyResponse(c, http.StatusOK, "Landing page updated successfully", nil, landingPage)
}
