package middleware

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/instructhub/backend/pkg/utils"
)

func IsAuthorized() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie("access_token")

		if err != nil || cookie.Value == "" {
			utils.SimpleResponse(c, 403, "Authorization token is empty.", "authentication_key_not_found", nil)
			c.Abort()
			return
		}

		parseError := false

		secret := os.Getenv("JWT_SECRET_KEY")
		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			// Ensure the signing method is as expected.
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				parseError = true
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid || parseError {
			utils.SimpleResponse(c, 403, "Unauthorized", "unauthorized", nil)
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.SimpleResponse(c, 403, "Unauthorized", "unauthorized", nil)
			c.Abort()
			return
		}
		expiresAtFloat, ok := claims["expires"].(float64)

		if !ok {
			utils.SimpleResponse(c, 403, "Invalid expries datatype", "unauthorized", nil)
			c.Abort()
			return
		}

		expiresAt := int64(expiresAtFloat)

		if time.Now().Unix() >= expiresAt {
			utils.SimpleResponse(c, 403, "Token expired", "token_expired", nil)
			c.Abort()
			return
		}

		userIDFloat, ok := claims["sub"].(float64)
		if !ok {
			utils.SimpleResponse(c, 403, "UserID error", "unauthorized", nil)
			c.Abort()
			return
		}
		userID := uint64(userIDFloat)

		c.Set("userID", userID)
		c.Next()
	}
}
