package controllers

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"
	"time"

	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/app/queries"
	"github.com/instructhub/backend/pkg/cache"
	"github.com/instructhub/backend/pkg/encryption"
	"github.com/instructhub/backend/pkg/utils"
	"github.com/markbates/goth/gothic"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gin-gonic/gin"
)

type EmailAuthRequest struct {
	Username string `json:"username" binding:"required,max=30,min=3,username"`
	Email    string `json:"email" binding:"required,email,max=320"`
	Password string `json:"password" binding:"required,max=128,min=8"`
}

// For user using email signup
func Signup(c *gin.Context) {
	var request EmailAuthRequest

	// Validate request body
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SimpleResponse(c, 400, "Invalid request", utils.ErrBadRequest, err.Error())
		return
	}

	// Check if email already been used
	_, err := queries.GetUserQueueByEmail(request.Email)
	if err == nil {
		utils.SimpleResponse(c, 400, "Email already been used", utils.ErrEmailAlreadyUsed, nil)
		return
	} else if err != mongo.ErrNoDocuments {
		utils.SimpleResponse(c, 500, "Internal server error while checking email", utils.ErrGetData, nil)
		return
	}

	// Check if username already been used
	_, err = queries.GetUserQueueByUsername(request.Username)
	if err == nil {
		utils.SimpleResponse(c, 400, "Username already been used", utils.ErrUsernameAlreadyUsed, nil)
		return
	} else if err != mongo.ErrNoDocuments {
		utils.SimpleResponse(c, 500, "Internal server error while checking username", utils.ErrGetData, err.Error())
		return
	}

	user := models.User{
		ID:        encryption.GenerateID(),
		Username:  request.Username,
		Email:     request.Email,
		Password:  request.Password,
		Verify:    false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Hash password
	hashedPassword, err := encryption.HashPassword(user.Password)
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error", utils.ErrHashData, err.Error())
		return
	}
	user.Password = hashedPassword

	// Generate verification token
	verifyToken, err := encryption.GenerateRandomBase64String(512)
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error", utils.ErrHashData, err.Error())
		return
	}

	type EmailData struct {
		VerifyURL string
		UserName  string
	}
	data := EmailData{
		VerifyURL: utils.BackendURL + "/auth/email/verify/" + verifyToken,
		UserName:  user.Username,
	}
	var emailBody bytes.Buffer
	t := template.New("Email verification")
	t, err = t.ParseFiles("template/email_verificaiton.html")
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error", utils.ErrParseFile, err.Error())
	}
	t.ExecuteTemplate(&emailBody, "email_verificaiton.html", data)
	err = utils.SendEmail(user.Email, "Verification your email", emailBody.String())
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error while send verification email", utils.ErrSendEmail, err.Error())
		return
	}

	// Store the verification key in Redis (with expiration)
	err = cache.RedisClient.Set(c, verifyToken, user.ID, 15*time.Minute).Err()
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error while storing verification key", utils.ErrSaveData, err.Error())
		return
	}

	// Create user in the queue
	err = queries.CreateUserQueue(user)
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error", utils.ErrSaveData, err.Error())
		return
	}

	// Set a cookie for the user (this could be adjusted to fit your needs)
	userIDString := strconv.FormatInt(int64(user.ID), 10)
	c.SetCookie("userID", userIDString, 60*60, "/", "", false, false)

	utils.SimpleResponse(c, 200, "Signup successful please verify email", nil, nil)
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
		utils.SimpleResponse(c, 400, "Invalid request", utils.ErrBadRequest, err.Error())
		return
	}

	if request.Username == "" && request.Email == "" {
		utils.SimpleResponse(c, 400, "Username or email is required", utils.ErrBadRequest, nil)
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
		utils.SimpleResponse(c, 400, "Invalid username or email", utils.ErrInvalidUsernameOrEmail, nil)
		return
	}

	match, err := encryption.ComparePasswordAndHash(request.Password, user.Password)
	if err != nil || !match {
		utils.SimpleResponse(c, 400, "Invalid password", utils.ErrInvalidPassword, nil)
		return
	}

	type notVerify struct {
		Verify bool `json:"verify"`
	}
	if !user.Verify {
		userIDString := strconv.FormatInt(int64(user.ID), 10)
		c.SetCookie("userID", userIDString, 60*60, "/", "", false, false)
		utils.SimpleResponse(c, 403, "Email not verify", utils.ErrEmailNotVerify, notVerify{
			Verify: false,
		})
		return
	}

	err = utils.GenerateUserSession(c, user.ID)
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error", utils.ErrGenerateSession, err.Error())
		return
	}

	utils.SimpleResponse(c, 200, "Login successful", nil, notVerify{
		Verify: true,
	})
}

