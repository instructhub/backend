package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/controllers"
)

func AuthRoute(r *gin.Engine) {
	auth := r.Group("/auth")

	auth.POST("/signup", controllers.Signup)
	auth.POST("/login", controllers.Login)

  oauth := auth.Group("/oauth")
  oauth.GET("/google", controllers.GoogleOAuthHandler)
  oauth.GET("/google/callback", controllers.GoogleOAuthCallbackHandler)
}
