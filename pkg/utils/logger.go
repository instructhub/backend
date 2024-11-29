package utils

import (
	"os"

	"go.uber.org/zap"
)

var Logger *zap.Logger

func init() {
	Logger = zap.Must(zap.NewProduction())
	if os.Getenv("GIN_MODE") == "debug" {
		Logger = zap.Must(zap.NewDevelopment())
	}
}
