package middleware

import (
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/instructhub/backend/app/queues"
)

func IsAuthorized() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie("access_token")

		if err != nil || cookie.Value == "" {
			c.JSON(403, gin.H{
				"msg": "Authorization token is empty.",
			})
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
			c.JSON(403, gin.H{
				"msg": "Unauthorized",
			})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			expiresAtFloat, ok := claims["expires"].(float64)

			if !ok {
				c.JSON(403, gin.H{
					"msg": "Invalid datatype",
				})
				c.Abort()
				return
			}

			expiresAt := int64(expiresAtFloat)

			if time.Now().Unix() >= expiresAt {
				c.JSON(403, gin.H{
					"msg": "Token expired",
				})
				c.Abort()
				return
			}

			userIDFloat, ok := claims["sub"].(float64)
			contextID, err := strconv.ParseUint(c.Param("id"), 10, 64)

			if !ok || err != nil {
				c.JSON(403, gin.H{
					"msg": "Invalid datatype",
				})
				c.Abort()
				return
			}

			userID := uint64(userIDFloat)

			if _, err := queues.GetUserQueueByID(userID); err != nil || contextID != userID {
				c.JSON(403, gin.H{
					"msg": "UserID Error",
				})
				c.Abort()
				return
			}
		} else {
			c.JSON(403, gin.H{
				"msg": "Unauthorized",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