// Call Oauth login with google github etc
func OAuthHandler(c *gin.Context, cprovider string) {
	q := c.Request.URL.Query()
	q.Add("provider", cprovider)
	c.Request.URL.RawQuery = q.Encode()
	gothic.BeginAuthHandler(c.Writer, c.Request)
}

// OAuth callback handler for Google, GitHub, etc.
func OAuthCallbackHandler(c *gin.Context, cprovider string) {
	q := c.Request.URL.Query()
	q.Add("provider", cprovider)
	c.Request.URL.RawQuery = q.Encode()

	request, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		utils.SimpleResponse(c, 400, "Invalid request", utils.ErrBadRequest, err.Error())
		return
	}

	var user models.User
	user, err = queries.GetUserQueueByEmail(request.Email)

	if err != nil && err != mongo.ErrNoDocuments {
		utils.SimpleResponse(c, 500, "Internal server error", utils.ErrGetData, err.Error())
		return
	}

	if err == nil {
		for i, p := range user.Providers {
			if p.Provider != request.Provider {
				continue
			}
			if user.Providers[i].OAuthID != request.UserID {
				utils.SimpleResponse(c, 403, "OAuthID mismatched!", utils.ErrUnauthorized, nil)
				return
			}
			err = utils.GenerateUserSession(c, user.ID)
			if err != nil {
				utils.SimpleResponse(c, 500, "Internal server error", utils.ErrGenerateSession, err.Error())
				return
			}

			c.HTML(200, "auth_successful.html", gin.H{
				"Title":   "Login Successful",
				"Message": "Welcome back! You've successfully logged in.",
			})
			return
		}
		// If provider is new, append it
		user.Providers = append(user.Providers, models.Provider{Provider: request.Provider, OAuthID: request.UserID})
		user.UpdatedAt = time.Now()

		queries.AppendUserProviderQueue(uint64(user.ID), user)

		err = utils.GenerateUserSession(c, user.ID)
		if err != nil {
			utils.SimpleResponse(c, 500, "Internal server error", utils.ErrGenerateSession, err.Error())
			return
		}

		c.HTML(200, "auth_successful.html", gin.H{
			"Title":   "New login option added successfully!",
			"Message": "Welcome back! You've successfully logged in.",
		})
		return
	}

	// New user creation process
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
		utils.SimpleResponse(c, 500, "Internal server error", utils.ErrSaveData, err.Error())
		return
	}

	err = utils.GenerateUserSession(c, user.ID)
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error", utils.ErrGenerateSession, err.Error())
		return
	}

	// Respond with success
	c.HTML(200, "auth_successful.html", gin.H{
		"Title":   "You have successfully signed up!",
		"Message": "Signup successful! We’re glad to have you with us.",
	})
}

func RefreshAccessToken(c *gin.Context) {
	// Retrieve the refresh token from the cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		utils.SimpleResponse(c, 403, "Invalid refresh token", utils.ErrUnauthorized, nil)
		return
	}

	session, err := queries.GetSessionQueue(refreshToken)
	if err == mongo.ErrNoDocuments {
		utils.SimpleResponse(c, 403, "Invalid refresh token", utils.ErrUnauthorized, nil)
		return
	}
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal error", utils.ErrGetData, err.Error())
		return
	}

	userID := session.UserID

	// Generate a new access token with the desired expiration
	accessTokenExpiresAt := time.Now().Add(time.Minute * time.Duration(utils.CookieAccessTokenExpires))
	accessToken, err := encryption.GenerateNewJwtToken(userID, []string{}, accessTokenExpiresAt)
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal error", utils.ErrGenerateSession, err.Error())
		return
	}

	// Set the new access token in the response cookie
	c.SetCookie("access_token", accessToken, utils.CookieAccessTokenExpires*60, "/", "", false, false)

	// Return a successful response
	utils.SimpleResponse(c, 200, "Successfully refreshed access token", nil, nil)
}

func LogOut(c *gin.Context) {
	// Retrieve the refresh token from the cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		utils.SimpleResponse(c, 403, "Invalid refresh token", utils.ErrUnauthorized, nil)
		return
	}

	// Attempt to delete the session associated with the refresh token
	result := queries.DeleteSessionQueue(refreshToken)
	if result.Err() == mongo.ErrNoDocuments {
		utils.SimpleResponse(c, 403, "Invalid refresh token", utils.ErrUnauthorized, nil)
		return
	}
	if result.Err() != nil {
		utils.SimpleResponse(c, 500, "Internal error", utils.ErrGetData, result.Err().Error())
		return
	}

	// Clear the refresh token and access token cookies by setting their expiry date to -1
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	c.SetCookie("access_token", "", -1, "/", "", false, false)

	// Return a successful response
	utils.SimpleResponse(c, 200, "Logged out successfully", "", nil)
}

