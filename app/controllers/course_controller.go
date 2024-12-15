package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/app/queries"
	"github.com/instructhub/backend/pkg/encryption"
	git "github.com/instructhub/backend/pkg/gitea"
	store "github.com/instructhub/backend/pkg/s3"
	"github.com/instructhub/backend/pkg/utils"
	"gorm.io/gorm"
)

func CreateNewCourse(c *gin.Context) {
	type CreateCourseRequest struct {
		Name        string `json:"name" binding:"required,max=50"`
		Description string `json:"description" binding:"required,max=200"`
	}

	var request CreateCourseRequest

	ContextUserID, exist := c.Get("userID")
	if !exist {
		utils.SimpleResponse(c, 400, "Error get userID", utils.ErrBadRequest, nil)
		return
	}

	userID := ContextUserID.(uint64)

	userIDString := utils.Uint64ToStr(userID)

	// Validate request body
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SimpleResponse(c, 400, "Invalid request", utils.ErrBadRequest, err.Error())
		return
	}

	course := models.Course{
		ID:          encryption.GenerateID(),
		Creator:     userID,
		Name:        request.Name,
		Description: request.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	repoOptions := gitea.CreateRepoOption{
		Name:          utils.Uint64ToStr(uint64(course.ID)),
		DefaultBranch: "en",
		AutoInit:      true,
		Private:       true,
	}

	repo, _, err := git.GiteaClient.CreateOrgRepo(utils.GiteaORGName, repoOptions)
	if err != nil {
		c.Error(err)
		utils.SimpleResponse(c, 500, "Internal server error while create new course", utils.ErrCreateNewCourse, nil)
		return
	}

	giteaFile := gitea.CreateFileOptions{
		FileOptions: gitea.FileOptions{
			Message: "init: Initialize the course",
			Committer: gitea.Identity{
				Name: userIDString,
			},
			Author: gitea.Identity{
				Name: userIDString,
			},
		},
		Content: encryption.Base64Encode(""),
	}

	_, _, err = git.GiteaClient.CreateFile(repo.Owner.UserName, repo.Name, "course_data.json", giteaFile)
	if err != nil {
		c.Error(err)
		utils.SimpleResponse(c, 500, "Error save course file", utils.ErrCreateNewCourse, nil)
		return
	}

	result := queries.CreateNewCourse(course)
	if result.Error != nil || result.RowsAffected == 0 {
		c.Error(result.Error)
		utils.SimpleResponse(c, 500, "Internal server error while save course to database", utils.ErrSaveData, nil)
		return
	}

	utils.SimpleResponse(c, 201, "Successful create new course", nil, course)
}

// TODO: change all content to base64
type CourseItemRequest struct {
	ID       *uint64           `json:"id" binding:"omitempty,number"`
	StageID  *uint64           `json:"stage_id,omitempty" binding:"omitempty,number"`
	Position int               `json:"position" binding:"required"`
	Type     models.CourseType `json:"type" binding:"number"`
	Name     string            `json:"name" binding:"max=50"`
	Updated  bool              `json:"updated"`
	Content  *string           `json:"content,omitempty" binding:"omitempty,max=100000"`
}

type CourseStageRequest struct {
	ID       *uint64 `json:"id" binding:"omitempty,number"`
	Position int     `json:"position" binding:"required,number"`
	Name     string  `json:"name" binding:"required,max=30"`

	CourseItems []CourseItemRequest `json:"course_items" binding:"max=20,dive"`
}

type UpdateRequestCourse struct {
	Stages      []CourseStageRequest `json:"stages" binding:"min=1,max=10,dive"`
	Description string               `json:"description" binding:"required,max=100"`
}

