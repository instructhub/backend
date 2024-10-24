package models

import "time"

type Session struct {
	SecretKey string `json:"secret_key"`
	UserID    uint64 `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type AccessToken struct {
	Token string `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
