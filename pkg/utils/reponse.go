package utils

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SimpleResponse simplifies the gin response process and handles errors appropriately.
func SimpleResponse(c *gin.Context, statusCode int, message string, errorCode interface{}, data interface{}) {
	// Check if the status code indicates a server error
	if statusCode >= 500 && statusCode < 600 {
		// Log the error with structured fields
		Logger.With(
			zap.Int("status", statusCode),
			zap.String("message", message),
			zap.Any("data", data),
			zap.String("error_code", fmt.Sprintf("%v", errorCode)),
		).Error("Server error occurred")
		data = nil
	}

	// Prepare the response payload
	response := gin.H{"message": message}

	if data != nil {
		response["result"] = data
	}

	// If it's an error, include an "error" field
	var errorCodePtr *string

	switch v := errorCode.(type) {
	case string:
		errorCodePtr = &v
	case *string:
		errorCodePtr = v
	case nil:
		errorCodePtr = nil
	default:
		panic("invalid errorCode type")
	}

	if errorCodePtr != nil {
		response["error"] = *errorCodePtr
	}

	// Send JSON response
	c.JSON(statusCode, response)
}
