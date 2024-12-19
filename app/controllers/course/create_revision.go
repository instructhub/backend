package courses

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/app/queries"
	"github.com/instructhub/backend/pkg/encryption"
	git "github.com/instructhub/backend/pkg/gitea"
	"github.com/instructhub/backend/pkg/utils"
)

type CourseItemRequest struct {
	ID       *string           `json:"id" binding:"omitempty,number"`
	StageID  *string           `json:"stage_id,omitempty" binding:"omitempty,number"`
	Position int               `json:"position" binding:"required"`
	Type     models.CourseType `json:"type" binding:"number"`
	Name     string            `json:"name" binding:"max=50"`
	Updated  *bool             `json:"updated,omitempty" binding:"omitempty"`
	Content  *string           `json:"content,omitempty" binding:"omitempty,max=100000"`
}

type CourseStageRequest struct {
	ID       *string `json:"id" binding:"omitempty,number"`
	Position int     `json:"position" binding:"required,number"`
	Name     string  `json:"name" binding:"required,max=30"`

	CourseItems []CourseItemRequest `json:"course_items" binding:"max=20,dive"`
}

type UpdateRequestCourse struct {
	Stages      []CourseStageRequest `json:"stages" binding:"min=1,max=10,dive"`
	Description string               `json:"description" binding:"required,max=100"`
}

// CreateNewRevision handles course content updates
func CreateNewRevision(c *gin.Context) {
	var request UpdateRequestCourse

	// Validate request body
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.FullyResponse(c, 400, "Invalid request", utils.ErrBadRequest, err.Error())
		return
	}

	// Parse course ID and get user ID from context
	courseID, userID, err := getCourseIDAndUserID(c)
	if err != nil {
		utils.FullyResponse(c, 400, "Error getting course or user ID", utils.ErrBadRequest, err.Error())
		return
	}

	// Sort stages and items by position
	sortStagesAndItems(&request)

	// Fetch old course data
	oldCourseData, result := queries.GetCourseWithDetails(courseID)
	if result.RowsAffected == 0 {
		utils.FullyResponse(c, 404, "Course not exist", utils.ErrCourseNotExist, nil)
		return
	} else if result.Error != nil {
		c.Error(result.Error)
		utils.FullyResponse(c, 500, "Error fetching course", utils.ErrGetData, nil)
		return
	}

	// Validate stage and item positions
	if err := validatePositions(request.Stages); err != nil {
		utils.FullyResponse(c, 400, err.Error(), utils.ErrBadRequest, nil)
		return
	}

	// Process updates for course items
	request, updateFiles, err := processCourseItems(request, oldCourseData)
	if err != nil {
		utils.FullyResponse(c, 500, "Error processing course items", utils.ErrSaveCourseFile, nil)
		return
	}

	// Prepare and encode course data
	courseDataJson, err := encodeCourseData(request)
	if err != nil {
		utils.FullyResponse(c, 500, "Error encoding course data", utils.ErrParseData, nil)
		return
	}

	// Add or update course data in git
	courseRevision, err := updateCourseInGit(courseID, updateFiles, request.Description, courseDataJson, userID)
	if err != nil {
		utils.FullyResponse(c, 500, "Error updating course in git", utils.ErrSaveCourseFile, nil)
		return
	}

	// Create a course revision revision
	err = createCourseRevision(*courseRevision)
	if err != nil {
		utils.FullyResponse(c, 500, "Error creating course revision", utils.ErrSaveData, nil)
		return
	}

	utils.FullyResponse(c, 201, "Successfully created a new revision request", nil, courseRevision)
}

// getCourseIDAndUserID retrieves the course ID from the URL parameters and the user ID from the context
func getCourseIDAndUserID(c *gin.Context) (uint64, uint64, error) {
	courseID, err := strconv.ParseUint(c.Param("courseID"), 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid course ID")
	}

	ContextUserID, exists := c.Get("userID")
	if !exists {
		return 0, 0, fmt.Errorf("userID not found in context")
	}
	userID := ContextUserID.(uint64)
	return courseID, userID, nil
}

// sortStagesAndItems sorts the stages and course items by position
func sortStagesAndItems(request *UpdateRequestCourse) {
	sort.Slice(request.Stages, func(i, j int) bool {
		return request.Stages[i].Position < request.Stages[j].Position
	})
	for i := range request.Stages {
		sort.Slice(request.Stages[i].CourseItems, func(x, y int) bool {
			return request.Stages[i].CourseItems[x].Position < request.Stages[i].CourseItems[y].Position
		})
	}
}

