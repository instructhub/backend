package models

import "time"

// Session type
type Session struct {
	SecretKey string `json:"secret_key" bson:"secret_key"`
	UserAgent string `json:"user_agent" bson:"user_agent"`
	UserID    uint64 `json:"user_id" bson:"user_id"`
	ExpiresAt time.Time `json:"expires_at" bson:"expires_at"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

// Access token type
type AccessToken struct {
	Token string `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
