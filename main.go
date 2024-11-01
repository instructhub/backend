package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/routes"
	"github.com/instructhub/backend/pkg/initialization"
	"github.com/instructhub/backend/pkg/middleware"
	"github.com/instructhub/backend/pkg/utils"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	root := gin.New()

	root.SetTrustedProxies([]string{"127.0.0.1"})
	root.StaticFile("/favicon.ico", "./static/favicon.ico")
	root.Use(middleware.CustomLogger())

	// Init all dependencies
	initialization.Init()

	r := root.Group("/api/v" + os.Getenv("VERSION"))

	utils.PrintAppBanner()

	routes.AuthRoute(r)
	routes.UserRoute(r)
	routes.CourseRoute(r)

	if err := root.Run(); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
