package controllers

import (
	"fmt"
	"time"

	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/app/queues"
	"github.com/instructhub/backend/pkg/encryption"
	"github.com/instructhub/backend/pkg/utils"
	"github.com/markbates/goth/gothic"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gin-gonic/gin"
)

// For user using email signup
func Signup(c *gin.Context) {
	type SignupRequest struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	var request SignupRequest

	// Validate request body
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SimpleResponse(c, 400, "Invalid request", err.Error())
		return
	}

	if err := utils.Validator().Struct(request); err != nil {
		utils.SimpleResponse(c, 400, "Invalid request", err.Error())
		return
	}

	// Check if email already been used
	_, err := queues.GetUserQueueByEmail(request.Email)
	if err == nil {
		utils.SimpleResponse(c, 400, "Email already been used", nil)
		return
	} else if err != mongo.ErrNoDocuments {
		utils.SimpleResponse(c, 500, "Internal server error while checking email", err.Error())
		return
	}

	// Check if username already been used
	_, err = queues.GetUserQueueByUsername(request.Username)
	if err == nil {
		utils.SimpleResponse(c, 400, "Username already been used", nil)
		return
	} else if err != mongo.ErrNoDocuments {
		utils.SimpleResponse(c, 500, "Internal server error while checking username", err.Error())
		return
	}

	user := models.User{
		ID:        encryption.GenerateID(),
		Username:  request.Username,
		Email:     request.Email,
		Password:  request.Password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Hash password
	hashedPassword, err := encryption.HashPassword(user.Password)
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error", err.Error())
		return
	}
	user.Password = hashedPassword

	err = queues.CreateUserQueue(user)
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error", err.Error())
		return
	}

	err = utils.GenerateUserSession(c, user.ID)
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error", err.Error())
		return
	}

	utils.SimpleResponse(c, 200, "Signup successful", nil)
}


// For login with email
func Login(c *gin.Context) {
	type LoginRequest struct {
		Username string `json:"username,omitempty"`
		Email    string `json:"email,omitempty" binding:"email"`
		Password string `json:"password" binding:"required"`
	}

	var request LoginRequest

	// Validate request body
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SimpleResponse(c, 400, "Invalid request", err.Error())
		return
	}

	if err := utils.Validator().Struct(request); err != nil {
		utils.SimpleResponse(c, 400, "Invalid request", err.Error())
		return
	}

	if request.Username == "" && request.Email == "" {
		utils.SimpleResponse(c, 400, "Username or email is required", nil)
		return
	}

	var user models.User
	var err error

	if request.Username != "" {
		user, err = queues.GetUserQueueByUsername(request.Username)
	} else {
		user, err = queues.GetUserQueueByEmail(request.Email)
	}

	if err != nil {
		utils.SimpleResponse(c, 400, "Invalid username or email", nil)
		return
	}

	match, err := encryption.ComparePasswordAndHash(request.Password, user.Password)
	if err != nil || !match {
		utils.SimpleResponse(c, 400, "Invalid password", nil)
		return
	}

	err = utils.GenerateUserSession(c, user.ID)
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error", err.Error())
		return
	}

	utils.SimpleResponse(c, 200, "Login successful", nil)
}

// Call Oauth login with google github etc
func OAuthHandler(c *gin.Context, cprovider string) {
	q := c.Request.URL.Query()
	q.Add("provider", cprovider)
	c.Request.URL.RawQuery = q.Encode()
	gothic.BeginAuthHandler(c.Writer, c.Request)
}

// Oauth call back for google github etc
func OAuthCallbackHandler(c *gin.Context, cprovider string) {
	q := c.Request.URL.Query()
	q.Add("provider", cprovider)
	c.Request.URL.RawQuery = q.Encode()
	request, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		utils.SimpleResponse(c, 500, "Interal server error", err)
		return
	}

	var user models.User
	user, err = queues.GetUserQueueByEmail(request.Email)

	if err != nil && err != mongo.ErrNoDocuments {
		utils.SimpleResponse(c, 500, "Internal server error", err.Error())
		return
	}

	if err == nil {
		for i, p := range user.Providers {
			if p.Provider == request.Provider {
				if user.Providers[i].OAuthID == request.UserID {
					err = utils.GenerateUserSession(c, user.ID)
					if err != nil {
						utils.SimpleResponse(c, 500, "Internal server error", err.Error())
						return
					}

					utils.SimpleResponse(c, 200, "Login successful and added another provider", nil)
					return
				} else {
					utils.SimpleResponse(c, 403, "OAuthID mismatched!", nil)
					return
				}
			}
		}
		// If provider is new, append it
		user.Providers = append(user.Providers, models.Provider{Provider: request.Provider, OAuthID: request.UserID})
		user.UpdatedAt = time.Now()

		queues.AppendUserProviderQueue(uint64(user.ID), user)

		err = utils.GenerateUserSession(c, user.ID)
		if err != nil {
			utils.SimpleResponse(c, 500, "Internal server error", err.Error())
			return
		}

		utils.SimpleResponse(c, 200, "Login successful", nil)
		return
	}

	user = models.User{
		ID:        encryption.GenerateID(),
		Avatar:    request.AvatarURL,
		Username:  request.Name,
		Email:     request.Email,
		Providers: []models.Provider{{Provider: request.Provider, OAuthID: request.UserID}}, // IDK why it's double braces
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = queues.CreateUserQueue(user)
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error", err.Error())
		return
	}

	err = utils.GenerateUserSession(c, user.ID)
	if err != nil {
		fmt.Println(err.Error())
		utils.SimpleResponse(c, 500, "Internal server error", err.Error())
		return
	}

	// Respond with success
	utils.SimpleResponse(c, 200, "Signup successful", nil)
}

func RefreshAccessToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		utils.SimpleResponse(c, 403, "Invide refresh token", nil)
	}

	userID := uint64(utils.Atoi(c.Param("userID")))

	session, err := queues.GetSessionQueue(refreshToken)
	if err == mongo.ErrNoDocuments {
		utils.SimpleResponse(c, 403, "Invide refresh token", nil)
		return
	}
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal error", err.Error())
	}

	if session.UserID != userID {
		utils.SimpleResponse(c, 403, "User ID not match with refresh_token", nil)
		return
	}

	accessTokenExpiresAt := time.Now().Add(time.Minute * time.Duration(utils.CookieAccessTokenExpires))
	accessToken, err := encryption.GenerateNewJwtToken(userID, []string{}, accessTokenExpiresAt)
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal error", nil)
		return
	}

	c.SetCookie("access_token", accessToken, utils.CookieAccessTokenExpires * 60, "/", "", false, true)
}
