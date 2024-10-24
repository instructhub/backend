package models

import "time"

type User struct {
	ID        uint64    `json:"id" bson:"id" binding:"required"`
	Username  string    `json:"username" bson:"username" binding:"required"` // Unique
	Email     string    `json:"email" bson:"email" binding:"required,email"` // Unique
	Password  string    `json:"password" bson:"password" binding:"required"` // Hashed password
	CreatedAt time.Time `json:"created_at" bson:"created_at" binding:"required"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at" binding:"required"`
}
