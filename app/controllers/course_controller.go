package controllers

import (
	"strconv"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/models"
	"github.com/instructhub/backend/app/queries"
	"github.com/instructhub/backend/pkg/encryption"
	gt "github.com/instructhub/backend/pkg/gitea"
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

	utils.SimpleResponse(c, 200, "Successful create new course", nil)
}

// func UploadImage(c *gin.Context) {
// 	file, err := c.FormFile("image")
// 	if err != nil {
// 		utils.SimpleResponse(c, 400, "Image required", nil)
// 		return
// 	}

// 	// Open the uploaded file
// 	src, err := file.Open()
// 	if err != nil {
// 		utils.SimpleResponse(c, 400, "Error open image", nil)
// 		return
// 	}
// 	defer src.Close()

// 	// Upload the file to S3
// 	_, err = s3.Client.PutObject(context.TODO(), &s3.PutObjectInput{
// 		Bucket: ,
// 		Key:    &file.Filename,
// 		Body:   src,
// 	})
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to upload the file"})
// 		return
// 	}

// 	// Construct the file URL
// 	fileURL := fmt.Sprintf("%s/%s/%s", os.Getenv("S3_ENDPOINT"), ctrl.Bucket, file.Filename)

// 	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "url": fileURL})
// }
