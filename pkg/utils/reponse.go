package utils

import (
	"log"

	"github.com/gin-gonic/gin"
)

// For simplize gin response
func SimpleResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	if (statusCode > 499 && statusCode < 600) {
		log.Println(message, data)
	}
	c.JSON(statusCode, gin.H{"message": message, "data": data})
}
