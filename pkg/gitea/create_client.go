package gt

import (
	"fmt"
	"log"
	"os"

	"code.gitea.io/sdk/gitea"
)

var GiteaClient *gitea.Client

func InitGiteaClient() {
    client, err := gitea.NewClient(os.Getenv("GITEA_URL"), gitea.SetToken(os.Getenv("GITEA_TOKEN")))
    if err != nil {
		log.Fatalln("error connect to gitea")
	}
    GiteaClient = client

	fmt.Println("Succesful connect to gitea")
}