// validatePositions ensures that stage and item positions are correctly ordered
func validatePositions(stages []CourseStageRequest) error {
	for i, stage := range stages {
		if stage.Position != i+1 {
			return fmt.Errorf("invalid stage position at index %d, expected %d but got %d", i, i+1, stage.Position)
		}
		for j, item := range stage.CourseItems {
			if item.Position != j+1 {
				return fmt.Errorf("invalid item position at stage %d, item index %d, expected %d but got %d", stage.Position, j, j+1, item.Position)
			}
		}
	}
	return nil
}

// processCourseItems processes course items, identifies new and deleted items, and prepares update files
func processCourseItems(request UpdateRequestCourse, oldCourseData models.Course) (UpdateRequestCourse, []git.File, error) {
	updateFiles := []git.File{}
	newCourseItems := make(map[string]bool)

	// Identify new and updated items
	for _, stage := range request.Stages {
		for _, item := range stage.CourseItems {
			if item.ID == nil {
				continue
			}
			newCourseItems[*item.ID] = true
		}
	}

	// Identify deleted items
	for _, stage := range *oldCourseData.CourseStages {
		for _, item := range *stage.CourseItems {
			if _, exists := newCourseItems[utils.Uint64ToStr(item.ID)]; !exists {
				updateFiles = append(updateFiles, git.File{
					Path:      utils.Uint64ToStr(item.ID),
					Operation: git.OperationDelete,
				})
			}
		}
	}

	// FIXME: Need to check if the id really new
	// Process new and updated course items
	for i := range request.Stages {
		stage := &request.Stages[i]
		if stage.ID == nil {
			stageID := encryption.GenerateID()
			stage.ID = utils.Uint64ToStrPtr(stageID)
		}

		for j := range stage.CourseItems {
			item := &stage.CourseItems[j]
			if item.ID == nil {
				itemID := encryption.GenerateID()
				item.ID = utils.Uint64ToStrPtr(itemID)
				updateFiles = append(updateFiles, git.File{
					Path:      *item.ID,
					Content:   encryption.Base64Encode(*item.Content),
					Operation: git.OperationCreate,
				})
			} else if item.Updated != nil && *item.Updated {
				updateFiles = append(updateFiles, git.File{
					Path:      *item.ID,
					Content:   encryption.Base64Encode(*item.Content),
					Operation: git.OperationUpdate,
				})
			}
			item.Content = nil
			item.Updated = nil
		}
	}
	return request, updateFiles, nil
}

// encodeCourseData marshals course data into JSON and encodes it as base64
func encodeCourseData(request UpdateRequestCourse) (string, error) {
	courseDataJson, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	return encryption.Base64Encode(string(courseDataJson)), nil
}

// updateCourseInGit updates the course data and files in the git repository
func updateCourseInGit(courseID uint64, updateFiles []git.File, description, courseDataJson string, userID uint64) (courseRevision *models.CourseRevision, err error) {
	updateFiles = append(updateFiles, git.File{
		Content:   courseDataJson,
		Path:      "course_data.json",
		Operation: git.OperationUpdate,
	})

	identity := gitea.Identity{
		Name:  utils.Uint64ToStr(userID),
		Email: git.GenerateCommmitEmail(userID),
	}

	branchID := encryption.GenerateID()

	modifyRequest := git.ModifyRequest{
		Author:    identity,
		Committer: identity,
		Files:     updateFiles,
		Message:   description,
		NewBranch: utils.Uint64ToStr(branchID),
	}

	// Perform the modification on git
	if err := git.ModifyMultipleFiles(utils.GiteaORGName, utils.Uint64ToStr(courseID), modifyRequest); err != nil {
		return nil, err
	}

	// Create pull request for the changes
	prOptions := gitea.CreatePullRequestOption{
		Head:  utils.Uint64ToStr(branchID),
		Base:  "en",
		Title: description,
	}

	pullRequest, _, err := git.GiteaClient.CreatePullRequest(utils.GiteaORGName, utils.Uint64ToStr(courseID), prOptions)

	courseRevision = &models.CourseRevision{
		ID:            encryption.GenerateID(),
		CourseID:      courseID,
		BranchID:      branchID,
		Description:   description,
		PullRequestID: int(pullRequest.Index),
		EditorID:      userID,
		Status:        models.RevisionOpen,
		UpdatedAt:     time.Now(),
		CreatedAt:     time.Now(),
	}

	return courseRevision, err
}

// createCourseRevision creates a revision entry for the course revision
func createCourseRevision(courseRevision models.CourseRevision) error {
	// Save the course revision revision
	result := queries.CreateNewCourseRevision(courseRevision)
	if result.RowsAffected == 0 {
		return fmt.Errorf("failed create new course revision")
	}
	return result.Error
}
