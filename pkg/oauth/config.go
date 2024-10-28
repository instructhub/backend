package config

import (
	"fmt"
	"log"
	"os"

	"github.com/instructhub/backend/pkg/utils"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gitlab"
	"github.com/markbates/goth/providers/google"
)

// Init oauth for goth
func OAuthInit() {
	err := godotenv.Load("template.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Change your url
	goth.UseProviders(
		google.New(
			os.Getenv("GOOGLE_CLIENT_ID"),
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			fmt.Sprintf("%s/auth/oauth/google/callback", utils.BaseURL),
			"email",
			"profile",
		),

		github.New(
			os.Getenv("GITHUB_CLIENT_ID"),
			os.Getenv("GITHUB_CLIENT_SECRET"),
			fmt.Sprintf("%s/auth/oauth/github/callback", utils.BaseURL),
		),

		gitlab.New(
			os.Getenv("GITLAB_CLIENT_ID"),
			os.Getenv("GITLAB_CLIENT_SECRET"),
			fmt.Sprintf("%s/auth/oauth/gitlab/callback", utils.BaseURL),
		),
	)
}
