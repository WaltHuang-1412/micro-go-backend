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
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("❌ Failed to connect to DB:", err)
	}
	defer db.Close()

	// 自動重試 DB 連線
	maxRetries := 10
	for i := 1; i <= maxRetries; i++ {
		if err := db.Ping(); err == nil {
			fmt.Println("✅ Connected to DB!")
			break
		} else {
			fmt.Printf("⏳ Waiting for DB... (attempt %d/%d)\n", i, maxRetries)
			time.Sleep(2 * time.Second)
		}
		if i == maxRetries {
			log.Fatal("❌ DB not reachable after retrying.")
		}
	}

	r := gin.Default()
	r.Use(middlewares.CORSMiddleware())

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	fmt.Println("✅ Swagger UI route registered at /swagger/*any")

	// API 路由
	api := r.Group("/api/v1")
	{
		api.POST("/register", handlers.Register(db))
		fmt.Println("✅ /api/v1/register route ready")

		api.POST("/login", handlers.Login(db))
		fmt.Println("✅ /api/v1/login route ready")

		protected := api.Group("")
		protected.Use(middlewares.JWTAuthMiddleware())
		{
			protected.GET("/profile", handlers.Profile())
			fmt.Println("✅ /api/v1/profile (protected) route ready")

			// ✅ plans 分組
			plans := protected.Group("/plans")
			{

				sections := plans.Group("/sections")
				{
					sections.GET("", handlers.GetSections(db))
					fmt.Println("✅ /api/v1/plans/sections GET route ready")
					sections.POST("", handlers.CreateSection(db))
					fmt.Println("✅ /api/v1/plans/sections POST route ready")
					sections.DELETE("/:id", handlers.DeleteSection(db))
					fmt.Println("✅ /api/v1/plans/sections/:id DELETE route ready")
					sections.PUT("/:id", handlers.UpdateSection(db))
					fmt.Println("✅ /api/v1/sections PUT route ready")

				}
				tasks := plans.Group("/tasks")
				{
					tasks.POST("", handlers.CreateTask(db))
					fmt.Println("✅ /api/v1/plans/tasks POST route ready")
					tasks.PUT("/:id", handlers.UpdateTask(db))
					fmt.Println("✅ /api/v1/plans/tasks/:id PUT route ready")
					tasks.DELETE("/:id", handlers.DeleteTask(db))
					fmt.Println("✅ /api/v1/plans/tasks/:id DELETE route ready")
				}
				plans.GET("/sections-with-tasks", handlers.GetSectionsWithTasks(db))
				fmt.Println("✅ /api/v1/plans/sections-with-tasks GET route ready")
				plans.PUT("/sections-with-tasks", handlers.UpdateSectionsWithTasks(db))
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
	r.Run(":" + port)
}
