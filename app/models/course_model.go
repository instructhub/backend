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
	ID          uint64    `json:"id,string" gorm:"primaryKey"`
	CreatorID   uint64    `json:"creator_id,string" gorm:"not null"`
	Name        string    `json:"name" gorm:"not null;size:255"`
	Description string    `json:"description" gorm:"type:text"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`

	CourseStages *[]CourseStage `json:"course_stages,omitempty" gorm:"foreignKey:CourseID"`
}

// CourseStage type / table
type CourseStage struct {
	ID        uint64    `json:"id,string" gorm:"primaryKey"`
	CourseID  uint64    `json:"course_id,string" gorm:"not null;index"`
	Position  int       `json:"position"`
	Name      string    `json:"name" gorm:"not null;size:255"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	Active    *bool     `json:"active" gorm:"default:true"`

	// Foreign key
	Course      *Course       `json:"course,omitempty" gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	CourseItems *[]CourseItem `json:"course_items,omitempty" gorm:"foreignKey:StageID"`
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
	ID        uint64     `json:"id,string" gorm:"primaryKey"`
	StageID   uint64     `json:"stage_id,string" gorm:"not null;index"`
	Position  int        `json:"position"`
	Type      CourseType `json:"type" gorm:"not null"`
	Name      string     `json:"name" gorm:"not null;size:255"`
	Active    *bool      `json:"active" gorm:"default:true"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`

	// Foreign key
	Stage *CourseStage `json:"stage,omitempty" gorm:"foreignKey:StageID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
}

// CourseImage type / table
type CourseImage struct {
	ID        uint64    `json:"id,string" gorm:"primaryKey"`
	ImageLink string    `json:"image_link" gorm:"not null;size:512"`
	CreatorID uint64    `json:"creator_id,string" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type RevisionStatus int8

const (
	RevisionOpen RevisionStatus = iota
	RevisionClose
	RevisionMerged
	// No one can leave comment or update anything
	RevisionLock
)

type CourseRevision struct {
	ID            uint64         `json:"id,string" gorm:"primaryKey"`
	CourseID      uint64         `json:"course_id,string" gorm:"not null;index"`
	BranchID      uint64         `json:"branch_id,string"`
	PullRequestID int            `json:"pull_request_id"`
	Description   string         `json:"description"`
	Status        RevisionStatus `json:"status"`
	EditorID      uint64         `json:"editor_id,string" gorm:"index"`
	ApproverID    *uint64        `json:"approver_id,string" gorm:"index"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime"`

	Course   *Course `json:"course,omitempty" gorm:"foreignKey:CourseID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Editor   *User   `json:"editor,omitempty" gorm:"foreignKey:EditorID;references:ID;constraint:OnDelete:SET NULL"`
	Approver *User   `json:"approver,omitempty" gorm:"foreignKey:ApproverID;references:ID;constraint:OnDelete:SET NULL"`
}
