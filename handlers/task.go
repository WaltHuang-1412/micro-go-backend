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
func CreateTask(database *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		var input models.CreateTaskInput
		if error := context.ShouldBindJSON(&input); error != nil {
			log.Printf("❌ Invalid input: %v", error)
			context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		userIdentifier := context.GetInt64("user_id")

		// ✅ 驗證該 section 是否屬於該 user
		var ownerIdentifier int64
		error := database.QueryRow("SELECT user_id FROM sections WHERE id = ?", input.SectionID).Scan(&ownerIdentifier)
		if error != nil || ownerIdentifier != userIdentifier {
			log.Printf("❌ Unauthorized to access section_id=%d by user_id=%d", input.SectionID, userIdentifier)
			context.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to add task to this section"})
			return
		}

		// ✅ 查詢目前 section 下最大的 sort_order
		var maxSort sql.NullInt64
		error = database.QueryRow("SELECT MAX(sort_order) FROM tasks WHERE section_id = ?", input.SectionID).Scan(&maxSort)
		if error != nil {
			log.Printf("❌ Failed to get max sort: %v", error)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get max sort"})
			return
		}

		newSort := 1
		if maxSort.Valid {
			newSort = int(maxSort.Int64) + 1
		}

		now := time.Now()
		result, error := database.Exec(`
			INSERT INTO tasks (user_id, section_id, title, content, is_completed, sort_order, created_at, updated_at)
			VALUES (?, ?, ?, ?, false, ?, ?, ?)`,
			userIdentifier, input.SectionID, input.Title, input.Content, newSort, now, now,
		)
		if error != nil {
			log.Printf("❌ Failed to insert task: %v", error)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
			return
		}

		identifier, _ := result.LastInsertId()
		log.Printf("✅ Task created: ID=%d, SectionID=%d", identifier, input.SectionID)
		context.JSON(http.StatusOK, gin.H{
			"id":           identifier,
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
// @Failure      403   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /plans/tasks/{id} [put]
func UpdateTask(database *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		identifier := context.Param("id")
		userIdentifier := context.GetInt64("user_id") // ✅ 從 middleware 拿 user_id

		var input models.UpdateTaskInput
		if error := context.ShouldBindJSON(&input); error != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// ✅ 確認 task 是否屬於該 user
		var taskOwnerIdentifier int64
		error := database.QueryRow("SELECT user_id FROM tasks WHERE id = ?", identifier).Scan(&taskOwnerIdentifier)
		if error != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "Task not found"})
			return
		}
		if taskOwnerIdentifier != userIdentifier {
			context.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to modify this task"})
			return
		}

		// ✅ 更新 task
		_, error = database.Exec(`
			UPDATE tasks
			SET title = ?, content = ?, is_completed = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?`, input.Title, input.Content, input.IsCompleted, identifier)
		if error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
			return
		}

		context.JSON(http.StatusOK, gin.H{"message": "Task updated"})
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
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /plans/tasks/{id} [delete]
func DeleteTask(database *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		identifier := context.Param("id")
		userIdentifier := context.GetInt64("user_id") // ✅ 拿目前登入的 user_id

		// ✅ 查出 task 所屬的 section_id 與擁有者 user_id
		var sectionIdentifier int64
		var taskOwnerIdentifier int64
		error := database.QueryRow(`
			SELECT s.id, s.user_id
			FROM tasks t
			JOIN sections s ON t.section_id = s.id
			WHERE t.id = ?`, identifier).Scan(&sectionIdentifier, &taskOwnerIdentifier)
		if error != nil {
			log.Printf("❌ Invalid task ID or join failed: %v", error)
			context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
			return
		}

		// ✅ 檢查擁有權
		if taskOwnerIdentifier != userIdentifier {
			log.Printf("❌ Unauthorized to delete task ID=%s by user_id=%d", identifier, userIdentifier)
			context.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to delete this task"})
			return
		}

		// ✅ 刪除該任務
		_, error = database.Exec("DELETE FROM tasks WHERE id = ?", identifier)
		if error != nil {
			log.Printf("❌ Failed to delete task %s: %v", identifier, error)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
			return
		}

		// ✅ 單一 SQL 完成重排
		_, error = database.Exec(`
			UPDATE tasks t
			JOIN (
				SELECT id, ROW_NUMBER() OVER (ORDER BY sort_order) AS new_sort
				FROM tasks
				WHERE section_id = ?
			) sorted
			ON t.id = sorted.id
			SET t.sort_order = sorted.new_sort;
		`, sectionIdentifier)
		if error != nil {
			log.Printf("❌ Failed to reorder tasks in section %d: %v", sectionIdentifier, error)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Task deleted, but failed to reorder"})
			return
		}

		log.Printf("✅ Task deleted and reordered: ID=%s", identifier)
		context.JSON(http.StatusOK, gin.H{"message": "Task deleted and reordered"})
	}
}
