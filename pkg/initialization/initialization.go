package initialization

import (
	"time"

	"github.com/instructhub/backend/pkg/database"
	"github.com/instructhub/backend/pkg/encryption"
	config "github.com/instructhub/backend/pkg/oauth"
	"github.com/instructhub/backend/pkg/utils"
	"golang.org/x/exp/rand"
)

// Init all need when server start
func Init() {
	database.InitMongoDB()
	database.SetupTTLIndex()
	encryption.InitSnowflake()
	utils.InitVariables()
	rand.Seed(uint64(time.Now().UnixNano()))
	config.OAuthInit()
}
