package encryption

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Generate new jwt token with credentials
func GenerateNewJwtToken(id uint64, credentials []string, expiresAt time.Time) (string, error) {
	// Set secret key from .env file.
	secret := os.Getenv("JWT_SECRET_KEY")

	// Set expires minutes count for secret key from .env file.

	// Create a new claims.
	claims := jwt.MapClaims{}

	// Set public claims:
	claims["id"] = id
	claims["sub"] = time.Now().Unix()
	claims["expires"] = expiresAt.Unix()

	// Set private token credentials:
	for _, credential := range credentials {
		claims[credential] = true
	}

	// Create a new JWT access token with claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate token.
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		// Return error, it JWT token generation failed.
		return "", err
	}

	return t, nil
}
