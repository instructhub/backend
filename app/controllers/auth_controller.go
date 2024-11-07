package controllers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/app/queries"
	"github.com/instructhub/backend/pkg/encryption"
	"github.com/instructhub/backend/pkg/utils"
	"github.com/markbates/goth/gothic"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gin-gonic/gin"
)

type EmailAuthRequest struct {
	Username string `json:"username" binding:"required,max=30,min=3,alphanum"`
	Email    string `json:"email" binding:"required,email,max=320"`
	Password string `json:"password" binding:"required,max=128,min=8"`
}

// For user using email signup
func Signup(c *gin.Context) {
	var request EmailAuthRequest

	// Validate request body
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SimpleResponse(c, 400, "Invalid request", err.Error())
		return
	}

	// Check if email already been used
	_, err := queries.GetUserQueueByEmail(request.Email)
	if err == nil {
		utils.SimpleResponse(c, 400, "Email already been used", nil)
		return
	} else if err != mongo.ErrNoDocuments {
		utils.SimpleResponse(c, 500, "Internal server error while checking email", err.Error())
		return
	}

	// Check if username already been used
	_, err = queries.GetUserQueueByUsername(request.Username)
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
		Verify: false,
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

	err = queries.CreateUserQueue(user)
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

type EmailLoginRequest struct {
	Username string `json:"username" binding:"max=30"`
	Email    string `json:"email" binding:"email,max=320"`
	Password string `json:"password" binding:"required,max=128,min=8"`
}

// For login with email
func Login(c *gin.Context) {
	var request EmailLoginRequest

	// Validate request body
	if err := c.ShouldBindJSON(&request); err != nil {
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
		user, err = queries.GetUserQueueByUsername(request.Username)
	} else {
		user, err = queries.GetUserQueueByEmail(request.Email)
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

	if !user.Verify {
		type notVerify struct{
			Verify bool `json:"verify"`
		}
		userIDString := strconv.FormatInt(int64(user.ID), 10)
		c.SetCookie("userID", userIDString, 15 * 60, "/", "", false, true)
		utils.SimpleResponse(c, 403, "Email not verify", notVerify{
			Verify: false,
		})
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
	user, err = queries.GetUserQueueByEmail(request.Email)

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

					c.HTML(200, "auth_successful.html", gin.H{
						"Title": "Login Successful",
						"Message": "Welcome back! You've successfully logged in.",
					})
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

		queries.AppendUserProviderQueue(uint64(user.ID), user)

		err = utils.GenerateUserSession(c, user.ID)
		if err != nil {
			utils.SimpleResponse(c, 500, "Internal server error", err.Error())
			return
		}

		c.HTML(200, "auth_successful.html", gin.H{
			"Title": "New login option added successfully!",
			"Message": "Welcome back! You've successfully logged in.",
		})
		return
	}

	user = models.User{
		ID:        encryption.GenerateID(),
		Avatar:    request.AvatarURL,
		Username:  request.Name,
		Email:     request.Email,
		Providers: []models.Provider{{Provider: request.Provider, OAuthID: request.UserID}},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = queries.CreateUserQueue(user)
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
	c.HTML(200, "auth_successful.html", gin.H{
		"Title": "You have successfully signed up!",
		"Message": "Signup successful! We’re glad to have you with us.",
	})
}

func RefreshAccessToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		utils.SimpleResponse(c, 403, "Invide refresh token", nil)
	}

	userID := uint64(utils.Atoi(c.Param("userID")))

	session, err := queries.GetSessionQueue(refreshToken)
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

	c.SetCookie("access_token", accessToken, utils.CookieAccessTokenExpires*60, "/", "", false, true)
	utils.SimpleResponse(c, 200, "Successful rotate access token", nil)
}

func LogOut(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		utils.SimpleResponse(c, 403, "Invalid refresh token", nil)
		return
	}

	result := queries.DeleteSessionQueue(refreshToken)
	if result.Err() == mongo.ErrNoDocuments {
		utils.SimpleResponse(c, 403, "Invalid refresh token", nil)
		return
	}
	if result.Err() != nil {
		utils.SimpleResponse(c, 500, "Internal error", result.Err().Error())
		return
	}

	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	c.SetCookie("access_token", "", -1, "/", "", false, true)

	utils.SimpleResponse(c, 200, "Logged out successfully", nil)
}

func CheckEmailVerify(c *gin.Context) {
	type resp struct {
		Verify bool `json:"verify"`
	}

	userID := uint64(utils.Atoi(c.Param("userID")))

	user, err := queries.GetUserQueueByID(userID)
	if err != nil {
		utils.SimpleResponse(c, 404, "This user not exist", nil)
	}

	utils.SimpleResponse(c, 200, "Successful get user verify status", resp{
		Verify: user.Verify,
	})
}
