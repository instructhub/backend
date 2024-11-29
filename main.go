package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/routes"
	_ "github.com/instructhub/backend/pkg/cache"
	"github.com/instructhub/backend/pkg/middleware"
	_ "github.com/instructhub/backend/pkg/oauth"
	"github.com/instructhub/backend/pkg/utils"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	root := gin.New()

	root.SetTrustedProxies([]string{"127.0.0.1"})
	root.StaticFile("/favicon.ico", "./static/favicon.ico")
	root.Use(middleware.CustomLogger())
	root.LoadHTMLGlob("template/*")
	// Init all dependencies

	r := root.Group("/api/v" + os.Getenv("VERSION"))

	utils.PrintAppBanner()

	routes.AuthRoute(r)
	routes.UserRoute(r)
	routes.CourseRoute(r)
	routes.PublicRouter(r)

	root.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error":   "resource_not_found",
			"message": "Resource not found",
		})
	})

	if err := root.Run(); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
