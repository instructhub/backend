package controllers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/app/queries"
	"github.com/instructhub/backend/pkg/encryption"
	gt "github.com/instructhub/backend/pkg/gitea"
	"github.com/instructhub/backend/pkg/s3"
	"github.com/instructhub/backend/pkg/utils"
)

func CreateNewCourse(c *gin.Context) {
	type CreateCourseRequest struct {
		CourseTitle            string              `json:"course_title" binding:"required"`
		CourseShortDescription string              `json:"course_short_description" binding:"required"`
		Files                  []models.CourseFile `json:"files" binding:"required,dive"`
	}

	var course models.Course
	var request CreateCourseRequest

	ContextUserID, exist := c.Get("userID")
	if !exist {
		utils.SimpleResponse(c, 500, "Error get userID", nil)
		return
	}

	userID := ContextUserID.(uint64)

	userIDString := strconv.FormatUint(uint64(userID), 10)

	// Validate request body
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SimpleResponse(c, 400, "Invalid request", err.Error())
		return
	}

	course.CourseCreator = userID
	course.CourseTitle = request.CourseTitle
	course.CourseShortDescription = request.CourseShortDescription
	course.CourseID = encryption.GenerateID()
	course.CreateAt = time.Now()
	course.UpdatedAt = time.Now()

	// Check if the stage duplicate
	stagesSeen := make(map[string]bool)
	for _, file := range request.Files {
		if _, exists := stagesSeen[file.Stage]; exists {
			utils.SimpleResponse(c, 400, "Duplicate course stage: "+file.Stage, nil)
			return
		}
		stagesSeen[file.Stage] = true
	}

	repoOptions := gitea.CreateRepoOption{
		Name:          strconv.FormatUint(uint64(course.CourseID), 10),
		Description:   "Title:" + request.CourseTitle + " Description:" + request.CourseShortDescription,
		DefaultBranch: "en",
		AutoInit:      true,
		Private:       true,
	}

	repo, _, err := gt.GiteaClient.CreateOrgRepo(utils.GiteaORGName, repoOptions)
	if err != nil {
		utils.SimpleResponse(c, 500, "Error create new course", err.Error())
		return
	}

	for _, file := range request.Files {
		giteaFile := gitea.CreateFileOptions{
			FileOptions: gitea.FileOptions{
				Message: file.Message,
				Committer: gitea.Identity{
					Name: userIDString,
				},
				Author: gitea.Identity{
					Name: userIDString,
				},
			},
			Content: file.Content,
		}

		_, _, err := gt.GiteaClient.CreateFile(repo.Owner.UserName, repo.Name, file.Stage, giteaFile)
		if err != nil {
			utils.SimpleResponse(c, 500, "Error save course file", err.Error())
			return
		}
	}

	err = queries.CraeteNewCourse(course)
	if err != nil {
		utils.SimpleResponse(c, 500, "Error save course to database", err.Error())
		return
	}

	utils.SimpleResponse(c, 201, "Successful create new course", nil)
}

func UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		utils.SimpleResponse(c, 400, "Image required", nil)
		return
	}

	courseID, err := strconv.ParseUint(c.Param("courseID"), 10, 64)
	if err != nil {
		utils.SimpleResponse(c, 400, "Missing course ID", nil)
		return
	}

	userIDUntype, exists := c.Get("userID")
	if !exists {
		utils.SimpleResponse(c, 403, "UserID not found in context", nil)
		return
	}

	userID := userIDUntype.(uint64)

	_, err = queries.GetCourseInformation(courseID)
	if err != nil {
		utils.SimpleResponse(c, 400, "This course doesn't exist", nil)
		return
	}

	// Set max file size to 10 MB
	const maxFileSize = 10 * 1024 * 1024
	var src bytes.Buffer

	if file.Size > maxFileSize {
		utils.SimpleResponse(c, 400, "Image too large(>10MB)", nil)
		return
	}

	// Open the uploaded file
	srcFile, err := file.Open()
	if err != nil {
		utils.SimpleResponse(c, 500, "Error opening image", nil)
		return
	}
	defer srcFile.Close()

	// Read the first few bytes of the file for magic number detection
	magic := make([]byte, 8) // Adjust the size based on your needs
	_, err = srcFile.Read(magic)
	if err != nil && err != io.EOF {
		utils.SimpleResponse(c, 500, "Error reading image", nil)
		return
	}

	isImage, contentType, err := utils.IsValidImageType(magic)
	if err != nil {
		utils.SimpleResponse(c, 400, err.Error(), nil)
		return
	}

	if !isImage {
		utils.SimpleResponse(c, 400, "This is not a image type", nil)
	}

	// Read original file into bytes.Buffer
	if _, err := src.ReadFrom(srcFile); err != nil {
		utils.SimpleResponse(c, 500, "Error reading image", nil)
		return
	}

	// Prepare the filename for S3
	courseIDString := strconv.Itoa(int(courseID))
	imageID := strconv.Itoa(int(encryption.GenerateID()))
	cleanedFilename := strings.ReplaceAll(file.Filename, " ", "")
	filePath := fmt.Sprintf("%s/%s-%s", courseIDString, imageID, cleanedFilename)

	// Upload the image to S3 with explicit content type
	_, err = store.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      &store.CourseImageBuckerName,
		Key:         &filePath,
		Body:        bytes.NewReader(src.Bytes()),
		ContentType: &contentType, // Set the content type here
	})
	if err != nil {
		utils.SimpleResponse(c, 500, "Unable to upload the file", nil)
		return
	}

	err = queries.CraeteCourseImage(models.CourseImage{
		ImageLink: filePath,
		CourseID:  courseID,
		Craetor:   userID,
		CreatedAt: time.Now(),
	})
	if err != nil {
		utils.SimpleResponse(c, 500, "Failed to upload to mongodb", nil)
		return
	}

	// Construct the file URL
	fileURL := fmt.Sprintf("%s/%s/%s", os.Getenv("S3_ENDPOINT"), store.CourseImageBuckerName, filePath)
	utils.SimpleResponse(c, 201, "File uploaded successfully", fileURL)
}
