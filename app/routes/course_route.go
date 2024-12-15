package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/controllers"
	"github.com/instructhub/backend/pkg/middleware"
)

func CourseRoute(r *gin.RouterGroup) {
	course := r.Group("/courses")
	course.Use(middleware.IsAuthorized())

	course.POST("/new", controllers.CreateNewCourse)
	// TODO: update metadata
	// course.PUT("/metadata", controllers.CreateNewCourse)
	course.POST("/revision/:courseID", controllers.UpdateCourseContent)
	course.POST("/revision/:courseID/:revisionID/approve", controllers.ApproveRevision)

	course.POST("/:courseID/image/upload", controllers.UploadImage)
}
