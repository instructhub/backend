package courses

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/app/queries"
	"github.com/instructhub/backend/pkg/encryption"
	store "github.com/instructhub/backend/pkg/s3"
	"github.com/instructhub/backend/pkg/utils"
	"gorm.io/gorm"
)

// UploadImage handles the image upload to S3 and saving the image metadata in the database
func UploadImage(c *gin.Context) {
	// Parse image file and course ID from request
	file, err := c.FormFile("image")
	if err != nil {
		utils.ServerErrorResponse(c, 400, "Image required", utils.ErrImageRequired, err)
		return
	}

	courseID, err := utils.StrToUint64(c.Param("courseID"))
	if err != nil {
		utils.ServerErrorResponse(c, 400, "Invalid course ID", utils.ErrMissingCourseID, err)
		return
	}

	userID, err := parseUserID(c)
	if err != nil {
		utils.ServerErrorResponse(c, 403, "UserID not found in context", utils.ErrUserIDNotFound, err)
		return
	}

	// Validate course existence
	if err := validateCourseExistence(courseID); err != nil {
		utils.ServerErrorResponse(c, 400, "This course doesn't exist", utils.ErrCourseNotExist, err)
		return
	}

	// Validate file size
	const maxFileSize = 10 * 1024 * 1024 // 10 MB
	if file.Size > maxFileSize {
		utils.ServerErrorResponse(c, 400, "Image too large ( > 10MB)", utils.ErrImageTooLarge, nil)
		return
	}

	// Open the uploaded file
	srcFile, err := file.Open()
	if err != nil {
		utils.ServerErrorResponse(c, 400, "Error opening image", utils.ErrOpeningImage, err)
		return
	}
	defer srcFile.Close()

	// Validate image type
	contentType, err := validateImageType(srcFile)
	if err != nil {
		utils.ServerErrorResponse(c, 400, "Uploaded file is not a valid image", utils.ErrInvalidImage, err)
		return
	}

	// Reset the file pointer to the beginning to ensure the full file can be read
	_, err = srcFile.Seek(0, io.SeekStart)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error resetting file pointer", utils.ErrResetFilePointer, err)
		return
	}

	// Read the entire file content into a buffer
	src := bytes.Buffer{}
	if _, err := src.ReadFrom(srcFile); err != nil {
		utils.ServerErrorResponse(c, 500, "Error reading image", utils.ErrReadingImage, err)
		return
	}

	// Generate S3 file path and upload the image
	filePath, imageID, err := generateFilePath(courseID, file.Filename)
	if err != nil {
		utils.ServerErrorResponse(c, 500, "Error generating file path", utils.ErrGeneratingFilePath, err)
		return
	}

	if err := uploadToS3(filePath, contentType, src.Bytes()); err != nil {
		utils.ServerErrorResponse(c, 500, "Error uploading file to S3", utils.ErrS3UploadFailed, err)
		return
	}

	// Save image metadata in the database
	if err := saveImageMetadata(imageID, userID, filePath); err != nil {
		utils.ServerErrorResponse(c, 500, "Error saving image metadata", utils.ErrSaveData, err)
		return
	}

	// Construct the file URL and send the response
	fileURL := fmt.Sprintf("%s/%s", store.StaticBucketUrl, filePath)
	utils.FullyResponse(c, 201, "File uploaded successfully", nil, fileURL)
}

// parseUserID retrieves the user ID from the context
func parseUserID(c *gin.Context) (uint64, error) {
	userIDUntype, exists := c.Get("userID")
	if !exists {
		return 0, fmt.Errorf("userID not found in context")
	}
	return userIDUntype.(uint64), nil
}

// validateCourseExistence checks if the course exists in the database
func validateCourseExistence(courseID uint64) error {
	_, result := queries.GetCourseInformation(courseID)
	if result.Error == gorm.ErrRecordNotFound {
		return fmt.Errorf("course not found")
	} else if result.Error != nil {
		return result.Error
	}
	return nil
}

// validateImageType checks if the uploaded file is a valid image
func validateImageType(srcFile io.Reader) (contentType string, err error) {
	magic := make([]byte, 8)
	_, err = srcFile.Read(magic)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("error reading file magic number")
	}

	isImage, contentType, err := utils.IsValidImageType(magic)
	if err != nil || !isImage {
		return "", fmt.Errorf("invalid image type")
	}
	return contentType, nil
}

// generateFilePath generates a unique file path for the image in S3
func generateFilePath(courseID uint64, filename string) (string, uint64, error) {
	courseIDString := utils.Uint64ToStr(courseID)
	imageID := encryption.GenerateID()
	cleanedFilename := strings.ReplaceAll(filename, " ", "")
	return fmt.Sprintf("%s/%s-%s", courseIDString, utils.Uint64ToStr(imageID), cleanedFilename), imageID, nil
}

// uploadToS3 uploads the image to S3 with the generated file path
func uploadToS3(filePath string, contentType string, fileContent []byte) error {
	_, err := store.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      &store.StaticBucket,
		Key:         &filePath,
		Body:        bytes.NewReader(fileContent),
		ContentType: &contentType, // Assuming contentType is set earlier
	})
	return err
}

// saveImageMetadata saves the image metadata (file path, course ID, user ID) in the database
func saveImageMetadata(imageID, userID uint64, filePath string) error {
	result := queries.CreateCourseImage(models.CourseImage{
		ImageLink: filePath,
		ID:        imageID,
		CreatorID: userID,
		CreatedAt: time.Now(),
	})
	if result.Error != nil || result.RowsAffected == 0 {
		return fmt.Errorf("failed to save image metadata")
	}
	return nil
}
