package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/app/queries"
	"github.com/instructhub/backend/pkg/utils"
	"github.com/jinzhu/copier"
)

func GetProfile(c *gin.Context) {
	jwtContextID, exists := c.Get("userID")
	if !exists {
		utils.SimpleResponse(c, 403, "UserID not found in context", utils.ErrUserIDNotFound, nil)
		return
	}

	userID := jwtContextID.(uint64)

	if _, err := queries.GetUserQueueByID(userID); err != nil {
		utils.SimpleResponse(c, 403, "UserID error", utils.ErrGetData, nil)
		return
	}

	user, err := queries.GetUserQueueByID(userID)

	if err != nil {
		c.Error(err)
		utils.SimpleResponse(c, 500, "Internal server error while get user data", utils.ErrGetData, nil)
		return
	}

	var userProfile models.UserProfile
	err = copier.Copy(&userProfile, &user)
	if err != nil {
		c.Error(err)
		utils.SimpleResponse(c, 500, "Internal server error while process image", utils.ErrChangeType, nil)
		return
	}

	utils.SimpleResponse(c, 200, "User profile acquire", nil, userProfile)
}
