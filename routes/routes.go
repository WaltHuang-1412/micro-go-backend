package routes

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/Walter1412/micro-backend/config"
	"github.com/Walter1412/micro-backend/middlewares"
	"github.com/Walter1412/micro-backend/services"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(router *gin.Engine, database *sql.DB, cfg *config.Config) {
	// Initialize services
	emailService := services.NewEmailService(cfg.Email)

	// CORS middleware
	router.Use(middlewares.CORSMiddleware())
	
	// Rate limiting middleware
	router.Use(middlewares.RateLimitMiddleware())

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	apiRouter := router.Group("/api/v1")
	
	// Public routes (no auth required)
	RegisterAuthRoutes(apiRouter, database, emailService)

	// Protected routes (JWT auth required)
	protected := apiRouter.Group("")
	protected.Use(middlewares.JWTAuthMiddleware())
	{
		RegisterProfileRoutes(protected)
		RegisterPlanRoutes(protected, database)
	}
}