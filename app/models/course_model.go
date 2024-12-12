package models

import (
	"time"

	db "github.com/instructhub/backend/pkg/database"
)

func init() {
	db.GetDB().AutoMigrate(&Course{})
	db.GetDB().AutoMigrate(&CourseStage{})
	db.GetDB().AutoMigrate(&CourseItem{})
	db.GetDB().AutoMigrate(&CourseImage{})
}

// Course type / table
type Course struct {
	ID          uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	Creator     uint64    `json:"creator" gorm:"not null"`
	Name        string    `json:"name" gorm:"not null;size:255"`
	Description string    `json:"description" gorm:"type:text"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// CourseStage type / table
type CourseStage struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	CourseID  uint      `json:"course_id" gorm:"not null;index"`
	Name      string    `json:"name" gorm:"not null;size:255"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Foreign key
	Course Course `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
}

type CourseType int8

const (
	CourseTextContent    CourseType = iota
	CourseVideoContent
	CourseQuestionContent
)

// CourseItem type / table
type CourseItem struct {
	ID        uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	StageID   uint       `json:"stage_id" gorm:"not null;index"`
	Type      CourseType `json:"type" gorm:"not null"`
	Name      string     `json:"name" gorm:"not null;size:255"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`

	// Foreign key
	Stage CourseStage `gorm:"foreignKey:StageID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
}

// CourseImage type / table
type CourseImage struct {
	ID        uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	ImageLink string    `json:"image_link" gorm:"not null;size:512"`
	Creator   uint64    `json:"creator" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}
