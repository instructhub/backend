package utils

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/pkg/encryption"
	"github.com/instructhub/backend/app/queues"
)

func GenerateUserSession(c *gin.Context, userID uint64) error {
	session := models.Session{
		SecretKey: encryption.RandStringRunes(256),
		UserID:    userID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * time.Duration(CookieRefreshTokenExpires)),
		CreatedAt: time.Now(),
	}

	err := queues.CreateSessionQueue(session)
	if err != nil {
		return err
	}

	accessTokenExpiresAt := time.Now().Add(time.Minute * time.Duration(CookieAccessTokenExpires))
	accessToken, err := encryption.GenerateNewJwtToken(userID, []string{}, accessTokenExpiresAt)
	if err != nil {
		return err
	}

	c.SetCookie("refresh_token", session.SecretKey, int(session.ExpiresAt.Unix()), "/refresh", "", false, true)
	c.SetCookie("access_token", accessToken, int(accessTokenExpiresAt.Unix()), "/", "", false, true)

	return nil
}