// CheckEmailVerify checks if the user's email has been verified
func CheckEmailVerify(c *gin.Context) {
	type resp struct {
		Verify bool `json:"verify"`
	}

	// Convert userID from string to uint64
	userID, err := utils.StringToUint64(c.Param("userID"))
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error", utils.ErrParseFile, err.Error())
		return
	}

	// Retrieve user from the database
	user, err := queries.GetUserQueueByID(userID)
	if err != nil {
		utils.SimpleResponse(c, 404, "User not found", utils.ErrGetData, nil)
		return
	}

	// Respond with user's verification status
	utils.SimpleResponse(c, 200, "Successfully retrieved email verification status", "", resp{
		Verify: user.Verify,
	})
}

// VerifyEmail handles the email verification process
func VerifyEmail(c *gin.Context) {
	// Get the verifyKey from the URL parameters
	verifyKey := c.Param("verifyKey")

	// Check if the verifyKey exists in Redis
	userIDString, err := cache.RedisClient.Get(c, verifyKey).Result()
	if err != nil {
		if err == redis.Nil {
			utils.SimpleResponse(c, 400, "Invalid verification key", utils.ErrBadRequest, nil)
			c.Redirect(303, utils.FrontendURl+"/login?verify=false")
			return
		}
		// Handle Redis-related errors
		utils.SimpleResponse(c, 500, "Internal server error while accessing Redis", utils.ErrGetData, err.Error())
		return
	}

	// Convert userID from string to uint64
	userID, err := utils.StringToUint64(userIDString)
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error while parsing user ID", utils.ErrParseFile, err.Error())
		return
	}

	// Update the user's email verification status in the database
	err = queries.UpdateUesrEmailVerifyStatus(userID, true)
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error while updating user email verification status", utils.ErrSaveData, err.Error())
		return
	}

	// Delete the verifyKey from Redis after successful verification
	err = cache.RedisClient.Del(c, verifyKey).Err()
	if err != nil {
		// Log the error but continue the verification process
		fmt.Println("Failed to delete verification key from Redis:", err.Error())
	}

	// Redirect to the frontend with the result of the verification
	c.Redirect(303, utils.FrontendURl+"/login?verify=true")
}

// FIXME: Need to use a rate limiter with 60 secs per request
// ResendVerificationEmail handles resending the verification email
func ResendVerificationEmail(c *gin.Context) {
	// Get the userID from the URL parameters
	userIDString := c.Param("userID")
	userID, err := utils.StringToUint64(userIDString)
	if err != nil {
		utils.SimpleResponse(c, 400, "Invalid User ID", utils.ErrParseData, err.Error())
		return
	}

	// Retrieve user details from the database
	user, err := queries.GetUserQueueByID(userID)
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error while fetching user", utils.ErrGetData, err.Error())
		return
	}

	// If the user is already verified, return a response
	if user.Verify {
		utils.SimpleResponse(c, 400, "User is already verified", nil, nil)
		return
	}

	// Generate a random verification token
	verifyToken, err := encryption.GenerateRandomBase64String(512)
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error while generating verification token", utils.ErrGenerateToken, err.Error())
		return
	}

	// Prepare email data
	type EmailData struct {
		VerifyURL string
		UserName  string
	}
	data := EmailData{
		VerifyURL: utils.BackendURL + "/auth/email/verify/" + verifyToken,
		UserName:  user.Username,
	}

	// Render the email body with the template
	var emailBody bytes.Buffer
	t := template.New("Email verification")
	t, err = t.ParseFiles("template/email_verificaiton.html")
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error while parsing email template", utils.ErrParseFile, err.Error())
		return
	}
	err = t.ExecuteTemplate(&emailBody, "email_verificaiton.html", data)
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error while executing email template", utils.ErrExecuteTemplate, err.Error())
		return
	}

	// Send the verification email
	err = utils.SendEmail(user.Email, "Verify your email", emailBody.String())
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error while sending verification email", utils.ErrSendEmail, err.Error())
		return
	}

	// Store the verification token in Redis with an expiration of 15 minutes
	err = cache.RedisClient.Set(c, verifyToken, user.ID, 15*time.Minute).Err()
	if err != nil {
		utils.SimpleResponse(c, 500, "Internal server error while storing verification token in Redis", utils.ErrStoreRedis, err.Error())
		return
	}

	// Return a success response
	utils.SimpleResponse(c, 200, "Verification email successfully sent", nil, nil)
}
