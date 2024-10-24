package initialization

import (
	"github.com/instructhub/backend/pkg/database"
	"github.com/instructhub/backend/pkg/utils"
)

func Init() {
	database.InitMongoDB()
	utils.InitSnowflake()
}
