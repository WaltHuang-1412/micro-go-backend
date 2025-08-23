package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/Walter1412/micro-backend/handlers"
)

func RegisterProfileRoutes(router *gin.RouterGroup) {
	router.GET("/profile", handlers.Profile())
}