package utils

import "os"

var (
	CookieRefreshTokenExpires int
	CookieAccessTokenExpires  int
)

func InitVariables() {
	CookieRefreshTokenExpires = Atoi(os.Getenv("COOKIE_REFRESH_TOKEN_EXPIRES"))
	CookieAccessTokenExpires = Atoi(os.Getenv("COOKIE_ACCESS_TOKEN_EXPIRES"))
}
