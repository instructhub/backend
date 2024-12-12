package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/routes"
	_ "github.com/instructhub/backend/pkg/cache"
	_ "github.com/instructhub/backend/pkg/database"
	"github.com/instructhub/backend/pkg/logger"
	"github.com/instructhub/backend/pkg/middleware"
	_ "github.com/instructhub/backend/pkg/oauth"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	root := gin.New()

	root.SetTrustedProxies([]string{"127.0.0.1"})
	root.StaticFile("/favicon.ico", "./static/favicon.ico")
	root.Use(middleware.CustomLogger())
	root.Use(middleware.ErrorLoggerMiddleware())
	root.LoadHTMLGlob("template/*")
	// Init all dependencies

	r := root.Group("/api/v" + os.Getenv("VERSION"))

	route(r)

	printAppInfo()

	root.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error":   "resource_not_found",
			"message": "Resource not found",
		})
	})

	if err := root.Run(); err != nil {
		logger.Log.Sugar().Fatal("Server failed to start: %v", err)
	}
}

func printAppInfo() {
	info := fmt.Sprintf(`
	InstructHub API
	Version: %s
	Gin Version: %s
	Domain: %s
	`, os.Getenv("VERSION"), gin.Version, os.Getenv("BASE_URL"))
	logger.Log.Info(info)
}

func route(r *gin.RouterGroup) {
	routes.AuthRoute(r)
	routes.UserRoute(r)
	routes.CourseRoute(r)
	routes.PublicRouter(r)
}
