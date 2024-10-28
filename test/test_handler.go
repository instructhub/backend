package testing

import (
	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/pkg/utils"
)

func Handler(c *gin.Context) {
	utils.SimpleResponse(c, 200, "Successfully in", nil)
}
