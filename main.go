package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/routes"
	"github.com/instructhub/backend/middleware"
	"github.com/instructhub/backend/pkg/initialization"
	"github.com/instructhub/backend/pkg/utils"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	r := gin.New()

	// Init all dependencies
	initialization.Init()

	utils.PrintAppBanner()

	routes.AuthRoute(r)

	r.StaticFile("/favicon.ico", "./static/favicon.ico")
	// use custom logger
	r.Use(middleware.CustomLogger())

	if err := r.Run(); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
