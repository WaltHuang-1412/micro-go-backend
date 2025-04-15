package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/Walter1412/micro-backend/models"
	"github.com/gin-gonic/gin"
)

// CreateTask godoc
// @Summary      建立任務（Task）
// @Description  建立新的任務，並自動排序
// @Tags         Plans
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        task  body  models.CreateTaskInput  true  "任務內容"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]string
// @Router       /plans/tasks [post]
func CreateTask(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.CreateTaskInput
		if err := c.ShouldBindJSON(&input); err != nil {
			log.Printf("❌ Invalid input: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		var maxSort sql.NullInt64
		err := db.QueryRow("SELECT MAX(sort_order) FROM tasks WHERE section_id = ?", input.SectionID).Scan(&maxSort)
		if err != nil {
			log.Printf("❌ Failed to get max sort: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get max sort"})
			return
		}

		newSort := 1
		if maxSort.Valid {
			newSort = int(maxSort.Int64) + 1
		}

		now := time.Now()
		res, err := db.Exec(`
			INSERT INTO tasks (section_id, title, content, is_completed, sort_order, created_at, updated_at)
			VALUES (?, ?, ?, false, ?, ?, ?)`,
			input.SectionID, input.Title, input.Content, newSort, now, now,
		)
		if err != nil {
			log.Printf("❌ Failed to insert task: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
			return
		}

		id, _ := res.LastInsertId()
		log.Printf("✅ Task created: ID=%d, SectionID=%d", id, input.SectionID)
		c.JSON(http.StatusOK, gin.H{
			"id":           id,
			"section_id":   input.SectionID,
			"title":        input.Title,
			"content":      input.Content,
			"sort_order":   newSort,
			"is_completed": false,
		})
	}
}

// UpdateTask godoc
// @Summary      更新任務（Task）
// @Description  根據 ID 更新任務內容
// @Tags         Plans
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path  int                 true  "任務 ID"
// @Param        task  body  models.UpdateTaskInput true  "更新資料"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /plans/tasks/{id} [put]
func UpdateTask(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var input models.UpdateTaskInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		_, err := db.Exec(`
			UPDATE tasks
			SET title = ?, content = ?, is_completed = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?`, input.Title, input.Content, input.IsCompleted, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Task updated"})
	}
}

// DeleteTask godoc
// @Summary      刪除任務（Task）
// @Description  根據 ID 刪除任務，並重新排序同區塊內的任務
// @Tags         Plans
// @Security     BearerAuth
// @Param        id   path  int  true  "任務 ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /plans/tasks/{id} [delete]
func DeleteTask(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// 先查出 section_id
		var sectionID int64
		err := db.QueryRow("SELECT section_id FROM tasks WHERE id = ?", id).Scan(&sectionID)
		if err != nil {
			log.Printf("❌ Invalid task ID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
			return
		}

		// 刪除該任務
		_, err = db.Exec("DELETE FROM tasks WHERE id = ?", id)
		if err != nil {
			log.Printf("❌ Failed to delete task %s: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
			return
		}

		// ✅ 單一 SQL 完成重排
		_, err = db.Exec(`
			UPDATE tasks t
			JOIN (
				SELECT id, ROW_NUMBER() OVER (ORDER BY sort_order) AS new_sort
				FROM tasks
				WHERE section_id = ?
			) sorted
			ON t.id = sorted.id
			SET t.sort_order = sorted.new_sort;
		`, sectionID)
		if err != nil {
			log.Printf("❌ Failed to reorder tasks in section %d: %v", sectionID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Task deleted, but failed to reorder"})
			return
		}

		log.Printf("✅ Task deleted and reordered: ID=%s", id)
		c.JSON(http.StatusOK, gin.H{"message": "Task deleted and reordered"})
	}
}
