package initialization

import (
	"time"

	"github.com/instructhub/backend/pkg/database"
	config "github.com/instructhub/backend/pkg/oauth"
	"github.com/instructhub/backend/pkg/utils"
	"golang.org/x/exp/rand"
)

func Init() {
	database.InitMongoDB()
	utils.InitSnowflake()
	utils.InitVariables()
	rand.Seed(uint64(time.Now().UnixNano()))
	config.OAuthInit()
}
