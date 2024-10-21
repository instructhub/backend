package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/routes"
	"github.com/instructhub/backend/middleware"
	"github.com/instructhub/backend/pkg/database"
	"github.com/instructhub/backend/pkg/utils"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	r := gin.New()

	// Init MongoDB connection
	database.InitMongoDB()

	// use custom logger
	r.Use(middleware.CustomLogger())

	utils.PrintAppBanner()

	routes.AuthRoute(r)

	if err := r.Run(); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
