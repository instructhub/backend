package controllers

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/app/queries"
	"github.com/instructhub/backend/pkg/cache"
	"github.com/instructhub/backend/pkg/encryption"
	"github.com/instructhub/backend/pkg/logger"
	"github.com/instructhub/backend/pkg/utils"
	"github.com/markbates/goth/gothic"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type EmailAuthRequest struct {
	Username    string `json:"username" binding:"required,max=32,min=3,username"`
	DisplayName string `json:"display_name" binding:"required,max=32,min=1,alphanumunicode"`
	Email       string `json:"email" binding:"required,email,max=320"`
	Password    string `json:"password" binding:"required,max=128,min=8"`
}

// For user using email signup
func Signup(c *gin.Context) {
	var request EmailAuthRequest

	// Validate request body
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.FullyResponse(c, 400, "Invalid request", utils.ErrBadRequest, err.Error())
		return
	}

	// Check if email already been used
	_, result := queries.GetUserQueueByEmail(request.Email)
	if result.Error == nil {
		utils.FullyResponse(c, 400, "Email already been used", utils.ErrEmailAlreadyUsed, nil)
		return
	} else if result.Error != gorm.ErrRecordNotFound {
		utils.ServerErrorResponse(c, 500, "Error checking email", utils.ErrGetData, result.Error)
		return
	}

	// Check if username already been used
	_, result = queries.GetUserQueueByUsername(request.Username)
	if result.Error == nil {
		utils.FullyResponse(c, 400, "Username already been used", utils.ErrUsernameAlreadyUsed, nil)
		return
	} else if result.Error != gorm.ErrRecordNotFound {
		utils.ServerErrorResponse(c, 500, "Error checking username", utils.ErrGetData, result.Error)
		return
	}

	user := models.User{
		ID:          encryption.GenerateID(),
		DisplayName: request.DisplayName,
		Username:    request.Username,
		Email:       request.Email,
		Password:    request.Password,
		Verify:      false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Hash password
	hashedPassword, err := encryption.HashPassword(user.Password)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error hash passwsord", utils.ErrHashData, err)
		return
	}
	user.Password = hashedPassword

	// Generate verification token
	verifyToken, err := encryption.GenerateRandomBase64String(512)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error hash generate verification token", utils.ErrHashData, err)
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
		utils.ServerErrorResponse(c, 500, "Error parse email template", utils.ErrParseFile, err)
		return
	}
	t.ExecuteTemplate(&emailBody, "email_verificaiton.html", data)
	err = utils.SendEmail(user.Email, "Verification your email", emailBody.String())
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error send verification email", utils.ErrSendEmail, err)
		return
	}

	// Store the verification key in Redis (with expiration)
	err = cache.RedisClient.Set(c, verifyToken, user.ID, 15*time.Minute).Err()
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error storing verification key", utils.ErrSaveData, err)
		return
	}

	// Create user in the queue
	result = queries.CreateUserQueue(user)
	if result.Error != nil || result.RowsAffected == 0 {
		utils.ServerErrorResponse(c, 500, "Error create new user", utils.ErrSaveData, result.Error)
		return
	}

	// Set a verify pedding jwt cookie for the user
	verifyPeddingTokenExpiresAt := time.Now().Add(time.Minute * 15)
	verifyPeddintToken, err := encryption.GenerateNewJwtToken(user.ID, []string{"pedding"}, verifyPeddingTokenExpiresAt)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error generate verify pedding token", utils.ErrGenerateToken, err)
		return
	}
	c.SetCookie("verify_pedding", verifyPeddintToken, 15*60, "/", "", false, false)

	utils.FullyResponse(c, 200, "Signup successful please verify email", nil, nil)
}

type EmailLoginRequest struct {
	Email    string `json:"email" binding:"email,max=320"`
	Password string `json:"password" binding:"required,max=128,min=8"`
}

