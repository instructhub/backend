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

type CourseStepRequest struct {
	ID       *string           `json:"id" binding:"omitempty,number"`
	ModuleID *string           `json:"module_id,omitempty" binding:"omitempty,number"`
	Position int               `json:"position" binding:"required"`
	Type     models.CourseType `json:"type" binding:"number"`
	Name     string            `json:"name" binding:"max=50"`
	Updated  *bool             `json:"updated,omitempty" binding:"omitempty"`
	Content  *string           `json:"content,omitempty" binding:"omitempty,base64,max=100000"`
}

type CourseModuleRequest struct {
	ID       *string `json:"id" binding:"omitempty,number"`
	Position int     `json:"position" binding:"required,number"`
	Name     string  `json:"name" binding:"required,max=30"`

	CourseSteps []CourseStepRequest `json:"course_steps" binding:"max=20,dive"`
}

type UpdateRequestCourse struct {
	Modules     []CourseModuleRequest `json:"modules" binding:"min=1,max=10,dive"`
	Description string                `json:"description" binding:"required,max=100"`
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

	// Sort modules and steps by position
	sortModulesAndSteps(&request)

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

	// Validate module and step positions
	if err := validatePositions(request.Modules); err != nil {
		utils.FullyResponse(c, 400, err.Error(), utils.ErrBadRequest, nil)
		return
	}

	// Process updates for course steps
	request, updateFiles, err := processCourseSteps(request, oldCourseData)
	if err != nil {
		utils.FullyResponse(c, 500, "Error processing course steps", utils.ErrSaveCourseFile, nil)
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

// sortModulesAndSteps sorts the modules and course steps by position
func sortModulesAndSteps(request *UpdateRequestCourse) {
	sort.Slice(request.Modules, func(i, j int) bool {
		return request.Modules[i].Position < request.Modules[j].Position
	})
	for i := range request.Modules {
		sort.Slice(request.Modules[i].CourseSteps, func(x, y int) bool {
			return request.Modules[i].CourseSteps[x].Position < request.Modules[i].CourseSteps[y].Position
		})
	}
}

// validatePositions ensures that module and step positions are correctly ordered
func validatePositions(modules []CourseModuleRequest) error {
	for i, module := range modules {
		if module.Position != i+1 {
			return fmt.Errorf("invalid module position at index %d, expected %d but got %d", i, i+1, module.Position)
		}
		for j, step := range module.CourseSteps {
			if step.Position != j+1 {
				return fmt.Errorf("invalid step position at module %d, step index %d, expected %d but got %d", module.Position, j, j+1, step.Position)
			}
		}
	}
	return nil
}

// processCourseSteps processes course steps, identifies new and deleted steps, and prepares update files
func processCourseSteps(request UpdateRequestCourse, oldCourseData models.Course) (UpdateRequestCourse, []git.File, error) {
	updateFiles := []git.File{}
	newCourseSteps := make(map[string]bool)

	// Identify new and updated steps
	for _, module := range request.Modules {
		for _, step := range module.CourseSteps {
			if step.ID == nil {
				continue
			}
			newCourseSteps[*step.ID] = true
		}
	}

	// Identify deleted steps
	for _, module := range *oldCourseData.CourseModules {
		for _, step := range *module.CourseSteps {
			if _, exists := newCourseSteps[utils.Uint64ToStr(step.ID)]; !exists {
				updateFiles = append(updateFiles, git.File{
					Path:      utils.Uint64ToStr(step.ID),
					Operation: git.OperationDelete,
				})
			}
		}
	}

	// FIXME: Need to check if the id really new
	// Process new and updated course steps
	for i := range request.Modules {
		module := &request.Modules[i]
		if module.ID == nil {
			moduleID := encryption.GenerateID()
			module.ID = utils.Uint64ToStrPtr(moduleID)
		}

		for j := range module.CourseSteps {
			step := &module.CourseSteps[j]
			if step.ID == nil {
				stepID := encryption.GenerateID()
				step.ID = utils.Uint64ToStrPtr(stepID)
				updateFiles = append(updateFiles, git.File{
					Path:      *step.ID,
					Content:   *step.Content,
					Operation: git.OperationCreate,
				})
			} else if step.Updated != nil && *step.Updated {
				updateFiles = append(updateFiles, git.File{
					Path:      *step.ID,
					Content:   *step.Content,
					Operation: git.OperationUpdate,
				})
			}
			step.Content = nil
			step.Updated = nil
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
