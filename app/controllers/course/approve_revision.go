package courses

import (
	"encoding/json"
	"fmt"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/app/queries"
	"github.com/instructhub/backend/pkg/encryption"
	git "github.com/instructhub/backend/pkg/gitea"
	"github.com/instructhub/backend/pkg/utils"
	"gorm.io/gorm"
)

// FIXME: Check user premission have permission to approve
// ApproveRevision handles approving a course revision
func ApproveRevision(c *gin.Context) {
	courseID, _, revisionID, err := parseIDs(c)
	if err != nil {
		utils.FullyResponse(c, 400, "Invalid course or revision ID", utils.ErrBadRequest, err.Error())
		return
	}

	// Fetch revision details
	revision, result := queries.GetCourseRevision(courseID, revisionID)
	if result.Error == gorm.ErrRecordNotFound {
		utils.FullyResponse(c, 404, "Revision not found", utils.ErrCourseNotExist, nil)
		return
	} else if result.Error != nil {
		utils.ServerErrorResponse(c, 500, "Error fetching course data", utils.ErrGetData, err)
		return
	}

	// Ensure revision is not already merged
	if revision.Status == models.RevisionMerged {
		utils.FullyResponse(c, 400, "Revision already merged", utils.ErrAlreadyMerged, nil)
		return
	}

	// Fetch course data from git
	revisionData, err := fetchCourseDataFromGit(revision)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error fetching course data", utils.ErrGetData, err)
		return
	}

	// Parse course data into the update request
	var updateRequest UpdateRequestCourse
	err = json.Unmarshal([]byte(revisionData), &updateRequest)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error unmarshaling data", utils.ErrUnmarshal, err)
		return
	}

	// Prepare course modules and steps for update
	needUpdateModules, needUpdateSteps, needCreateModules, needCreateSteps := prepareCourseData(courseID, revision, updateRequest)

	// Update modules and steps in the database
	err = updateCourseModules(needUpdateModules)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error updating module data", utils.ErrSaveData, err)
		return
	}

	err = updateCourseSteps(needUpdateSteps)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error updating step data", utils.ErrSaveData, err)
		return
	}

	// Create new modules and steps
	err = createCourseModules(needCreateModules)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error creating new module data", utils.ErrSaveData, err)
		return
	}

	err = createCourseSteps(needCreateSteps)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error creating new step data", utils.ErrSaveData, err)
		return
	}

	// Update revision status and merge the pull request
	err = mergeRevisionAndPullRequest(revision, courseID)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error merging revision and pull request", utils.ErrSaveData, nil)
		return
	}

	utils.FullyResponse(c, 200, "Revision successfully approved", nil, updateRequest)
}

// parseIDs extracts courseID, userID, and revisionID from context and validates them
func parseIDs(c *gin.Context) (uint64, uint64, uint64, error) {
	courseID, err := utils.StrToUint64(c.Param("courseID"))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid course ID")
	}

	revisionID, err := utils.StrToUint64(c.Param("revisionID"))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid revision ID")
	}

	ContextUserID, exists := c.Get("userID")
	if !exists {
		return 0, 0, 0, fmt.Errorf("user ID not found in context")
	}

	userID := ContextUserID.(uint64)
	return courseID, userID, revisionID, nil
}

// fetchCourseDataFromGit retrieves the course data from Git
func fetchCourseDataFromGit(revision models.CourseRevision) (string, error) {
	revisionChangeFile, _, err := git.GiteaClient.GetContents(utils.GiteaORGName, utils.Uint64ToStr(revision.CourseID), utils.Uint64ToStr(revision.BranchID), "/course_data.json")
	if err != nil {
		return "", err
	}

	revisionChangeDataString, err := encryption.Base64Decode(*revisionChangeFile.Content)
	if err != nil {
		return "", err
	}

	return revisionChangeDataString, nil
}

