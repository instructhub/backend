package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/controllers"
)

func AuthRoute(r *gin.Engine) {
	auth := r.Group("/auth")

	auth.POST("/login", controllers.Login)
}
