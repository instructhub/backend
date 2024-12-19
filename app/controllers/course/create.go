package courses

import (
	"fmt"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/app/queries"
	"github.com/instructhub/backend/pkg/encryption"
	git "github.com/instructhub/backend/pkg/gitea"
	"github.com/instructhub/backend/pkg/utils"
)

// createCourseRequest is the type for the request body of creating a new course.
type createCourseRequest struct {
	Name        string `json:"name" binding:"required,max=50"`
	Description string `json:"description" binding:"required,max=200"`
}

// CreateNewCourse creates a new course with the given request data.
func CreateNewCourse(c *gin.Context) {
	// Retrieve user ID from context
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		utils.FullyResponse(c, 400, "Error getting userID", utils.ErrBadRequest, err.Error())
		return
	}

	// Bind and validate request body
	var request createCourseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.FullyResponse(c, 400, "Invalid request", utils.ErrBadRequest, err.Error())
		return
	}

	// Create course object
	course := createCourse(userID, request)

	// Create a new repository for the course
	repo, err := createCourseRepo(course)
	if err != nil {
		c.Error(err)
		utils.FullyResponse(c, 500, "Error creating new course repo", utils.ErrCreateNewCourse, nil)
		return
	}

	// Create course data file in the repository
	if err := createCourseFile(repo, userID); err != nil {
		c.Error(err)
		utils.FullyResponse(c, 500, "Error saving course file", utils.ErrCreateNewCourse, nil)
		return
	}

	// Save the course information to the database
	if err := saveCourseToDatabase(course); err != nil {
		c.Error(err)
		utils.FullyResponse(c, 500, "Error saving course", utils.ErrSaveData, nil)
		return
	}

	// Return success response
	utils.FullyResponse(c, 201, "Successfully created new course", nil, course)
}

// createCourse creates a new course object with the provided user ID and request.
func createCourse(userID uint64, request createCourseRequest) models.Course {
	return models.Course{
		ID:          encryption.GenerateID(),
		CreatorID:   userID,
		Name:        request.Name,
		Description: request.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// createCourseRepo creates a new Git repository for the course.
func createCourseRepo(course models.Course) (*gitea.Repository, error) {
	repoOptions := gitea.CreateRepoOption{
		Name:          utils.Uint64ToStr(uint64(course.ID)),
		DefaultBranch: "en",
		AutoInit:      true,
		Private:       true,
	}
	repo, _, err := git.GiteaClient.CreateOrgRepo(utils.GiteaORGName, repoOptions)
	return repo, err
}

// createCourseFile creates the course data file in the new repository.
func createCourseFile(repo *gitea.Repository, userID uint64) error {
	giteaFile := gitea.CreateFileOptions{
		FileOptions: gitea.FileOptions{
			Message: "init: Initialize the course",
			Committer: gitea.Identity{
				Name:  utils.Uint64ToStr(userID),
				Email: git.GenerateCommmitEmail(userID),
			},
			Author: gitea.Identity{
				Name: utils.Uint64ToStr(userID),
			},
		},
		Content: encryption.Base64Encode(""),
	}

	_, _, err := git.GiteaClient.CreateFile(repo.Owner.UserName, repo.Name, "course_data.json", giteaFile)
	return err
}

// saveCourseToDatabase saves the new course to the database.
func saveCourseToDatabase(course models.Course) error {
	result := queries.CreateNewCourse(course)
	if result.Error != nil || result.RowsAffected == 0 {
		return fmt.Errorf("failed to save course to database")
	}
	return nil
}
