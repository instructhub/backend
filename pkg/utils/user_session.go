package utils

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/app/queries"
	"github.com/instructhub/backend/pkg/encryption"
	"go.mongodb.org/mongo-driver/mongo"
)

// Generate new user access_token and refresh_token
func GenerateUserSession(c *gin.Context, userID uint64) error {
	var err error
	secretKey, err := encryption.RandStringRunes(256)
	if err != nil {
		return err
	}
	session := models.Session{
		SessionID: encryption.GenerateID(),
		SecretKey: secretKey,
		UserID:    userID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * time.Duration(CookieRefreshTokenExpires)),
		CreatedAt: time.Now(),
	}

	for {
		_, err = queries.GetSessionQueue(session.SecretKey)
		if err == mongo.ErrNoDocuments {
			break
		} else if err != nil {
			fmt.Println(err.Error())
			return err
		}
	}

	err = queries.CreateSessionQueue(session)
	if err != nil {
		return err
	}

	accessTokenExpiresAt := time.Now().Add(time.Minute * time.Duration(CookieAccessTokenExpires))
	accessToken, err := encryption.GenerateNewJwtToken(userID, []string{}, accessTokenExpiresAt)
	if err != nil {
		return err
	}

	c.SetCookie("refresh_token", session.SecretKey, CookieRefreshTokenExpires * 24 * 60 * 60, BackendURL + "/auth/refresh/", "", false, true)
	c.SetCookie("access_token", accessToken, CookieAccessTokenExpires * 60, "/", "", false, true)

	return nil
}
