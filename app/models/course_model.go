package models

import "time"

type Course struct {
	CourseID               uint64    `json:"course_id" bson:"course_id"`
	CourseCreator          uint64    `json:"course_creator" bson:"course_creator"`
	CourseTitle            string    `json:"course_title" bson:"course_tile"`
	CourseShortDescription string    `json:"course_short_description" bson:"course_short_description"`
	UpdatedAt              time.Time `json:"updated_at" bson:"updated_at"`
	CreateAt               time.Time `json:"create_at" bson:"create_at"`
}

type CourseImage struct {
	ImageLink string `json:"image_link" bson:"image_link"`
	CourseID uint64 `json:"course_id" bson:"course_id"`
	Craetor uint64 `json:"creator" bson:"creator"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type CourseFile struct {
	Content  string `json:"content" binding:"required,base64,max=1000000,min=10"`
	Stage    string `json:"stage" binding:"required,max=10,min=3"`
	Message  string `json:"message" binding:"max=50"`
}
