package main

import (
	"fmt"
	"temp/middleware"
	"temp/pkg/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	r := gin.New()

	// use custom logger
	r.Use(middleware.CustomLogger())

	// test route
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello, World!")
	})

	utils.PrintAppBanner()

	if err := r.Run(); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
