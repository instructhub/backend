package courses

import (
	"github.com/gin-gonic/gin"
	git "github.com/instructhub/backend/pkg/gitea"
	"github.com/instructhub/backend/pkg/utils"
)

// GetItemContent handles get ccourse items content data
func GetItemContent(c *gin.Context) {
	// Parse course ID and item ID
	courseID, err := utils.StrToUint64(c.Param("courseID"))
	if err != nil {
		utils.SimpleResponse(c, 400, "Error getting course ID", utils.ErrBadRequest, nil)
		return
	}

	itemID, err := utils.StrToUint64(c.Param("itemID"))
	if err != nil {
		utils.SimpleResponse(c, 400, "Error getting item ID", utils.ErrBadRequest, nil)
		return
	}

	itemContent, _, err := git.GiteaClient.GetContents(utils.GiteaORGName, utils.Uint64ToStr(courseID), "", utils.Uint64ToStr(itemID))
	if err != nil {
		utils.SimpleResponse(c, 404, "Course or item not exist", utils.ErrCourseNotExist, nil)
		return
	}

	utils.SimpleResponse(c, 200, "Successfully get course item content", nil, itemContent.Content)
}
