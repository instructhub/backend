package initialization

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/instructhub/backend/pkg/database"
	"github.com/instructhub/backend/pkg/encryption"
	"github.com/instructhub/backend/pkg/gitea"
	config "github.com/instructhub/backend/pkg/oauth"
	"github.com/instructhub/backend/pkg/s3"
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
	gt.InitGiteaClient()
	utils.InitVaildator()
	s3.ConnectS3()
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#18FD7BFF")).Render("Successfully initialized all necessary services"))
}
