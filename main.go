package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Walter1412/micro-backend/handlers"
	"github.com/Walter1412/micro-backend/middlewares"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("❌ Failed to connect to DB:", err)
	}
	defer db.Close()

	// 🔁 自動重試連線最多 10 次
	for i := 0; i < 10; i++ {
		err := db.Ping()
		if err == nil {
			fmt.Println("✅ Connected to DB!")
			break
		}
		fmt.Printf("⏳ Waiting for DB to be ready... (attempt %d/10)\n", i+1)
		time.Sleep(2 * time.Second)
		if i == 9 {
			log.Fatal("❌ DB not reachable after retrying:", err)
		}
	}

	r := gin.Default()

	// ✅ API 路徑加上版本
	api := r.Group("/api/v1")
	{
		api.POST("/register", handlers.Register(db))
		api.POST("/login", handlers.Login(db))

		protected := api.Group("")
		protected.Use(middlewares.JWTAuthMiddleware())
		{
			protected.GET("/profile", handlers.Profile())
		}
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}

	fmt.Println("🚀 Server running at http://localhost:" + port)
	r.Run(":" + port)
}
