package models

import (
	"time"

	db "github.com/instructhub/backend/pkg/database"
)

func init() {
	db.GetDB().AutoMigrate(&User{})
	db.GetDB().AutoMigrate(&OauthProvider{})
}

// Users data type / table
type User struct {
	ID          uint64    `json:"id" gorm:"primaryKey" binding:"required"`
	Avatar      *string   `json:"avatar,omitempty"`
	Username    string    `json:"username" gorm:"unique" binding:"required"` // Unique
	DisplayName string    `json:"display_name" binding:"required,max=50"`
	Email       string    `json:"email" gorm:"unique" binding:"required,email"` // Unique
	Password    string    `json:"password,omitempty"`                           // Hashed password, omit for OAuth users
	Verify      bool      `json:"verify"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoUpdateTime" binding:"required"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoCreateTime" binding:"required"`

	OauthProviders *[]OauthProvider `gorm:"foreignKey:UserID"`
}

type Provider int8

const (
	ProviderGoogle Provider = iota
	ProviderGithub
	ProviderGitlab
)

// Oauth privder type / table
type OauthProvider struct {
	ID        uint64    `json:"id" gorm:"primaryKey"`
	UserID    uint64    `json:"user_id" gorm:"not null;index"`
	Provider  Provider  `json:"provider" gorm:"not null"`
	OAuthID   string    `json:"oauth_id" gorm:"unique;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Foreign key
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
}

// User data for user when it need to know thier personal profile
type UserProfile struct {
	ID        uint64    `json:"id" binding:"required"`
	Avatar    string    `json:"avatar,omitempty"`
	Username  string    `json:"username" binding:"required"`
	Email     string    `json:"email" binding:"required,email"`
	Verify    bool      `json:"verify" `
	CreatedAt time.Time `json:"created_at" binding:"required"`
	UpdatedAt time.Time `json:"updated_at" binding:"required"`
}
