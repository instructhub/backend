package models

import (
	"strings"
	"time"

	db "github.com/instructhub/backend/pkg/database"
)

func init() {
	db.GetDB().AutoMigrate(&Course{})
	db.GetDB().AutoMigrate(&CourseStage{})
	db.GetDB().AutoMigrate(&CourseItem{})
	db.GetDB().AutoMigrate(&CourseImage{})
	db.GetDB().AutoMigrate(&CourseRevision{})
}

// Course type / table
type Course struct {
	ID          uint64    `json:"id" gorm:"primaryKey"`
	Creator     uint64    `json:"creator" gorm:"not null"`
	Name        string    `json:"name" gorm:"not null;size:255"`
	Description string    `json:"description" gorm:"type:text"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`

	CourseStages *[]CourseStage `gorm:"foreignKey:CourseID"`
}

// CourseStage type / table
type CourseStage struct {
	ID        uint64    `json:"id" gorm:"primaryKey"`
	CourseID  uint64    `json:"course_id" gorm:"not null;index"`
	Position  int       `json:"position"`
	Name      string    `json:"name" gorm:"not null;size:255"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	Active    bool      `json:"active"`

	// Foreign key
	Course      Course        `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	CourseItems *[]CourseItem `gorm:"foreignKey:StageID"`
}

type CourseType int8

const (
	CourseTextContent CourseType = iota
	CourseVideoContent
	CourseQuestionContent
)

var (
	courseTypeMap = map[string]CourseType{
		"text":     CourseTextContent,
		"video":    CourseVideoContent,
		"question": CourseQuestionContent,
	}
)

func ParseStringToCourseType(str string) (CourseType, bool) {
	c, ok := courseTypeMap[strings.ToLower(str)]
	return c, ok
}

// CourseItem type / table
type CourseItem struct {
	ID        uint64     `json:"id" gorm:"primaryKey"`
	StageID   uint64     `json:"stage_id" gorm:"not null;index"`
	Position  int        `json:"position"`
	Type      CourseType `json:"type" gorm:"not null"`
	Name      string     `json:"name" gorm:"not null;size:255"`
	Active    bool       `json:"active"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`

	// Foreign key
	Stage CourseStage `gorm:"foreignKey:StageID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
}

// CourseImage type / table
type CourseImage struct {
	ID        uint64    `json:"id" gorm:"primaryKey"`
	ImageLink string    `json:"image_link" gorm:"not null;size:512"`
	Creator   uint64    `json:"creator" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type HistoryStatus int8

const (
	HistoryOpen HistoryStatus = iota
	HistoryClose
	HistoryMerged
	// No one can leave comment or update anything
	HistoryLock
)

type CourseRevision struct {
	ID            uint64        `json:"id" gorm:"primaryKey"`
	CourseID      uint64        `json:"course_id" gorm:"not null;index"`
	BranchID      uint64        `json:"branch_id"`
	PullRequestID int           `json:"pull_request_id"`
	Description   string        `json:"description"`
	Status        HistoryStatus `json:"status"`
	EditorID      *uint64       `json:"editor_id" gorm:"index"`
	UpdatedAt     time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt     time.Time     `json:"created_at" gorm:"autoCreateTime"`

	Course Course `gorm:"foreignKey:CourseID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	User   User   `gorm:"foreignKey:EditorID;references:ID;constraint:OnDelete:SET NULL"`
}
