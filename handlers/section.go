package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/Walter1412/micro-backend/models"
	"github.com/gin-gonic/gin"
)

// CreateSection godoc
// @Summary      建立新區塊（Section）
// @Description  建立一個新的區塊（自動補上 sort_order）
// @Tags         Plans
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        section  body  models.CreateSectionInput  true  "區塊資料"
// @Success      200      {object}  map[string]interface{}
// @Failure      400,500  {object}  map[string]string
// @Router       /plans/sections [post]
func CreateSection(database *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		var input models.CreateSectionInput
		if error := context.ShouldBindJSON(&input); error != nil {
			log.Printf("❌ Invalid input: %v", error)
			context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		userIdentifier := context.GetInt64("user_id") // 🔐 確保是 int64，避免型別問題

		// ✅ 取得目前使用者的最大 sort_order
		var maxSort sql.NullInt64
		error := database.QueryRow("SELECT MAX(sort_order) FROM sections WHERE user_id = ?", userIdentifier).Scan(&maxSort)
		if error != nil {
			log.Printf("❌ Failed to query max sort: %v", error)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get max sort"})
			return
		}

		newSort := 1
		if maxSort.Valid {
			newSort = int(maxSort.Int64) + 1
		}

		log.Printf("🧪 Creating section: user_id=%d, title=%s, sort_order=%d", userIdentifier, input.Title, newSort)

		// ✅ 插入資料
		result, error := database.Exec("INSERT INTO sections (user_id, title, sort_order) VALUES (?, ?, ?)", userIdentifier, input.Title, newSort)
		if error != nil {
			log.Printf("❌ Failed to insert section: %v", error)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create section"})
			return
		}

		insertedIdentifier, _ := result.LastInsertId()
		log.Printf("✅ Section created: ID=%d, Title=%s, Sort=%d, UserID=%d", insertedIdentifier, input.Title, newSort, userIdentifier)

		context.JSON(http.StatusOK, gin.H{
			"id":      insertedIdentifier,
			"title":   input.Title,
			"sort":    newSort,
			"user_id": userIdentifier,
		})
	}
}

// GetSections godoc
// @Summary      取得所有區塊（Section）
// @Description  依照排序列出所有區塊
// @Tags         Plans
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  models.Section
// @Failure      500  {object}  map[string]string
// @Router       /plans/sections [get]
func GetSections(database *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		userIdentifier := context.GetInt64("user_id") // ✅ 直接取得 int64 型別的 user_id

		rows, error := database.Query(`
			SELECT id, title, sort_order, created_at, updated_at
			FROM sections
			WHERE user_id = ?
			ORDER BY sort_order ASC`, userIdentifier)
		if error != nil {
			log.Printf("❌ Failed to query sections: %v", error)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sections"})
			return
		}
		defer rows.Close()

		var sections []models.Section
		for rows.Next() {
			var section models.Section
			if error := rows.Scan(&section.ID, &section.Title, &section.SortOrder, &section.CreatedAt, &section.UpdatedAt); error != nil {
				log.Printf("❌ Failed to scan section: %v", error)
				continue
			}
			sections = append(sections, section)
		}

		context.JSON(http.StatusOK, sections)
	}
}

