package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/controllers"
	"github.com/instructhub/backend/pkg/middleware"
)

func UserRoute(r *gin.RouterGroup) {
	user := r.Group("/user")
	user.Use(middleware.IsAuthorized())

	user.GET("/allprofile/:id", controllers.GetProfile) // Admin use only
}
