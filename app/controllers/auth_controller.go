package controllers

import (
	"github.com/instructhub/backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	utils.SimpleResponse(c, 200, "Login successful", nil)
}
