package models

import "time"

type User struct {
	ID        uint64    `json:"id" bson:"id" binding:"required"`
	Avatar    string    `json:"avatar,omitempty" bson:"avatar,omitempty"`
	Username  string    `json:"username" bson:"username" binding:"required"`  // Unique
	Email     string    `json:"email" bson:"email" binding:"required,email"`  // Unique
	Password  string    `json:"password,omitempty" bson:"password,omitempty"` // Hashed password, omit for OAuth users
	Provider  string    `json:"provider,omitempty" bson:"provider,omitempty"` // "google", "facebook", etc.
	OAuthID   string    `json:"oauth_id,omitempty" bson:"oauth_id,omitempty"` // Unique OAuth user ID
	CreatedAt time.Time `json:"created_at" bson:"created_at" binding:"required"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at" binding:"required"`
}
