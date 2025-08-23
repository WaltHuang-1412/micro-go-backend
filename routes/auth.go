package routes

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/Walter1412/micro-backend/handlers"
)

func RegisterAuthRoutes(router *gin.RouterGroup, database *sql.DB) {
	router.POST("/register", handlers.Register(database))
	router.POST("/login", handlers.Login(database))
}