// UpdateCourseContent handles course content updates
func UpdateCourseContent(c *gin.Context) {
	var request UpdateRequestCourse

	// Validate request body
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SimpleResponse(c, 400, "Invalid request", utils.ErrBadRequest, err.Error())
		return
	}

	// Parse course ID and get user ID from context
	courseID, err := strconv.ParseUint(c.Param("courseID"), 10, 64)
	ContextUserID, exist := c.Get("userID")
	if !exist || err != nil {
		utils.SimpleResponse(c, 400, "Error getting userID", utils.ErrBadRequest, nil)
		return
	}
	userID := ContextUserID.(uint64)
	userIDString := utils.Uint64ToStr(userID)
	// Sort stages and items by position
	sort.Slice(request.Stages, func(i, j int) bool {
		return request.Stages[i].Position < request.Stages[j].Position
	})
	for i := range request.Stages {
		sort.Slice(request.Stages[i].CourseItems, func(x, y int) bool {
			return request.Stages[i].CourseItems[x].Position < request.Stages[i].CourseItems[y].Position
		})
	}

	// Fetch old course data
	oldCourseData, result := queries.GetCourseWithDetails(courseID)
	if result.Error == gorm.ErrRecordNotFound {
		utils.SimpleResponse(c, 404, "Course not found", utils.ErrCourseNotExist, nil)
		return
	} else if result.Error != nil {
		c.Error(result.Error)
		utils.SimpleResponse(c, 500, "Internal server error while fetching course", utils.ErrGetData, nil)
		return
	}

	if userID != oldCourseData.Creator {
		utils.SimpleResponse(c, 403, "You are not the course creator, you cannot update it", utils.ErrUnauthorized, nil)
		return
	}

	// Validate stage and item positions
	for i, stage := range request.Stages {
		if stage.Position != i+1 {
			utils.SimpleResponse(c, 400, fmt.Sprintf("Invalid stage position at index %d, expected %d but got %d", i, i+1, stage.Position), utils.ErrBadRequest, nil)
			return
		}

		for j, item := range stage.CourseItems {
			if item.Position != j+1 {
				utils.SimpleResponse(c, 400, fmt.Sprintf("Invalid item position at stage %d, item index %d, expected %d but got %d", stage.Position, j, j+1, item.Position), utils.ErrBadRequest, nil)
				return
			}
		}
	}

	updateFiles := []git.File{}

	// Map all new course items
	newCourseItems := map[uint64]bool{}
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
			if !newCourseItems[item.ID] {
				fmt.Println("test")
				updateFiles = append(updateFiles, git.File{
					Path:      utils.Uint64ToStr(item.ID),
					Operation: git.OperationDelete,
				})
			}
		}
	}

	courseData := request

	// Assign IDs for new stages and items, and collect updates
	for i, stage := range request.Stages {
		if stage.ID == nil {
			stageID := encryption.GenerateID()
			stage.ID = &stageID
			courseData.Stages[i].ID = &stageID
		}

		for j, item := range stage.CourseItems {
			if item.ID == nil {
				itemID := encryption.GenerateID()
				item.ID = &itemID
				courseData.Stages[i].CourseItems[j].ID = &itemID
				courseData.Stages[i].CourseItems[j].Content = nil
				updateFiles = append(updateFiles, git.File{
					Path:      utils.Uint64ToStr(*item.ID),
					Content:   encryption.Base64Encode(*item.Content),
					Operation: git.OperationCreate,
				})
			} else if item.Updated {
				updateFiles = append(updateFiles, git.File{
					Path:      utils.Uint64ToStr(*item.ID),
					Content:   encryption.Base64Encode(*item.Content),
					Operation: git.OperationUpdate,
				})
			}
		}
	}
	courseDataJson, err := json.Marshal(courseData)
	if err != nil {
		c.Error(err)
		utils.SimpleResponse(c, 500, "Internal server error while parse json", utils.ErrParseData, nil)
		return
	}

	// Add the JSON data as a new file or update existing file
	updateFiles = append(updateFiles, git.File{
		Path:      "course_data.json",
		Content:   encryption.Base64Encode(string(courseDataJson)),
		Operation: git.OperationUpdate,
	})

	identity := gitea.Identity{
		Name:  userIDString,
		Email: fmt.Sprintf("%s@instructhub.org", userIDString),
	}

	branchID := encryption.GenerateID()

	modifyMultipleFilesData := git.ModifyRequest{
		Author:    identity,
		Committer: identity,
		Files:     updateFiles,
		Message:   request.Description,
		NewBranch: utils.Uint64ToStr(branchID),
	}

	err = git.ModifyMultipleFiles(utils.GiteaORGName, utils.Uint64ToStr(courseID), modifyMultipleFilesData)
	if err != nil {
		c.Error(err)
		utils.SimpleResponse(c, 500, "Internal server error save data to git", utils.ErrSaveCourseFile, nil)
		return
	}

	prOptions := gitea.CreatePullRequestOption{
		Head:  utils.Uint64ToStr(branchID),
		Base:  "en",
		Title: request.Description,
	}
	pullRequest, _, err := git.GiteaClient.CreatePullRequest(utils.GiteaORGName, utils.Uint64ToStr(courseID), prOptions)
	if err != nil {
		c.Error(err)
		utils.SimpleResponse(c, 500, "Internal server error save data to git", utils.ErrSaveCourseFile, nil)
		return
	}

	courseHistory := models.CourseRevision{
		ID:            encryption.GenerateID(),
		CourseID:      courseID,
		BranchID:      branchID,
		PullRequestID: int(pullRequest.ID),
		Description:   request.Description,
		EditorID:      &userID,
		Status:        models.HistoryOpen,
		UpdatedAt:     time.Now(),
		CreatedAt:     time.Now(),
	}

	result = queries.CreateNewCourseHistory(courseHistory)
	if result.Error != nil || result.RowsAffected == 0 {
		c.Error(err)
		utils.SimpleResponse(c, 500, "Internal server error save history data", utils.ErrSaveData, nil)
		return
	}

	utils.SimpleResponse(c, 201, "Successful create a new request", nil, nil)
}

