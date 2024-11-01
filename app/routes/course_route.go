package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/instructhub/backend/app/controllers"
	"github.com/instructhub/backend/pkg/middleware"
)

func CourseRoute(r *gin.RouterGroup) {
	course := r.Group("/course")
	course.Use(middleware.IsAuthorized())

	course.POST("/new", controllers.CreateNewCourse)
	course.POST("/:courseID/upload", controllers.UploadImage)
}
