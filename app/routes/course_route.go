package routes

import (
	"github.com/gin-gonic/gin"
	courses "github.com/instructhub/backend/app/controllers/course"
	"github.com/instructhub/backend/pkg/middleware"
)

func CourseRoute(r *gin.RouterGroup) {
	g := r.Group("/courses")

	// Get course public data
	g.GET("/:courseID", courses.GetCourse)
	g.GET("/:courseID/:stepID", courses.GetStepContent)
	g.GET("/landing/:courseID", courses.GetCourseLandingPageData)

	g.Use(middleware.IsAuthorized())
	// Course information
	g.POST("/new", courses.CreateNewCourse)
	g.POST("/landing/:courseID", courses.UpdateCourseLandingPage)

	// Revision
	g.POST("/revision/:courseID", courses.CreateNewRevision)
	g.POST("/revision/:courseID/:revisionID/approve", courses.ApproveRevision)

	// Image upload
	g.POST("/:courseID/image/upload", courses.UploadImage)
}
