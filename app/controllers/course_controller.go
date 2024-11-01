package controllers

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
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
	"github.com/nfnt/resize"
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
        Name: strconv.FormatUint(uint64(course.CourseID), 10),
		Description: "Title:" + request.CourseTitle + " Description:" + request.CourseShortDescription,
		DefaultBranch: "en",
		AutoInit: true,
		Private: true,
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
				Committer:  gitea.Identity{
					Name: userIDString,
				},
				Author:  gitea.Identity{
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

	// Check if the uploaded file is an image
	contentType := file.Header.Get("Content-Type")
	if !utils.IsValidImageType(contentType) {
		utils.SimpleResponse(c, 400, "Uploaded file is not a valid image", nil)
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
	const maxFileSize = 10 * 1024 * 1024 // 10 MB
	var src bytes.Buffer

	if file.Size > maxFileSize {
		// File is larger than 10 MB, compress it
		srcFile, err := file.Open()
		if err != nil {
			utils.SimpleResponse(c, 400, "Error opening image", nil)
			return
		}
		defer srcFile.Close()

		// Decode the image and resize it
		img, _, err := image.Decode(srcFile)
		if err != nil {
			utils.SimpleResponse(c, 500, "Error decoding image", nil)
			return
		}
		// Resize image to maximum width 1024px, keeping aspect ratio
		resizedImg := resize.Resize(1024, 0, img, resize.Lanczos3)

		// Encode the resized image into bytes.Buffer
		if err := jpeg.Encode(&src, resizedImg, &jpeg.Options{Quality: 80}); err != nil {
			utils.SimpleResponse(c, 500, "Error encoding resized image", nil)
			return
		}

		// Check if compressed image is still over 10 MB
		if src.Len() > maxFileSize {
			utils.SimpleResponse(c, 400, "Image too large after compression", nil)
			return
		}
	} else {
		// File is within 10 MB limit, upload directly
		srcFile, err := file.Open()
		if err != nil {
			utils.SimpleResponse(c, 400, "Error opening image", nil)
			return
		}
		defer srcFile.Close()

		// Read original file into bytes.Buffer
		if _, err := src.ReadFrom(srcFile); err != nil {
			utils.SimpleResponse(c, 500, "Error reading image", nil)
			return
		}
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

