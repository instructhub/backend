package routes

import (
	"github.com/gin-gonic/gin"
	courses "github.com/instructhub/backend/app/controllers/course"
	"github.com/instructhub/backend/pkg/middleware"
)

func CourseRoute(r *gin.RouterGroup) {
	g := r.Group("/courses")

	g.GET("/:courseID", courses.GetCourse)
	g.GET("/:courseID/:stepID", courses.GetStepContent)

	g.Use(middleware.IsAuthorized())
	g.POST("/new", courses.CreateNewCourse)
	// TODO: update metadata
	// course.PUT("/metadata", controllers.CreateNewCourse)
	g.POST("/revision/:courseID", courses.CreateNewRevision)
	g.POST("/revision/:courseID/:revisionID/approve", courses.ApproveRevision)

	g.POST("/:courseID/image/upload", courses.UploadImage)
}
