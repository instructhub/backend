package utils

import (
	"bytes"
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
func init() {
	GiteaORGName = os.Getenv("GITEA_ORG_NAME")
	CookieRefreshTokenExpires = Atoi(os.Getenv("COOKIE_REFRESH_TOKEN_EXPIRES"))
	CookieAccessTokenExpires = Atoi(os.Getenv("COOKIE_ACCESS_TOKEN_EXPIRES"))
	BaseURL = fmt.Sprintf("%s/api/v%s", os.Getenv("BASE_URL"), os.Getenv("VERSION"))
}

// Magic bytes for different image formats
var magicBytes = map[string][]byte{
	"image/jpeg": {0xFF, 0xD8},
	"image/png":  {0x89, 0x50, 0x4E, 0x47},
	"image/gif":  {0x47, 0x49, 0x46},
}

// Check if the uploaded file is an image using Magic Bytes
func IsValidImageType(magic []byte) (bool, string, error) {
	// Compare the bytes with known magic bytes
	for mimeType, magicPattern := range magicBytes {
		if bytes.HasPrefix(magic, magicPattern) {
			return true, mimeType, nil
		}
	}

	return false, "", fmt.Errorf("uploaded file is not a valid image")
}
