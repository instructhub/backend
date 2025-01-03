package models

import (
	"strings"
	"time"

	db "github.com/instructhub/backend/pkg/database"
	pq "github.com/lib/pq"
)

func init() {
	db.GetDB().AutoMigrate(&Course{})
	db.GetDB().AutoMigrate(&CourseModule{})
	db.GetDB().AutoMigrate(&CourseStep{})
	db.GetDB().AutoMigrate(&CourseImage{})
	db.GetDB().AutoMigrate(&CourseRevision{})
	db.GetDB().AutoMigrate(&CourseLandingPage{})
}

// Course type / table
type Course struct {
	ID          uint64    `json:"id,string" gorm:"primaryKey"`
	CreatorID   uint64    `json:"creator_id,string" gorm:"not null"`
	Name        string    `json:"name" gorm:"not null;size:255"`
	Description string    `json:"description" gorm:"type:text"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`

	CourseModules     *[]CourseModule    `json:"course_modules,omitempty" gorm:"foreignKey:CourseID"`
	CourseLandingPage *CourseLandingPage `json:"course_landing_page,omitempty" gorm:"foreignKey:CourseID"`
}

// CourseModule type / table
type CourseModule struct {
	ID        uint64    `json:"id,string" gorm:"primaryKey"`
	CourseID  uint64    `json:"course_id,string" gorm:"not null;index"`
	Position  int       `json:"position"`
	Name      string    `json:"name" gorm:"not null;size:255"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	Active    *bool     `json:"active" gorm:"default:true"`

	// Foreign key
	Course      *Course       `json:"course,omitempty" gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	CourseSteps *[]CourseStep `json:"course_steps,omitempty" gorm:"foreignKey:ModuleID"`
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

// CourseStep type / table
type CourseStep struct {
	ID        uint64     `json:"id,string" gorm:"primaryKey"`
	ModuleID  uint64     `json:"module_id,string" gorm:"not null;index"`
	Position  int        `json:"position"`
	Type      CourseType `json:"type" gorm:"not null"`
	Name      string     `json:"name" gorm:"not null;size:255"`
	Active    *bool      `json:"active" gorm:"default:true"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`

	// Foreign key
	Module *CourseModule `json:"module,omitempty" gorm:"foreignKey:ModuleID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
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

type ProgessState int8

const (
	NotStart ProgessState = iota
	Pause
	InProgess
	Finish
	Restart
)

type CourseProgess struct {
	ID            uint64       `json:"id,string" gorm:"primaryKey"`
	CourseID      uint64       `json:"course_id,string" gorm:"not null;index"`
	LearnerID     uint64       `json:"learner_id,string" gorm:"not null"`
	State         ProgessState `json:"state"`
	CurrentModule uint8        `json:"current_module"`
	CurrentStep   uint8        `json:"current_step"`
	FinishAt      time.Time    `json:"finish_at"`
	UpdatedAt     time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
	StartedAt     time.Time    `json:"stated_at" gorm:"autoCreateTime"`

	Course  *Course `json:"course,omitempty" gorm:"foreignKey:CourseID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Learner *User   `json:"editor,omitempty" gorm:"foreignKey:LearnerID;references:ID;constraint:OnDelete:SET NULL"`
}

type CourseLandingPage struct {
	CourseID       uint64          `json:"course_id,string" gorm:"not null"`
	Description    *string         `json:"description,omitempty"`
	ImageURL       *string         `json:"image_url,omitempty"`
	VideoURL       *string         `json:"video_url,omitempty"`
	SEOKeywords    *pq.StringArray `json:"seo_keywords,omitempty" gorm:"type:text[]"`
	Outcomes       *pq.StringArray `json:"outcomes,omitempty" gorm:"type:text[]"`
	Prerequisites  *pq.StringArray `json:"prerequisites,omitempty" gorm:"type:text[]"`
	TargetAudience *pq.StringArray `json:"target_audience,omitempty" gorm:"type:text[]"`

	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	Course *Course `json:"course,omitempty" gorm:"foreignKey:CourseID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
}