// For login with email
func Login(c *gin.Context) {
	var request EmailLoginRequest

	// Validate request body
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.FullyResponse(c, 400, "Invalid request", utils.ErrBadRequest, err.Error())
		return
	}

	if request.Email == "" {
		utils.FullyResponse(c, 400, "Email is required", utils.ErrBadRequest, nil)
		return
	}

	var user models.User
	var result *gorm.DB

	user, result = queries.GetUserQueueByEmail(request.Email)
	if result.Error == gorm.ErrRecordNotFound {
		utils.FullyResponse(c, 400, "Invalid email", utils.ErrInvalidUsernameOrEmail, nil)
		return
	} else if result.Error != nil {
		utils.ServerErrorResponse(c, 500, "Error check email", utils.ErrGetData, result.Error)
		return
	}

	if user.Password == "" {
		utils.FullyResponse(c, 400, "Invalid password", utils.ErrInvalidPassword, nil)
		return
	}

	match, err := encryption.ComparePasswordAndHash(request.Password, user.Password)
	if err != nil || !match {
		utils.FullyResponse(c, 400, "Invalid password", utils.ErrInvalidPassword, nil)
		return
	}

	type notVerify struct {
		Verify bool `json:"verify"`
	}
	if !user.Verify {
		// Set a verify pedding jwt cookie for the user
		verifyPeddingTokenExpiresAt := time.Now().Add(time.Minute * 15)
		verifyPeddintToken, err := encryption.GenerateNewJwtToken(user.ID, []string{"pedding"}, verifyPeddingTokenExpiresAt)
		if err != nil {
			utils.ServerErrorResponse(c, 500, "Error generate verify pedding token", utils.ErrGenerateToken, err)
			return
		}
		c.SetCookie("verify_pedding", verifyPeddintToken, 15*60, "/", "", false, false)
		utils.FullyResponse(c, 403, "Email not verify", utils.ErrEmailNotVerify, notVerify{
			Verify: false,
		})
		return
	}

	err = utils.GenerateUserSession(c, user.ID)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Internal server error", utils.ErrGenerateSession, err)
		return
	}

	utils.FullyResponse(c, 200, "Login successful", nil, notVerify{
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

func GetProviderFromString(provider string) (models.Provider, error) {
	switch provider {
	case "google":
		return models.ProviderGoogle, nil
	case "github":
		return models.ProviderGithub, nil
	case "gitlab":
		return models.ProviderGitlab, nil
	default:
		return -1, fmt.Errorf("unknown provider: %s", provider)
	}
}

// OAuth callback handler for Google, GitHub, etc.
func OAuthCallbackHandler(c *gin.Context, cprovider string) {
	// Add the provider to the query parameters to keep track of it
	q := c.Request.URL.Query()
	q.Add("provider", cprovider)
	c.Request.URL.RawQuery = q.Encode()

	// Complete the user authentication
	request, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		utils.FullyResponse(c, 400, "Invalid request", utils.ErrBadRequest, err.Error())
		return
	}

	var user models.User
	// Get user and associated OAuth providers by email
	user, result := queries.GetUserAndProvider(request.Email)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		utils.ServerErrorResponse(c, 500, "Error getting user", utils.ErrGetData, result.Error)
		return
	}

	// Convert the provider string to the corresponding enum value
	provider, ok := models.ParseStringToProviderType(request.Provider)
	if !ok {
		utils.ServerErrorResponse(c, 500, "Error parsing provider data", utils.ErrParseData, fmt.Errorf("internal server error while parsing provider data"))
		return
	}

	// If the user is found
	if result.Error == nil {
		logger.Log.Info("test")
		// Iterate through associated OAuth providers
		for _, p := range *user.OauthProviders {
			// Skip to the next iteration if the provider doesn't match
			if p.Provider != provider {
				continue
			}

			// Check if the OAuthID matches
			if p.OAuthID != request.UserID {
				utils.FullyResponse(c, 403, "OAuthID mismatched!", utils.ErrUnauthorized, nil)
				return
			}

			// Generate user session after successful authentication
			err = utils.GenerateUserSession(c, user.ID)
			if err != nil {
				utils.ServerErrorResponse(c, 500, "Error generating session", utils.ErrGenerateSession, err)
				return
			}

			// Send a successful login response
			c.HTML(200, "auth_successful.html", gin.H{
				"Title":   "Login Successful",
				"Message": "Welcome back! You've successfully logged in.",
			})
			return
		}

		// If provider is new, add it to the database
		reseult := queries.AddUserProvider(models.OauthProvider{
			ID:        encryption.GenerateID(),
			UserID:    user.ID,
			Provider:  provider,
			OAuthID:   request.UserID,
			UpdatedAt: time.Now(),
			CreatedAt: time.Now(),
		})

		// Check if there was an error while adding the provider
		if result.Error != nil || reseult.RowsAffected == 0 {
			utils.ServerErrorResponse(c, 500, "Error adding OAuth provider", utils.ErrSaveData, result.Error)
			return
		}

		// Generate user session after successful provider addition
		err = utils.GenerateUserSession(c, user.ID)
		if err != nil {
			utils.ServerErrorResponse(c, 500, "Error generating session", utils.ErrGenerateSession, err)
			return
		}

		// Send a successful response when a new login option is added
		c.HTML(200, "auth_successful.html", gin.H{
			"Title":   "New login option added successfully!",
			"Message": "Welcome back! You've successfully logged in.",
		})
		return
	}

	// Handle the case when the user is not found, and create a new user
	userID := encryption.GenerateID()
	userIDString := utils.Uint64ToStr(uint64(userID))
	// New user creation process
	user = models.User{
		ID:          userID,
		Avatar:      &request.AvatarURL,
		DisplayName: request.Name,
		Username:    userIDString,
		Email:       request.Email,
		Verify:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Check if the generated username already exists, and regenerate if needed
	for i := 0; i < 5; i++ {
		_, result := queries.GetUserQueueByUsername(user.Username)
		if result.Error == gorm.ErrRecordNotFound {
			break // No conflict, break the loop
		}
		if i == 4 {
			utils.ServerErrorResponse(c, 500, "Error generating username", utils.ErrSaveData, fmt.Errorf("error generating username"))
			return
		}

		// Generate a new user ID and username if a duplicate is found
		userID = encryption.GenerateID()                 // Generate new user ID
		userIDString = utils.Uint64ToStr(uint64(userID)) // Generate new username string
		user.Username = userIDString                     // Update the username field
		user.ID = userID                                 // Update the user ID field
	}

	// Create the new user in the database
	result = queries.CreateUserQueue(user)
	if result.Error != nil || result.RowsAffected == 0 {
		utils.ServerErrorResponse(c, 500, "Error creating user", utils.ErrSaveData, result.Error)
		return
	}

	reseult := queries.AddUserProvider(models.OauthProvider{
		ID:        encryption.GenerateID(),
		UserID:    user.ID,
		Provider:  provider,
		OAuthID:   request.UserID,
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	})
	// Check if there was an error while adding the provider
	if result.Error != nil || reseult.RowsAffected == 0 {
		utils.ServerErrorResponse(c, 500, "Error adding OAuth provider", utils.ErrSaveData, result.Error)
		return
	}

	// Generate user session after successful user creation
	err = utils.GenerateUserSession(c, user.ID)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error generating session", utils.ErrGenerateSession, err)
		return
	}

	// Send a successful response for new user signup
	c.HTML(200, "auth_successful.html", gin.H{
		"Title":   "You have successfully signed up!",
		"Message": "Signup successful! We’re glad to have you with us.",
	})
}

