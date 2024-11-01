package utils

import (
	"fmt"
	"os"
)

var (
	CookieRefreshTokenExpires int
	CookieAccessTokenExpires  int
	BaseURL                   string
	GiteaORGName              string
)

// Init some usefil variables
func InitVariables() {
	GiteaORGName = os.Getenv("GITEA_ORG_NAME")
	CookieRefreshTokenExpires = Atoi(os.Getenv("COOKIE_REFRESH_TOKEN_EXPIRES"))
	CookieAccessTokenExpires = Atoi(os.Getenv("COOKIE_ACCESS_TOKEN_EXPIRES"))
	BaseURL = fmt.Sprintf("%s/api/v%s", os.Getenv("BASE_URL"), os.Getenv("VERSION"))
}

func IsValidImageType(contentType string) bool {
	switch contentType {
	case "image/jpeg", "image/png", "image/gif":
		return true
	default:
		return false
	}
}
