package routes

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/Walter1412/micro-backend/handlers"
)

func RegisterPlanRoutes(router *gin.RouterGroup, database *sql.DB) {
	plans := router.Group("/plans")
	{
		sections := plans.Group("/sections")
		{
			sections.GET("", handlers.GetSections(database))
			sections.POST("", handlers.CreateSection(database))
			sections.DELETE("/:id", handlers.DeleteSection(database))
			sections.PUT("/:id", handlers.UpdateSection(database))
		}

		tasks := plans.Group("/tasks")
		{
			tasks.POST("", handlers.CreateTask(database))
			tasks.PUT("/:id", handlers.UpdateTask(database))
			tasks.DELETE("/:id", handlers.DeleteTask(database))
		}

		plans.GET("/sections-with-tasks", handlers.GetSectionsWithTasks(database))
		plans.PUT("/sections-with-tasks", handlers.UpdateSectionsWithTasks(database))
	}
}