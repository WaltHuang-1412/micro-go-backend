// @title           Micro Backend API
// @version         1.0
// @description     使用 JWT 的簡易用戶驗證 API
// @host            localhost:8088
// @BasePath        /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"github.com/Walter1412/micro-backend/docs" // ✅ 引用 swagger docs
	"github.com/Walter1412/micro-backend/handlers"
	"github.com/Walter1412/micro-backend/middlewares"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// ✅ 設定 Swagger 變數
	docs.SwaggerInfo.Host = os.Getenv("SWAGGER_HOST")
	docs.SwaggerInfo.Schemes = []string{os.Getenv("SWAGGER_SCHEME")}

	// 讀取 DB 設定
	databaseUser := os.Getenv("DB_USER")
	databasePassword := os.Getenv("DB_PASSWORD")
	databaseHost := os.Getenv("DB_HOST")
	databasePort := os.Getenv("DB_PORT")
	databaseName := os.Getenv("DB_NAME")

	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", databaseUser, databasePassword, databaseHost, databasePort, databaseName)

	database, error := sql.Open("mysql", dataSourceName)
	if error != nil {
		log.Fatal("❌ Failed to connect to DB:", error)
	}
	defer database.Close()

	// 自動重試 DB 連線
	maxRetries := 10
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if error := database.Ping(); error == nil {
			fmt.Println("✅ Connected to DB!")
			break
		} else {
			fmt.Printf("⏳ Waiting for DB... (attempt %d/%d)\n", attempt, maxRetries)
			time.Sleep(2 * time.Second)
		}
		if attempt == maxRetries {
			log.Fatal("❌ DB not reachable after retrying.")
		}
	}

	router := gin.Default()
	router.Use(middlewares.CORSMiddleware())

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	fmt.Println("✅ Swagger UI route registered at /swagger/*any")

	// API 路由
	apiRouter := router.Group("/api/v1")
	{
		apiRouter.POST("/register", handlers.Register(database))
		fmt.Println("✅ /api/v1/register route ready")

		apiRouter.POST("/login", handlers.Login(database))
		fmt.Println("✅ /api/v1/login route ready")

		protected := apiRouter.Group("")
		protected.Use(middlewares.JWTAuthMiddleware())
		{
			protected.GET("/profile", handlers.Profile())
			fmt.Println("✅ /api/v1/profile (protected) route ready")

			// ✅ plans 分組
			plans := protected.Group("/plans")
			{

				sections := plans.Group("/sections")
				{
					sections.GET("", handlers.GetSections(database))
					fmt.Println("✅ /api/v1/plans/sections GET route ready")
					sections.POST("", handlers.CreateSection(database))
					fmt.Println("✅ /api/v1/plans/sections POST route ready")
					sections.DELETE("/:id", handlers.DeleteSection(database))
					fmt.Println("✅ /api/v1/plans/sections/:id DELETE route ready")
					sections.PUT("/:id", handlers.UpdateSection(database))
					fmt.Println("✅ /api/v1/sections PUT route ready")

				}
				tasks := plans.Group("/tasks")
				{
					tasks.POST("", handlers.CreateTask(database))
					fmt.Println("✅ /api/v1/plans/tasks POST route ready")
					tasks.PUT("/:id", handlers.UpdateTask(database))
					fmt.Println("✅ /api/v1/plans/tasks/:id PUT route ready")
					tasks.DELETE("/:id", handlers.DeleteTask(database))
					fmt.Println("✅ /api/v1/plans/tasks/:id DELETE route ready")
				}
				plans.GET("/sections-with-tasks", handlers.GetSectionsWithTasks(database))
				fmt.Println("✅ /api/v1/plans/sections-with-tasks GET route ready")
				plans.PUT("/sections-with-tasks", handlers.UpdateSectionsWithTasks(database))
				fmt.Println("✅ /api/v1/plans/sections-with-tasks PUT route ready")
			}
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}

	fmt.Println("🚀 Server running at http://localhost:" + port)
	fmt.Println("🌐 Swagger UI available at http://localhost:" + port + "/swagger/index.html")
	router.Run(":" + port)
}