func ApproveRevision(c *gin.Context) {
	// Parse course ID and get user ID from context
	courseID, err := utils.StrToUint64(c.Param("courseID"))
	ContextUserID, exist := c.Get("userID")
	if !exist || err != nil {
		utils.SimpleResponse(c, 400, "Error getting userID", utils.ErrBadRequest, nil)
		return
	}

	revisionID, err := utils.StrToUint64(c.Param("revisionID"))
	if !exist || err != nil {
		utils.SimpleResponse(c, 400, "Error getting revisionID", utils.ErrBadRequest, nil)
		return
	}

	userID := ContextUserID.(uint64)
	utils.Uint64ToStr(userID)
	// Fetch old course data
	revision, result := queries.GetCourseRevision(courseID, revisionID)
	if result.Error == gorm.ErrRecordNotFound {
		utils.SimpleResponse(c, 404, "Revision not found", utils.ErrCourseNotExist, nil)
		return
	} else if result.Error != nil {
		c.Error(result.Error)
		utils.SimpleResponse(c, 500, "Internal server error while fetching revision", utils.ErrGetData, nil)
		return
	}

	revisionChangeFile, _, err := git.GiteaClient.GetContents(utils.GiteaORGName, utils.Uint64ToStr(revision.CourseID), utils.Uint64ToStr(revision.BranchID), "/course_data.json")
	if err != nil {
		fmt.Println(err.Error())
	}

	revisionChangeDataString, err := encryption.Base64Decode(*revisionChangeFile.Content)
	if err != nil {
		c.Error(err)
		utils.SimpleResponse(c, 500, "Internal server error while fetching course data", utils.ErrGetData, nil)
		return
	}

	var updateRequest UpdateRequestCourse

	err = json.Unmarshal([]byte(revisionChangeDataString), &updateRequest)
	if err != nil {
		c.Error(err)
		utils.SimpleResponse(c, 500, "Internal server error while unmarshal data ", utils.ErrUnmarshal, nil)
		return
	}

	utils.SimpleResponse(c, 200, "Successful approve this revision", nil, updateRequest)
}

func UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		utils.SimpleResponse(c, 400, "Image required", utils.ErrImageRequired, nil)
		return
	}

	courseID, err := strconv.ParseUint(c.Param("courseID"), 10, 64)
	if err != nil {
		utils.SimpleResponse(c, 400, "Missing course ID", utils.ErrMissingCourseID, nil)
		return
	}

	userIDUntype, exists := c.Get("userID")
	if !exists {
		utils.SimpleResponse(c, 403, "UserID not found in context", utils.ErrUserIDNotFound, nil)
		return
	}

	userID := userIDUntype.(uint64)

	_, result := queries.GetCourseInformation(courseID)
	if result.Error == gorm.ErrRecordNotFound {
		utils.SimpleResponse(c, 400, "This course doesn't exist", utils.ErrCourseNotExist, nil)
		return
	} else if result.Error != nil {
		c.Error(result.Error)
		utils.SimpleResponse(c, 500, "Internal server error while get course data", utils.ErrGetData, nil)
		return
	}

	// Set max file size to 5 MB
	const maxFileSize = 10 * 1024 * 1024

	if file.Size > maxFileSize {
		utils.SimpleResponse(c, 400, "Image too large ( > 10MB)", utils.ErrImageTooLarge, nil)
		return
	}

	// Open the uploaded file
	srcFile, err := file.Open()
	if err != nil {
		utils.SimpleResponse(c, 400, "Error opening image", utils.ErrOpeningImage, nil)
		return
	}
	defer srcFile.Close()

	// Read the first few bytes of the file for magic number detection
	magic := make([]byte, 8) // Adjust size as needed
	_, err = srcFile.Read(magic)
	if err != nil && err != io.EOF {
		c.Error(err)
		utils.SimpleResponse(c, 500, "Internal server error while reading image", utils.ErrReadingImage, nil)
		return
	}

	// Check if it's a valid image type
	isImage, contentType, err := utils.IsValidImageType(magic)
	if err != nil || !isImage {
		utils.SimpleResponse(c, 400, "Uploaded file is not a valid image", utils.ErrInvalidImage, nil)
		return
	}

	// Reset the file pointer to the beginning to ensure the full file can be read
	_, err = srcFile.Seek(0, io.SeekStart)
	if err != nil {
		utils.SimpleResponse(c, 500, "Error resetting file pointer", utils.ErrResetFilePointer, nil)
		return
	}

	// Read the entire file into a bytes.Buffer
	var src bytes.Buffer
	if _, err := src.ReadFrom(srcFile); err != nil {
		c.Error(err)
		utils.SimpleResponse(c, 500, "Internal server error while reading image", utils.ErrReadingImage, nil)
		return
	}

	// Prepare the filename for S3
	courseIDString := strconv.Itoa(int(courseID))
	imageID := strconv.Itoa(int(encryption.GenerateID()))
	cleanedFilename := strings.ReplaceAll(file.Filename, " ", "")
	filePath := fmt.Sprintf("%s/%s-%s", courseIDString, imageID, cleanedFilename)

	// Upload the image to S3 with explicit content type
	_, err = store.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      &store.StaticBucket,
		Key:         &filePath,
		Body:        bytes.NewReader(src.Bytes()),
		ContentType: &contentType, // Set the content type here
	})
	if err != nil {
		c.Error(err)
		utils.SimpleResponse(c, 500, "Internal server error while upload the file", utils.ErrS3UploadFailed, nil)
		return
	}

	result = queries.CreateCourseImage(models.CourseImage{
		ImageLink: filePath,
		ID:        courseID,
		Creator:   userID,
		CreatedAt: time.Now(),
	})
	if result.Error != nil || result.RowsAffected == 0 {
		c.Error(err)
		utils.SimpleResponse(c, 500, "Internal server error while upload to database", utils.ErrSaveData, nil)
		return
	}

	// Construct the file URL
	fileURL := fmt.Sprintf("%s/%s", store.StaticBucketUrl, filePath)
	utils.SimpleResponse(c, 201, "File uploaded successfully", nil, fileURL)
}
