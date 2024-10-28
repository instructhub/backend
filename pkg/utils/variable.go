package utils

import (
	"fmt"
	"os"
)

var (
	CookieRefreshTokenExpires int
	CookieAccessTokenExpires  int
	BaseURL                   string
)

// Init some usefil variables
func InitVariables() {
	CookieRefreshTokenExpires = Atoi(os.Getenv("COOKIE_REFRESH_TOKEN_EXPIRES"))
	CookieAccessTokenExpires = Atoi(os.Getenv("COOKIE_ACCESS_TOKEN_EXPIRES"))
	BaseURL = fmt.Sprintf("%s/api/v%s", os.Getenv("BASE_URL"), os.Getenv("VERSION"))
}
