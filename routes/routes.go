package routes

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/Walter1412/micro-backend/middlewares"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(router *gin.Engine, database *sql.DB) {
	// CORS middleware
	router.Use(middlewares.CORSMiddleware())

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	apiRouter := router.Group("/api/v1")
	
	// Public routes (no auth required)
	RegisterAuthRoutes(apiRouter, database)

	// Protected routes (JWT auth required)
	protected := apiRouter.Group("")
	protected.Use(middlewares.JWTAuthMiddleware())
	{
		RegisterProfileRoutes(protected)
		RegisterPlanRoutes(protected, database)
	}
}