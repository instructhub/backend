package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/controllers"
	"github.com/instructhub/backend/app/middleware"
)

func AuthRoute(r *gin.RouterGroup) {
	user := r.Group("/user")
	user.Use(middleware.IsAuthorized())

	user.GET("/allprofile/:id", controllers.GetProfile) // Admin use only

	auth := r.Group("/auth")

	auth.POST("/signup", controllers.Signup)
	auth.POST("/login", controllers.Login)

	oauth := auth.Group("/oauth")

	// The provider is hard-coded to prevent attacks, not sure if it works though :p
	oauth.GET("/google", func(c *gin.Context) { controllers.OAuthHandler(c, "google") })
	oauth.GET("/google/callback", func(c *gin.Context) { controllers.OAuthCallbackHandler(c, "google") })

	oauth.GET("/github", func(c *gin.Context) { controllers.OAuthHandler(c, "github") })
	oauth.GET("/github/callback", func(c *gin.Context) { controllers.OAuthCallbackHandler(c, "github") })

	oauth.GET("/gitlab", func(c *gin.Context) { controllers.OAuthHandler(c, "gitlab") })
	oauth.GET("/gitlab/callback", func(c *gin.Context) { controllers.OAuthCallbackHandler(c, "gitlab") })
}
