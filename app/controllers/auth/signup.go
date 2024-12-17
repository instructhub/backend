package auth

import (
	"bytes"
	"html/template"
	"time"

	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/app/queries"
	"github.com/instructhub/backend/pkg/cache"
	"github.com/instructhub/backend/pkg/encryption"
	"github.com/instructhub/backend/pkg/utils"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type emailAuthRequest struct {
	Username    string `json:"username" binding:"required,max=32,min=3,username"`
	DisplayName string `json:"display_name" binding:"required,max=32,min=1,alphanumunicode"`
	Email       string `json:"email" binding:"required,email,max=320"`
	Password    string `json:"password" binding:"required,max=128,min=8"`
}

// For user using email signup
func Signup(c *gin.Context) {
	var request emailAuthRequest

	// Validate request body
	if err := c.ShouldBindJSON(&request); err != nil {
		handleError(c, 400, "Invalid request", utils.ErrBadRequest, err)
		return
	}

	// Check if email already been used
	_, result := queries.GetUserQueueByEmail(request.Email)
	if result.Error == nil {
		handleError(c, 400, "Email already been used", utils.ErrEmailAlreadyUsed, nil)
		return
	} else if result.Error != gorm.ErrRecordNotFound {
		handleError(c, 500, "Internal server error while checking email", utils.ErrGetData, result.Error)
		return
	}

	// Check if username already been used
	_, result = queries.GetUserQueueByUsername(request.Username)
	if result.Error == nil {
		handleError(c, 400, "Username already been used", utils.ErrUsernameAlreadyUsed, nil)
		return
	} else if result.Error != gorm.ErrRecordNotFound {
		handleError(c, 500, "Internal server error while checking username", utils.ErrGetData, result.Error)
		return
	}

	// Create user
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
	hashedPassword, err := hashUserPassword(user.Password)
	if err != nil {
		handleError(c, 500, "Internal server error while hashing password", utils.ErrHashData, err)
		return
	}
	user.Password = hashedPassword

	// Generate verification token
	verifyToken, err := encryption.GenerateRandomBase64String(512)
	if err != nil {
		handleError(c, 500, "Internal server error while generating verification token", utils.ErrHashData, err)
		return
	}

	// Send verification email
	if err := sendVerificationEmail(user.Email, user.Username, verifyToken); err != nil {
		handleError(c, 500, "Internal server error while sending verification email", utils.ErrSendEmail, err)
		return
	}

	// Store the verification token in Redis
	if err := cache.RedisClient.Set(c, verifyToken, user.ID, 15*time.Minute).Err(); err != nil {
		handleError(c, 500, "Internal server error while storing verification key", utils.ErrSaveData, err)
		return
	}

	// Create user in the queue
	result = queries.CreateUserQueue(user)
	if result.Error != nil || result.RowsAffected == 0 {
		handleError(c, 500, "Internal server error while creating new user", utils.ErrSaveData, result.Error)
		return
	}

	// Set a verify pending JWT cookie
	verifyPeddingToken, err := generateEmailVerifyJWT(user.ID)
	if err != nil {
		handleError(c, 500, "Internal server error while generating verify pending token", utils.ErrGenerateToken, err)
		return
	}
	c.SetCookie("verify_pedding", verifyPeddingToken, 15*60, "/", "", false, false)

	utils.SimpleResponse(c, 200, "Signup successful, please verify email", nil, nil)
}

func handleError(c *gin.Context, statusCode int, message string, errCode string, err error) {
	if err != nil {
		c.Error(err)
	}
	utils.SimpleResponse(c, statusCode, message, errCode, err.Error())
}

func hashUserPassword(password string) (string, error) {
	return encryption.HashPassword(password)
}

func sendVerificationEmail(email, username, verifyToken string) error {
	data := struct {
		VerifyURL string
		UserName  string
	}{
		VerifyURL: utils.BackendURL + "/auth/email/verify/" + verifyToken,
		UserName:  username,
	}

	var emailBody bytes.Buffer
	t, err := template.New("Email verification").ParseFiles("template/email_verification.html")
	if err != nil {
		return err
	}

	err = t.ExecuteTemplate(&emailBody, "email_verification.html", data)
	if err != nil {
		return err
	}

	return utils.SendEmail(email, "Verify your email", emailBody.String())
}

func generateEmailVerifyJWT(userID uint64) (string, error) {
	expiration := time.Now().Add(15 * time.Minute)
	return encryption.GenerateNewJwtToken(userID, []string{"pedding"}, expiration)
}