// DeleteSection godoc
// @Summary      刪除區塊（Section）
// @Description  根據 ID 刪除一個區塊，並重新排序該使用者的其他區塊
// @Tags         Plans
// @Security     BearerAuth
// @Param        id  path  int  true  "Section ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /plans/sections/{id} [delete]
func DeleteSection(database *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		userIdentifier := context.GetInt64("user_id")
		identifier := context.Param("id")

		// 1️⃣ 驗證該 section 是否屬於目前登入者
		var exists bool
		error := database.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM sections WHERE id = ? AND user_id = ?
			)
		`, identifier, userIdentifier).Scan(&exists)
		if error != nil || !exists {
			log.Printf("❌ Section %s not found or not owned by user %d", identifier, userIdentifier)
			context.JSON(http.StatusBadRequest, gin.H{"error": "Section not found or unauthorized"})
			return
		}

		// 2️⃣ 刪除該 section
		_, error = database.Exec("DELETE FROM sections WHERE id = ? AND user_id = ?", identifier, userIdentifier)
		if error != nil {
			log.Printf("❌ Failed to delete section %s: %v", identifier, error)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete section"})
			return
		}

		// 3️⃣ 重新初始化排序變數
		_, error = database.Exec("SET @rank := 0")
		if error != nil {
			log.Printf("❌ Failed to reset rank variable")
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Section deleted, but failed to reorder"})
			return
		}

		// 4️⃣ 重排該使用者的 sections 排序
		_, error = database.Exec(`
			UPDATE sections
			SET sort_order = (@rank := @rank + 1)
			WHERE user_id = ?
			ORDER BY sort_order ASC
		`, userIdentifier)
		if error != nil {
			log.Printf("❌ Failed to reorder sections for user %d: %v", userIdentifier, error)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Section deleted, but failed to reorder"})
			return
		}

		log.Printf("✅ Section deleted and reordered: ID=%s, UserID=%d", identifier, userIdentifier)
		context.JSON(http.StatusOK, gin.H{"message": "Section deleted and reordered"})
	}
}

// UpdateSection godoc
// @Summary      更新區塊（Section 標題）
// @Description  根據 ID 修改區塊的標題，僅限本人操作
// @Tags         Plans
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path     int                       true  "Section ID"
// @Param        section body     models.UpdateSectionInput true  "更新資料"
// @Success      200     {object} map[string]interface{}
// @Failure      400     {object} map[string]string
// @Failure      500     {object} map[string]string
// @Router       /plans/sections/{id} [put]
func UpdateSection(database *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		identifier := context.Param("id")
		userIdentifier := context.GetInt64("user_id")

		var input models.UpdateSectionInput
		if error := context.ShouldBindJSON(&input); error != nil {
			log.Printf("❌ Invalid input: %v", error)
			context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// ✅ 確認該 section 是該使用者的
		var exists bool
		error := database.QueryRow("SELECT EXISTS (SELECT 1 FROM sections WHERE id = ? AND user_id = ?)", identifier, userIdentifier).Scan(&exists)
		if error != nil || !exists {
			log.Printf("❌ Section %s not found or not owned by user %d", identifier, userIdentifier)
			context.JSON(http.StatusBadRequest, gin.H{"error": "Section not found or unauthorized"})
			return
		}

		// ✅ 更新區塊
		_, error = database.Exec("UPDATE sections SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?", input.Title, identifier, userIdentifier)
		if error != nil {
			log.Printf("❌ Failed to update section title: %v", error)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update section"})
			return
		}

		log.Printf("✅ Section updated: ID=%s, Title=%s, UserID=%d", identifier, input.Title, userIdentifier)
		context.JSON(http.StatusOK, gin.H{
			"message": "Section updated",
			"id":      identifier,
			"title":   input.Title,
		})
	}
}

// GetSectionsWithTasks godoc
// @Summary      取得所有區塊（含任務）
// @Description  回傳每個區塊與其所屬任務（僅限本人），依照排序排列
// @Tags         Plans
// @Security     BearerAuth
// @Success      200  {array}  models.SectionWithTasks
// @Failure      500  {object}  map[string]string
// @Router       /plans/sections-with-tasks [get]
func GetSectionsWithTasks(database *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		userIdentifier := context.GetInt64("user_id")

		// 1️⃣ 查詢所有屬於該 user 的 sections
		sectionRows, error := database.Query(`
			SELECT id, title, sort_order, created_at, updated_at
			FROM sections
			WHERE user_id = ?
			ORDER BY sort_order ASC`, userIdentifier)
		if error != nil {
			log.Printf("❌ Failed to query sections: %v", error)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sections"})
			return
		}
		defer sectionRows.Close()

		sectionsMap := make(map[int64]*models.SectionWithTasks)
		var sectionIdentifiers []int64

		for sectionRows.Next() {
			var section models.SectionWithTasks
			if error := sectionRows.Scan(&section.ID, &section.Title, &section.SortOrder, &section.CreatedAt, &section.UpdatedAt); error != nil {
				log.Printf("❌ Failed to scan section: %v", error)
				continue
			}
			section.Tasks = []models.Task{}
			sectionsMap[section.ID] = &section
			sectionIdentifiers = append(sectionIdentifiers, section.ID)
		}

		if len(sectionIdentifiers) == 0 {
			context.JSON(http.StatusOK, []models.SectionWithTasks{})
			return
		}

		// 2️⃣ 查詢所有對應的 tasks
		query, args := buildTaskQuery(sectionIdentifiers)
		taskRows, error := database.Query(query, args...)
		if error != nil {
			log.Printf("❌ Failed to query tasks: %v", error)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
			return
		}
		defer taskRows.Close()

		for taskRows.Next() {
			var task models.Task
			if error := taskRows.Scan(&task.ID, &task.SectionID, &task.Content, &task.IsCompleted, &task.SortOrder, &task.CreatedAt, &task.UpdatedAt, &task.Title); error != nil {
				log.Printf("❌ Failed to scan task: %v", error)
				continue
			}
			if section, isValid := sectionsMap[task.SectionID]; isValid {
				section.Tasks = append(section.Tasks, task)
			}
		}

		// 3️⃣ 整理成 slice
		var result []models.SectionWithTasks
		for _, identifier := range sectionIdentifiers {
			result = append(result, *sectionsMap[identifier])
		}

		context.JSON(http.StatusOK, result)
	}
}

func buildTaskQuery(sectionIdentifiers []int64) (string, []interface{}) {
	query := `
		SELECT id, section_id, content, is_completed, sort_order, created_at, updated_at, title
		FROM tasks
		WHERE section_id IN (?` + strings.Repeat(",?", len(sectionIdentifiers)-1) + `)
		ORDER BY sort_order ASC`
	args := make([]interface{}, len(sectionIdentifiers))
	for index, identifier := range sectionIdentifiers {
		args[index] = identifier
	}
	return query, args
}

// UpdateSectionsWithTasks godoc
// @Summary      批次更新區塊與任務排序
// @Description  依據傳入資料更新 sections 與 tasks 的 sort_order（title/content 不會變動）
// @Tags         Plans
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  []models.SectionWithTasks  true  "排序資料"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /plans/sections-with-tasks [put]
func UpdateSectionsWithTasks(database *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		userIdentifier := context.GetInt64("user_id")

		var sections []models.SectionWithTasks
		if error := context.ShouldBindJSON(&sections); error != nil {
			log.Printf("❌ Invalid input: %v", error)
			context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		transaction, error := database.Begin()
		if error != nil {
			log.Printf("❌ Failed to begin transaction: %v", error)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "DB transaction error"})
			return
		}

		for index, section := range sections {
			// ✅ 檢查 section 是否屬於該使用者
			var ownerIdentifier int64
			error := transaction.QueryRow("SELECT user_id FROM sections WHERE id = ?", section.ID).Scan(&ownerIdentifier)
			if error != nil || ownerIdentifier != userIdentifier {
				transaction.Rollback()
				log.Printf("❌ Unauthorized section update or not found: section_id=%d, user_id=%d", section.ID, userIdentifier)
				context.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized section update"})
				return
			}

			// ✅ 更新 section 的排序
			_, error = transaction.Exec("UPDATE sections SET sort_order = ? WHERE id = ?", index+1, section.ID)
			if error != nil {
				transaction.Rollback()
				log.Printf("❌ Failed to update section sort_order: %v", error)
				context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update section sort"})
				return
			}

			// ✅ 處理每個 task
			for taskIndex, task := range section.Tasks {
				// ✅ 檢查 task 是否存在，並取得原 section_id
				var originalSectionIdentifier int64
				error := transaction.QueryRow("SELECT section_id FROM tasks WHERE id = ?", task.ID).Scan(&originalSectionIdentifier)
				if error != nil {
					transaction.Rollback()
					log.Printf("❌ Task not found: task_id=%d", task.ID)
					context.JSON(http.StatusBadRequest, gin.H{"error": "Task not found"})
					return
				}

				// ✅ 無論是否跨 section，一律更新 section_id + sort_order
				_, error = transaction.Exec("UPDATE tasks SET section_id = ?, sort_order = ? WHERE id = ?", section.ID, taskIndex+1, task.ID)
				if error != nil {
					transaction.Rollback()
					log.Printf("❌ Failed to update task (id=%d) sort/section: %v", task.ID, error)
					context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
					return
				}
			}
		}

		if error := transaction.Commit(); error != nil {
			log.Printf("❌ Failed to commit transaction: %v", error)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
			return
		}

		log.Println("✅ Sort orders and task-section updated successfully")
		context.JSON(http.StatusOK, gin.H{"message": "Sort orders updated"})
	}
}
