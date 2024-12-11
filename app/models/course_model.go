package models

import "time"

type Course struct {
	ID               uint64    `json:"id" bson:"id"`
	Creator          uint64    `json:"creator" bson:"creator"`
	Name            string    `json:"name" bson:"name"`
	Description string    `json:"description" bson:"description"`
	UpdatedAt              time.Time `json:"updated_at" bson:"updated_at"`
	CreateAt               time.Time `json:"create_at" bson:"create_at"`
}

type CourseImage struct {
	ImageLink string `json:"image_link" bson:"image_link"`
	ID uint64 `json:"id" bson:"id"`
	Craetor uint64 `json:"creator" bson:"creator"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type CourseFile struct {
	Content  string `json:"content" binding:"required,base64,max=1000000,min=10"`
	Stage    string `json:"stage" binding:"required,max=10,min=3"`
	Message  string `json:"message" binding:"max=50"`
}
