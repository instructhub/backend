package utils

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateRandomBase64String() (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	state := base64.URLEncoding.EncodeToString(randomBytes)
	return state, nil
}
