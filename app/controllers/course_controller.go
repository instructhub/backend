package controllers

import (
	"bytes"
	"context"
	"fmt"
	"io"
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
		CourseTitle            string `json:"course_title" binding:"required"`
		CourseShortDescription string `json:"course_short_description" binding:"required"`
	}

	var request CreateCourseRequest

	ContextUserID, exist := c.Get("userID")
	if !exist {
		utils.SimpleResponse(c, 400, "Error get userID", utils.ErrBadRequest, nil)
		return
	}

	userID := ContextUserID.(uint64)

	// Validate request body
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SimpleResponse(c, 400, "Invalid request", utils.ErrBadRequest, err.Error())
		return
	}

	course := models.Course{
		CourseCreator:          userID,
		CourseTitle:            request.CourseTitle,
		CourseShortDescription: request.CourseShortDescription,
		CourseID:               encryption.GenerateID(),
		CreateAt:               time.Now(),
		UpdatedAt:              time.Now(),
	}

	repoOptions := gitea.CreateRepoOption{
		Name:          strconv.FormatUint(uint64(course.CourseID), 10),
		Description:   "Title:" + request.CourseTitle + " Description:" + request.CourseShortDescription,
		DefaultBranch: "en",
		AutoInit:      true,
		Private:       true,
	}

	_, _, err := gt.GiteaClient.CreateOrgRepo(utils.GiteaORGName, repoOptions)
	if err != nil {
		c.Error(err)
		utils.SimpleResponse(c, 500, "Internal server error while create new course", utils.ErrCreateNewCourse, nil)
		return
	}

	err = queries.CraeteNewCourse(course)
	if err != nil {
		c.Error(err)
		utils.SimpleResponse(c, 500, "Internal server error while save course to database", utils.ErrSaveData, nil)
		return
	}

	utils.SimpleResponse(c, 201, "Successful create new course", nil, nil)
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

	_, err = queries.GetCourseInformation(courseID)
	if err != nil {
		utils.SimpleResponse(c, 400, "This course doesn't exist", utils.ErrCourseNotExist, nil)
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

	err = queries.CraeteCourseImage(models.CourseImage{
		ImageLink: filePath,
		CourseID:  courseID,
		Craetor:   userID,
		CreatedAt: time.Now(),
	})

	if err != nil {
		c.Error(err)
		utils.SimpleResponse(c, 500, "Internal server error while upload to MongoDB", utils.ErrSaveData, nil)
		return
	}

	// Construct the file URL
	fileURL := fmt.Sprintf("%s/%s", store.StaticBucketUrl, filePath)
	utils.SimpleResponse(c, 201, "File uploaded successfully", nil, fileURL)
}
