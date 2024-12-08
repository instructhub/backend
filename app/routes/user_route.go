package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/controllers"
	"github.com/instructhub/backend/pkg/middleware"
)

func UserRoute(r *gin.RouterGroup) {
	user := r.Group("/users")
	user.Use(middleware.IsAuthorized())

	// Cheeck for login or not
	user.GET("/login/check", controllers.CheckLogin) 
	// Get user personal profile
	user.GET("/personal/profile", controllers.GetProfile)
}
