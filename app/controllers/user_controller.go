package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/queues"
	"github.com/instructhub/backend/pkg/utils"
)

func GetProfile(c *gin.Context) {
	contextID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	user, err := queues.GetUserQueueByID(contextID)

	if err != nil {
		utils.SimpleResponse(c, 200, "UserID Error", err)
		return
	}

	utils.SimpleResponse(c, 200, "User profile acquire", user)
}
