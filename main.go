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
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"github.com/Walter1412/micro-backend/config"
	"github.com/Walter1412/micro-backend/docs"
	"github.com/Walter1412/micro-backend/routes"
)

func main() {
	// 載入配置
	configuration := config.LoadConfig()

	// 設定 Swagger 變數
	docs.SwaggerInfo.Host = configuration.Swagger.Host
	docs.SwaggerInfo.Schemes = []string{configuration.Swagger.Scheme}

	// 連接資料庫
	database, err := sql.Open("mysql", configuration.GetDSN())
	if err != nil {
		log.Fatal("❌ Failed to connect to DB:", err)
	}
	defer database.Close()

	// 自動重試 DB 連線
	maxRetries := 10
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := database.Ping(); err == nil {
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

	// 初始化路由
	router := gin.Default()
	routes.RegisterRoutes(router, database, configuration)

	fmt.Println("🚀 Server running at http://localhost:" + configuration.Server.Port)
	fmt.Println("🌐 Swagger UI available at http://localhost:" + configuration.Server.Port + "/swagger/index.html")
	router.Run(":" + configuration.Server.Port)
}
