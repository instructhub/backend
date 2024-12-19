package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/app/queries"
	"github.com/instructhub/backend/pkg/utils"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

func CheckLogin(c *gin.Context) {
	utils.FullyResponse(c, 200, "User already login", nil, nil)
}

func GetProfile(c *gin.Context) {
	jwtContextID, exists := c.Get("userID")
	if !exists {
		utils.FullyResponse(c, 403, "UserID not found in context", utils.ErrUserIDNotFound, nil)
		return
	}

	userID := jwtContextID.(uint64)

	user, result := queries.GetUserQueueByID(userID)
	if result.Error == gorm.ErrRecordNotFound {
		utils.FullyResponse(c, 403, "UserID error", utils.ErrGetData, nil)
		return
	} else if result.Error != nil {
		utils.ServerErrorResponse(c, 500, "Error get user data", utils.ErrGetData, result.Error)
		return
	}

	var userProfile models.UserProfile
	err := copier.Copy(&userProfile, &user)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error process image", utils.ErrChangeType, err)
		return
	}

	utils.FullyResponse(c, 200, "User profile acquire", nil, userProfile)
}