func RefreshAccessToken(c *gin.Context) {
	// Retrieve the refresh token from the cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		utils.FullyResponse(c, 403, "Invalid refresh token", utils.ErrUnauthorized, nil)
		return
	}

	// Check refresh token valid
	session, result := queries.GetSessionQueueBySecretKey(refreshToken)
	if result.Error == gorm.ErrRecordNotFound {
		utils.FullyResponse(c, 403, "Invalid refresh token", utils.ErrUnauthorized, nil)
		return
	} else if result.Error != nil {
		utils.ServerErrorResponse(c, 500, "Error get session", utils.ErrGetData, result.Error)
		return
	}

	// Delete this session
	result = queries.DeleteSessionQueue(refreshToken)
	if result.Error != nil {
		utils.ServerErrorResponse(c, 500, "Error delete session", utils.ErrDeleteData, result.Error)
		return
	} else if result.RowsAffected == 0 {
		utils.FullyResponse(c, 403, "Invalid refresh token", utils.ErrUnauthorized, nil)
		return
	}

	// Set the new access token in the response cookie
	err = utils.GenerateUserSession(c, session.UserID)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error generate session", utils.ErrGenerateSession, err)
		return
	}
	// Return a successful response
	utils.FullyResponse(c, 200, "Successfully refreshed access token", nil, nil)
}

func LogOut(c *gin.Context) {
	// Retrieve the refresh token from the cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		utils.FullyResponse(c, 403, "Invalid refresh token", utils.ErrUnauthorized, nil)
		return
	}

	// Attempt to delete the session associated with the refresh token
	result := queries.DeleteSessionQueue(refreshToken)
	if result.Error != nil {
		utils.ServerErrorResponse(c, 500, "Error delete session", utils.ErrGetData, result.Error)
		return
	} else if result.Error == gorm.ErrRecordNotFound {
		utils.FullyResponse(c, 403, "Invalid refresh token", utils.ErrUnauthorized, nil)
		return
	}

	// Clear the refresh token and access token cookies by setting their expiry date to -1
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	c.SetCookie("access_token", "", -1, "/", "", false, false)

	// Return a successful response
	utils.FullyResponse(c, 200, "Logged out successfully", "", nil)
}

