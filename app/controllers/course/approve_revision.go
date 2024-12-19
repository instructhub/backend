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

	// Prepare course stages and items for update
	needUpdateStages, needUpdateItems, needCreateStages, needCreateItems := prepareCourseData(courseID, revision, updateRequest)

	// Update stages and items in the database
	err = updateCourseStages(needUpdateStages)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error updating stage data", utils.ErrSaveData, err)
		return
	}

	err = updateCourseItems(needUpdateItems)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error updating item data", utils.ErrSaveData, err)
		return
	}

	// Create new stages and items
	err = createCourseStages(needCreateStages)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error creating new stage data", utils.ErrSaveData, err)
		return
	}

	err = createCourseItems(needCreateItems)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error creating new item data", utils.ErrSaveData, err)
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

// prepareCourseData prepares stages and items for update or creation
func prepareCourseData(courseID uint64, revision models.CourseRevision, updateRequest UpdateRequestCourse) (
	[]models.CourseStage, []models.CourseItem, []models.CourseStage, []models.CourseItem) {

	oldCourseStages := map[string]models.CourseStage{}
	oldCourseItems := map[string]models.CourseItem{}
	newCourseStages := map[string]CourseStageRequest{}
	newCourseItems := map[string]CourseItemRequest{}

	needUpdateStages := []models.CourseStage{}
	needUpdateItems := []models.CourseItem{}
	needCreateStages := []models.CourseStage{}
	needCreateItems := []models.CourseItem{}

	// Populate old data
	for _, stage := range *revision.Course.CourseStages {
		oldCourseStages[utils.Uint64ToStr(stage.ID)] = stage
		for _, item := range *stage.CourseItems {
			oldCourseItems[utils.Uint64ToStr(item.ID)] = item
		}
	}

	// Process new course stages and items
	for _, stage := range updateRequest.Stages {
		newCourseStages[*stage.ID] = stage
		if _, ok := oldCourseStages[*stage.ID]; !ok {
			createStage := models.CourseStage{
				ID:        utils.StrToUint64NoError(*stage.ID),
				CourseID:  courseID,
				Position:  stage.Position,
				Name:      stage.Name,
				UpdatedAt: time.Now(),
				CreatedAt: time.Now(),
				Active:    utils.BoolPtr(true),
			}
			needCreateStages = append(needCreateStages, createStage)
		}

		for _, item := range stage.CourseItems {
			newCourseItems[*item.ID] = item
			if _, ok := oldCourseItems[*item.ID]; !ok {
				createItem := models.CourseItem{
					ID:        utils.StrToUint64NoError(*item.ID),
					StageID:   utils.StrToUint64NoError(*stage.ID),
					Position:  item.Position,
					Name:      item.Name,
					Type:      item.Type,
					Active:    utils.BoolPtr(true),
					UpdatedAt: time.Now(),
					CreatedAt: time.Now(),
				}
				needCreateItems = append(needCreateItems, createItem)
			}
		}
	}

	// Identify stages and items to update or delete
	for i := range *revision.Course.CourseStages {
		stage := &(*revision.Course.CourseStages)[i]
		if newStage, ok := newCourseStages[utils.Uint64ToStr(stage.ID)]; !ok {
			// Mark for deletion
			stage.Active = utils.BoolPtr(false)
			needUpdateStages = append(needUpdateStages, *stage)
		} else if newStage.Position != stage.Position || newStage.Name != stage.Name {
			// Mark for update
			stage.Position = newStage.Position
			stage.Name = newStage.Name
			stage.UpdatedAt = time.Now()
			needUpdateStages = append(needUpdateStages, *stage)
		}

		for j := range *stage.CourseItems {
			item := &(*stage.CourseItems)[j]
			if newItem, ok := newCourseItems[utils.Uint64ToStr(item.ID)]; !ok {
				// Mark for deletion
				item.Active = utils.BoolPtr(false)
				needUpdateItems = append(needUpdateItems, *item)
			} else if item.Name != newItem.Name || item.Position != newItem.Position {
				// Mark for update
				item.Position = newItem.Position
				item.Name = newItem.Name
				item.UpdatedAt = time.Now()
				needUpdateItems = append(needUpdateItems, *item)
			}
		}
	}

	return needUpdateStages, needUpdateItems, needCreateStages, needCreateItems
}

// updateCourseStages updates existing course stages in the database
func updateCourseStages(needUpdateStages []models.CourseStage) error {
	for _, stage := range needUpdateStages {
		result := queries.UpdateCourseStage(stage)
		if result.Error != nil || result.RowsAffected == 0 {
			return fmt.Errorf("failed to update stage data")
		}
	}
	return nil
}

// updateCourseItems updates existing course items in the database
func updateCourseItems(needUpdateItems []models.CourseItem) error {
	for _, item := range needUpdateItems {
		result := queries.UpdateCourseItem(item)
		if result.Error != nil || result.RowsAffected == 0 {
			return fmt.Errorf("failed to update item data")
		}
	}
	return nil
}

// createCourseStages creates new course stages in the database
func createCourseStages(needCreateStages []models.CourseStage) error {
	if len(needCreateStages) == 0 {
		return nil
	}
	result := queries.CreateCourseStages(needCreateStages)
	if result.Error != nil || result.RowsAffected == 0 {
		return fmt.Errorf("failed to create stage data")
	}
	return nil
}

// createCourseItems creates new course items in the database
func createCourseItems(needCreateItems []models.CourseItem) error {
	if len(needCreateItems) == 0 {
		return nil
	}
	result := queries.CreateCourseItems(needCreateItems)
	if result.Error != nil || result.RowsAffected == 0 {
		return fmt.Errorf("failed to create item data")
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
