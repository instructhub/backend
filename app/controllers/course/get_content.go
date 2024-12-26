package courses

import (
	"github.com/gin-gonic/gin"
	git "github.com/instructhub/backend/pkg/gitea"
	"github.com/instructhub/backend/pkg/utils"
)

// GetStepContent handles get ccourse steps content data
func GetStepContent(c *gin.Context) {
	// Parse course ID and step ID
	courseID, err := utils.StrToUint64(c.Param("courseID"))
	if err != nil {
		utils.FullyResponse(c, 400, "Error getting course ID", utils.ErrBadRequest, nil)
		return
	}

	stepID, err := utils.StrToUint64(c.Param("stepID"))
	if err != nil {
		utils.FullyResponse(c, 400, "Error getting step ID", utils.ErrBadRequest, nil)
		return
	}

	stepContent, _, err := git.GiteaClient.GetContents(utils.GiteaORGName, utils.Uint64ToStr(courseID), "", utils.Uint64ToStr(stepID))
	if err != nil {
		utils.FullyResponse(c, 404, "Course or step not exist", utils.ErrCourseNotExist, nil)
		return
	}

	utils.FullyResponse(c, 200, "Successfully get course step content", nil, stepContent.Content)
}