// CheckEmailVerify checks if the user's email has been verified
func CheckEmailVerify(c *gin.Context) {
	type resp struct {
		Verify bool `json:"verify"`
	}

	// Convert userID from string to uint64
	userID, err := utils.StrToUint64(c.Param("userID"))
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error parse userid", utils.ErrParse, err)
		return
	}

	// Retrieve user from the database
	user, result := queries.GetUserQueueByID(userID)
	if result.Error == gorm.ErrRecordNotFound {
		utils.FullyResponse(c, 404, "User not found", utils.ErrGetData, nil)
		return
	} else if result.Error != nil {
		utils.ServerErrorResponse(c, 500, "Error get data", utils.ErrGetData, result.Error)
		return
	}

	// Respond with user's verification status
	utils.FullyResponse(c, 200, "Successfully retrieved email verification status", "", resp{
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
			utils.FullyResponse(c, 400, "Invalid verification key", utils.ErrBadRequest, nil)
			c.Redirect(303, utils.FrontendURl+"/login?verify=false")
			return
		}
		// Handle Redis-related errors
		utils.ServerErrorResponse(c, 500, "Error accessing Redis", utils.ErrGetData, err)
		return
	}

	// Convert userID from string to uint64
	userID, err := utils.StrToUint64(userIDString)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error parsing user ID", utils.ErrParseFile, err)
		return
	}

	// Update the user's email verification status in the database
	result := queries.UpdateUserVerifyStatus(userID, true)
	if result.Error != nil || result.RowsAffected == 0 {
		utils.ServerErrorResponse(c, 500, "Error updating user email verification status", utils.ErrSaveData, result.Error)
		return
	}

	// Delete the verifyKey from Redis after successful verification
	err = cache.RedisClient.Del(c, verifyKey).Err()
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error delete data frm redis", utils.ErrSaveData, err)
		return
	}

	_, exist := c.Get("userID")
	if exist {
		err = utils.GenerateUserSession(c, userID)
		if err != nil {
			utils.ServerErrorResponse(c, 500, "Error generate session", utils.ErrGenerateSession, err)
			return
		}
		c.Redirect(303, utils.FrontendURl+"/")
	}

	// Redirect to the frontend with the result of the verification
	c.Redirect(303, utils.FrontendURl+"/login?verify=true")
}

// TODO: Need to use a rate limiter with 60 secs per request
// ResendVerificationEmail handles resending the verification email
func ResendVerificationEmail(c *gin.Context) {
	// Get the userID from the URL parameters
	ContextUserID, exist := c.Get("userID")
	if !exist {
		utils.FullyResponse(c, 403, "Unauthorized", utils.ErrUnauthorized, nil)
		return
	}

	userID := ContextUserID.(uint64)

	// Retrieve user details from the database
	user, result := queries.GetUserQueueByID(userID)
	if result.Error == gorm.ErrRecordNotFound {
		utils.FullyResponse(c, 404, "User not find", utils.ErrGetData, nil)
		return
	} else if result.Error != nil {
		utils.ServerErrorResponse(c, 500, "Error fetching user", utils.ErrGetData, result.Error)
		return
	}

	// If the user is already verified, return a response
	if user.Verify {

		utils.FullyResponse(c, 400, "User is already verified", nil, nil)
		return
	}

	// Generate a random verification token
	verifyToken, err := encryption.GenerateRandomBase64String(512)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error generating verification token", utils.ErrGenerateToken, err)
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
		utils.ServerErrorResponse(c, 500, "Error parsing email template", utils.ErrParseFile, err)
		return
	}
	err = t.ExecuteTemplate(&emailBody, "email_verificaiton.html", data)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error executing email template", utils.ErrExecuteTemplate, err)
		return
	}

	// Send the verification email
	err = utils.SendEmail(user.Email, "Verify your email", emailBody.String())
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error sending verification email", utils.ErrSendEmail, err)
		return
	}

	// Store the verification token in Redis with an expiration of 15 minutes
	err = cache.RedisClient.Set(c, verifyToken, user.ID, 15*time.Minute).Err()
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error storing verification token in Redis", utils.ErrStoreRedis, err)
		return
	}

	// Return a success response
	utils.FullyResponse(c, 200, "Verification email successfully sent", nil, nil)
}
