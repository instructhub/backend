package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/controllers"
	"github.com/instructhub/backend/app/middleware"
	testing "github.com/instructhub/backend/test"
)

func AuthRoute(r *gin.Engine) {
	test := r.Group("/test")
	test.Use(middleware.IsAuthorized())
	test.GET("/authorize", testing.Handler)

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
