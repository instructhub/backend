package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/queries"
	"github.com/instructhub/backend/pkg/utils"
)

func GetProfile(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil{
		utils.SimpleResponse(c, 403, "Invalid datatype", nil)
		return
	}

	jwtContextID, exists := c.Get("userID")
	if !exists {
		utils.SimpleResponse(c, 403, "UserID not found in context", nil)
		return
	}

	jwtContextIDUint := jwtContextID.(uint64)

	if _, err := queries.GetUserQueueByID(userID); err != nil || jwtContextIDUint != userID {
		utils.SimpleResponse(c, 403, "UserID error", nil)
		return
	}

	user, err := queries.GetUserQueueByID(userID)

	if err != nil {
		utils.SimpleResponse(c, 200, "UserID Error", err)
		return
	}

	utils.SimpleResponse(c, 200, "User profile acquire", user)
}
