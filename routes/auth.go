package routes

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/Walter1412/micro-backend/handlers"
	"github.com/Walter1412/micro-backend/services"
)

func RegisterAuthRoutes(router *gin.RouterGroup, database *sql.DB, emailService *services.EmailService) {
	router.POST("/register", handlers.Register(database))
	router.POST("/login", handlers.Login(database))
	router.POST("/forgot-password", handlers.ForgotPassword(database, emailService))
	router.POST("/reset-password", handlers.ResetPassword(database))
	
	// 開發測試端點
	router.GET("/dev/latest-token", handlers.GetLatestToken(database))
}