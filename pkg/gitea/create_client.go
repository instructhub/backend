package gt

import (
	"os"

	"code.gitea.io/sdk/gitea"
	"github.com/instructhub/backend/pkg/logger"
)

var GiteaClient *gitea.Client

func init() {
    client, err := gitea.NewClient(os.Getenv("GITEA_URL"), gitea.SetToken(os.Getenv("GITEA_TOKEN")))
    if err != nil {
		logger.Log.Fatal("error connect to gitea")
	}
    GiteaClient = client
}
