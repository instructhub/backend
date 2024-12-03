package models

import "time"

// Users data type
type User struct {
	ID        uint64     `json:"id" bson:"id" binding:"required"`
	Avatar    string     `json:"avatar,omitempty" bson:"avatar,omitempty"`
	Username  string     `json:"username" bson:"username" binding:"required"` // Unique
	Nickname  string     `json:"nickname" bson:"nickname" binding:"required,max=50"`
	Email     string     `json:"email" bson:"email" binding:"required,email"`  // Unique
	Password  string     `json:"password,omitempty" bson:"password,omitempty"` // Hashed password, omit for OAuth users
	Providers []Provider `json:"providers" bson:"providers"`
	Verify    bool       `json:"verify" bson:"verify"`
	CreatedAt time.Time  `json:"created_at" bson:"created_at" binding:"required"`
	UpdatedAt time.Time  `json:"updated_at" bson:"updated_at" binding:"required"`
}

type UserProfile struct {
	ID       uint64 `json:"id" bson:"id" binding:"required"`
	Avatar   string `json:"avatar,omitempty" bson:"avatar,omitempty"`
	Username string `json:"username" bson:"username" binding:"required"`
	Email    string `json:"email" bson:"email" binding:"required,email"`
	Verify   bool   `json:"verify" bson:"verify"`
}

// Oauth privder type
type Provider struct {
	Provider string `json:"provider,omitempty" bson:"provider,omitempty"` // "google", "facebook", etc.
	OAuthID  string `json:"oauth_id,omitempty" bson:"oauth_id,omitempty"` // Unique OAuth user ID
}
