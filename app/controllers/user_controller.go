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
	utils.SimpleResponse(c, 200, "User already login", nil, nil)
}

func GetProfile(c *gin.Context) {
	jwtContextID, exists := c.Get("userID")
	if !exists {
		utils.SimpleResponse(c, 403, "UserID not found in context", utils.ErrUserIDNotFound, nil)
		return
	}

	userID := jwtContextID.(uint64)

	user, result := queries.GetUserQueueByID(userID); 
	if result.Error == gorm.ErrRecordNotFound {
		utils.SimpleResponse(c, 403, "UserID error", utils.ErrGetData, nil)
		return
	} else if result.Error != nil {
		c.Error(result.Error)
		utils.SimpleResponse(c, 500, "Internal server error while get user data", utils.ErrGetData, nil)
		return
	}

	var userProfile models.UserProfile
	err := copier.Copy(&userProfile, &user)
	if err != nil {
		c.Error(err)
		utils.SimpleResponse(c, 500, "Internal server error while process image", utils.ErrChangeType, nil)
		return
	}

	utils.SimpleResponse(c, 200, "User profile acquire", nil, userProfile)
}

