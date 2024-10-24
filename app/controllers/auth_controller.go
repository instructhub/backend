package controllers

import (
	"time"

	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/app/queues"
	"github.com/instructhub/backend/pkg/utils"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gin-gonic/gin"
)

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
		ID:        utils.GenerateID(),
		Username:  request.Username,
		Email:     request.Email,
		Password:  request.Password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(user.Password)
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

	utils.SimpleResponse(c, 200, "Signup successful", user)
}

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

	match, err := utils.ComparePasswordAndHash(request.Password, user.Password)
	if err != nil || !match {
		utils.SimpleResponse(c, 400, "Invalid password", nil)
		return
	}

	session := models.Session{
		SecretKey: utils.RandStringRunes(256),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * time.Duration(utils.CookieRefreshTokenExpires)),
		CreatedAt: time.Now(),
	}

	err = queues.CreateSessionQueue(session)
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error", err.Error())
		return
	}

	accessTokenExpiresAt := time.Now().Add(time.Minute * time.Duration(utils.CookieAccessTokenExpires))
	accessToken, err := utils.GenerateNewJwtToken(user.ID, []string{}, accessTokenExpiresAt)
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error", err.Error())
		return
	}

	c.SetCookie("refresh_token", session.SecretKey, int(session.ExpiresAt.Unix()), "/refresh", "", false, true)
	c.SetCookie("access_token", accessToken, int(accessTokenExpiresAt.Unix()), "/", "", false, true)
	utils.SimpleResponse(c, 200, "Login successful", nil)
}
