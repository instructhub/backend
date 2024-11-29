package routes

import (
	"github.com/gin-gonic/gin"
)

func PublicRouter(r *gin.RouterGroup) {
	 r.Group("/public")
}
