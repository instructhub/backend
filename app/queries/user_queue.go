package queries

import (
	"github.com/instructhub/backend/app/models"
	db "github.com/instructhub/backend/pkg/database"
	"gorm.io/gorm"
)

// Get user by email
func GetUserQueueByEmail(email string) (user models.User,result *gorm.DB) {
	result = db.GetDB().Where("email = ?", email).First(&user)
	return user, result
}

// Get user by username
func GetUserQueueByUsername(username string) (user models.User,result *gorm.DB) {
	// Use Where to filter by username
	result = db.GetDB().Where("username = ?", username).First(&user)
	return user, result
}

// Get user by user ID
func GetUserQueueByID(id uint64) (user models.User,result *gorm.DB) {
	result = db.GetDB().Where("id = ?", id).First(&user)
	return user, result
}

// Create new user data
func CreateUserQueue(user models.User) *gorm.DB {
	result := db.GetDB().Create(&user)
	return result
}

func UpdateUserVerifyStatus(userID uint64, verify bool) *gorm.DB {
	// Update only specific fields (e.g., email) where id matches
	var user models.User
	result := db.GetDB().
		Model(&user).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"verify": verify,
		})

	return result
}

// Update user third-part oauth provider data
func AddUserProvider(oauthProvider models.OauthProvider) *gorm.DB {
	result := db.GetDB().Create(&oauthProvider)
	return result
}

// Get user and associated provider data
func GetUserAndProvider(userEmail string) (user models.User, result *gorm.DB) {
	// Preload the associated OAuth provider(s)
	user.Email = userEmail
	result = db.GetDB().
		Preload("OauthProviders").
		Where("email = ?", userEmail).
		First(&user)
	return user, result
}
