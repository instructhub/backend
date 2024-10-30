package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/middleware"
)

func UserRoute(r *gin.RouterGroup) {
	user := r.Group("/user")
	user.Use(middleware.IsAuthorized())
}