// prepareCourseData prepares modules and steps for update or creation
func prepareCourseData(courseID uint64, revision models.CourseRevision, updateRequest UpdateRequestCourse) (
	[]models.CourseModule, []models.CourseStep, []models.CourseModule, []models.CourseStep) {

	oldCourseModules := map[string]models.CourseModule{}
	oldCourseSteps := map[string]models.CourseStep{}
	newCourseModules := map[string]CourseModuleRequest{}
	newCourseSteps := map[string]CourseStepRequest{}

	needUpdateModules := []models.CourseModule{}
	needUpdateSteps := []models.CourseStep{}
	needCreateModules := []models.CourseModule{}
	needCreateSteps := []models.CourseStep{}

	// Populate old data
	for _, module := range *revision.Course.CourseModules {
		oldCourseModules[utils.Uint64ToStr(module.ID)] = module
		for _, step := range *module.CourseSteps {
			oldCourseSteps[utils.Uint64ToStr(step.ID)] = step
		}
	}

	// Process new course modules and steps
	for _, module := range updateRequest.Modules {
		newCourseModules[*module.ID] = module
		if _, ok := oldCourseModules[*module.ID]; !ok {
			createModule := models.CourseModule{
				ID:        utils.StrToUint64NoError(*module.ID),
				CourseID:  courseID,
				Position:  module.Position,
				Name:      module.Name,
				UpdatedAt: time.Now(),
				CreatedAt: time.Now(),
				Active:    utils.BoolPtr(true),
			}
			needCreateModules = append(needCreateModules, createModule)
		}

		for _, step := range module.CourseSteps {
			newCourseSteps[*step.ID] = step
			if _, ok := oldCourseSteps[*step.ID]; !ok {
				createStep := models.CourseStep{
					ID:        utils.StrToUint64NoError(*step.ID),
					ModuleID:  utils.StrToUint64NoError(*module.ID),
					Position:  step.Position,
					Name:      step.Name,
					Type:      step.Type,
					Active:    utils.BoolPtr(true),
					UpdatedAt: time.Now(),
					CreatedAt: time.Now(),
				}
				needCreateSteps = append(needCreateSteps, createStep)
			}
		}
	}

	// Identify modules and steps to update or delete
	for i := range *revision.Course.CourseModules {
		module := &(*revision.Course.CourseModules)[i]
		if newModule, ok := newCourseModules[utils.Uint64ToStr(module.ID)]; !ok {
			// Mark for deletion
			module.Active = utils.BoolPtr(false)
			needUpdateModules = append(needUpdateModules, *module)
		} else if newModule.Position != module.Position || newModule.Name != module.Name {
			// Mark for update
			module.Position = newModule.Position
			module.Name = newModule.Name
			module.UpdatedAt = time.Now()
			needUpdateModules = append(needUpdateModules, *module)
		}

		for j := range *module.CourseSteps {
			step := &(*module.CourseSteps)[j]
			if newStep, ok := newCourseSteps[utils.Uint64ToStr(step.ID)]; !ok {
				// Mark for deletion
				step.Active = utils.BoolPtr(false)
				needUpdateSteps = append(needUpdateSteps, *step)
			} else if step.Name != newStep.Name || step.Position != newStep.Position {
				// Mark for update
				step.Position = newStep.Position
				step.Name = newStep.Name
				step.UpdatedAt = time.Now()
				needUpdateSteps = append(needUpdateSteps, *step)
			}
		}
	}

	return needUpdateModules, needUpdateSteps, needCreateModules, needCreateSteps
}

// updateCourseModules updates existing course modules in the database
func updateCourseModules(needUpdateModules []models.CourseModule) error {
	for _, module := range needUpdateModules {
		result := queries.UpdateCourseModule(module)
		if result.Error != nil || result.RowsAffected == 0 {
			return fmt.Errorf("failed to update module data")
		}
	}
	return nil
}

// updateCourseSteps updates existing course steps in the database
func updateCourseSteps(needUpdateSteps []models.CourseStep) error {
	for _, step := range needUpdateSteps {
		result := queries.UpdateCourseStep(step)
		if result.Error != nil || result.RowsAffected == 0 {
			return fmt.Errorf("failed to update step data")
		}
	}
	return nil
}

// createCourseModules creates new course modules in the database
func createCourseModules(needCreateModules []models.CourseModule) error {
	if len(needCreateModules) == 0 {
		return nil
	}
	result := queries.CreateCourseModules(needCreateModules)
	if result.Error != nil || result.RowsAffected == 0 {
		return fmt.Errorf("failed to create module data")
	}
	return nil
}

// createCourseSteps creates new course steps in the database
func createCourseSteps(needCreateSteps []models.CourseStep) error {
	if len(needCreateSteps) == 0 {
		return nil
	}
	result := queries.CreateCourseSteps(needCreateSteps)
	if result.Error != nil || result.RowsAffected == 0 {
		return fmt.Errorf("failed to create step data")
	}
	return nil
}

// mergeRevisionAndPullRequest merges the revision and the pull request
func mergeRevisionAndPullRequest(revision models.CourseRevision, courseID uint64) error {
	revision.UpdatedAt = time.Now()
	revision.Status = models.RevisionMerged
	result := queries.UpdateCourseRevision(revision)
	if result.Error != nil || result.RowsAffected == 0 {
		return fmt.Errorf("failed to update revision data")
	}

	mergeOptions := gitea.MergePullRequestOption{
		Style: "merge",
	}
	_, _, err := git.GiteaClient.MergePullRequest(utils.GiteaORGName, utils.Uint64ToStr(courseID), int64(revision.PullRequestID), mergeOptions)
	if err != nil {
		return fmt.Errorf("failed to merge pull request")
	}
	return nil
}
