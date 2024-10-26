package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/github"
	// "github.com/markbates/goth/providers/gitlab"
	"github.com/markbates/goth/providers/google"
)

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
			"http://localhost:8080/auth/oauth/google/callback",
			"email",
			"profile",
		),

		github.New(
			os.Getenv("GITHUB_CLIENT_ID"),
			os.Getenv("GITHUB_CLIENT_SECRET"),
			"http://localhost:8080/auth/oauth/github/callback",
		),

		// gitlab.New(
		// 	os.Getenv("GITLAB_KEY"),
		// 	os.Getenv("GITLAB_SECRET"),
		// 	"http://localhost:3000/auth/gitlab/callback",
		// ),
	)
}
