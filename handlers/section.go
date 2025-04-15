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
func CreateSection(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.CreateSectionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			log.Printf("❌ Invalid input: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		userID := c.GetInt64("user_id") // 🔐 確保是 int64，避免型別問題

		// ✅ 取得目前使用者的最大 sort_order
		var maxSort sql.NullInt64
		err := db.QueryRow("SELECT MAX(sort_order) FROM sections WHERE user_id = ?", userID).Scan(&maxSort)
		if err != nil {
			log.Printf("❌ Failed to query max sort: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get max sort"})
			return
		}

		newSort := 1
		if maxSort.Valid {
			newSort = int(maxSort.Int64) + 1
		}

		log.Printf("🧪 Creating section: user_id=%d, title=%s, sort_order=%d", userID, input.Title, newSort)

		// ✅ 插入資料
		res, err := db.Exec("INSERT INTO sections (user_id, title, sort_order) VALUES (?, ?, ?)", userID, input.Title, newSort)
		if err != nil {
			log.Printf("❌ Failed to insert section: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create section"})
			return
		}

		insertedID, _ := res.LastInsertId()
		log.Printf("✅ Section created: ID=%d, Title=%s, Sort=%d, UserID=%d", insertedID, input.Title, newSort, userID)

		c.JSON(http.StatusOK, gin.H{
			"id":      insertedID,
			"title":   input.Title,
			"sort":    newSort,
			"user_id": userID,
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
func GetSections(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id") // ✅ 直接取得 int64 型別的 user_id

		rows, err := db.Query(`
			SELECT id, title, sort_order, created_at, updated_at
			FROM sections
			WHERE user_id = ?
			ORDER BY sort_order ASC`, userID)
		if err != nil {
			log.Printf("❌ Failed to query sections: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sections"})
			return
		}
		defer rows.Close()

		var sections []models.Section
		for rows.Next() {
			var s models.Section
			if err := rows.Scan(&s.ID, &s.Title, &s.SortOrder, &s.CreatedAt, &s.UpdatedAt); err != nil {
				log.Printf("❌ Failed to scan section: %v", err)
				continue
			}
			sections = append(sections, s)
		}

		c.JSON(http.StatusOK, sections)
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
func DeleteSection(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		id := c.Param("id")

		// 1️⃣ 驗證該 section 是否屬於目前登入者
		var exists bool
		err := db.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM sections WHERE id = ? AND user_id = ?
			)
		`, id, userID).Scan(&exists)
		if err != nil || !exists {
			log.Printf("❌ Section %s not found or not owned by user %d", id, userID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Section not found or unauthorized"})
			return
		}

		// 2️⃣ 刪除該 section
		_, err = db.Exec("DELETE FROM sections WHERE id = ? AND user_id = ?", id, userID)
		if err != nil {
			log.Printf("❌ Failed to delete section %s: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete section"})
			return
		}

		// 3️⃣ 重新初始化排序變數
		_, err = db.Exec("SET @rank := 0")
		if err != nil {
			log.Printf("❌ Failed to reset rank variable")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Section deleted, but failed to reorder"})
			return
		}

		// 4️⃣ 重排該使用者的 sections 排序
		_, err = db.Exec(`
			UPDATE sections
			SET sort_order = (@rank := @rank + 1)
			WHERE user_id = ?
			ORDER BY sort_order ASC
		`, userID)
		if err != nil {
			log.Printf("❌ Failed to reorder sections for user %d: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Section deleted, but failed to reorder"})
			return
		}

		log.Printf("✅ Section deleted and reordered: ID=%s, UserID=%d", id, userID)
		c.JSON(http.StatusOK, gin.H{"message": "Section deleted and reordered"})
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
func UpdateSection(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := c.GetInt64("user_id")

		var input models.UpdateSectionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			log.Printf("❌ Invalid input: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// ✅ 確認該 section 是該使用者的
		var exists bool
		err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM sections WHERE id = ? AND user_id = ?)", id, userID).Scan(&exists)
		if err != nil || !exists {
			log.Printf("❌ Section %s not found or not owned by user %d", id, userID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Section not found or unauthorized"})
			return
		}

		// ✅ 更新區塊
		_, err = db.Exec("UPDATE sections SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?", input.Title, id, userID)
		if err != nil {
			log.Printf("❌ Failed to update section title: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update section"})
			return
		}

		log.Printf("✅ Section updated: ID=%s, Title=%s, UserID=%d", id, input.Title, userID)
		c.JSON(http.StatusOK, gin.H{
			"message": "Section updated",
			"id":      id,
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
func GetSectionsWithTasks(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")

		// 1️⃣ 查詢所有屬於該 user 的 sections
		sectionRows, err := db.Query(`
			SELECT id, title, sort_order, created_at, updated_at
			FROM sections
			WHERE user_id = ?
			ORDER BY sort_order ASC`, userID)
		if err != nil {
			log.Printf("❌ Failed to query sections: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sections"})
			return
		}
		defer sectionRows.Close()

		sectionsMap := make(map[int64]*models.SectionWithTasks)
		var sectionIDs []int64

		for sectionRows.Next() {
			var s models.SectionWithTasks
			if err := sectionRows.Scan(&s.ID, &s.Title, &s.SortOrder, &s.CreatedAt, &s.UpdatedAt); err != nil {
				log.Printf("❌ Failed to scan section: %v", err)
				continue
			}
			s.Tasks = []models.Task{}
			sectionsMap[s.ID] = &s
			sectionIDs = append(sectionIDs, s.ID)
		}

		if len(sectionIDs) == 0 {
			c.JSON(http.StatusOK, []models.SectionWithTasks{})
			return
		}

		// 2️⃣ 查詢所有對應的 tasks
		query, args := buildTaskQuery(sectionIDs)
		taskRows, err := db.Query(query, args...)
		if err != nil {
			log.Printf("❌ Failed to query tasks: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
			return
		}
		defer taskRows.Close()

		for taskRows.Next() {
			var t models.Task
			if err := taskRows.Scan(&t.ID, &t.SectionID, &t.Content, &t.IsCompleted, &t.SortOrder, &t.CreatedAt, &t.UpdatedAt, &t.Title); err != nil {
				log.Printf("❌ Failed to scan task: %v", err)
				continue
			}
			if section, ok := sectionsMap[t.SectionID]; ok {
				section.Tasks = append(section.Tasks, t)
			}
		}

		// 3️⃣ 整理成 slice
		var result []models.SectionWithTasks
		for _, id := range sectionIDs {
			result = append(result, *sectionsMap[id])
		}

		c.JSON(http.StatusOK, result)
	}
}

func buildTaskQuery(sectionIDs []int64) (string, []interface{}) {
	query := `
		SELECT id, section_id, content, is_completed, sort_order, created_at, updated_at, title
		FROM tasks
		WHERE section_id IN (?` + strings.Repeat(",?", len(sectionIDs)-1) + `)
		ORDER BY sort_order ASC`
	args := make([]interface{}, len(sectionIDs))
	for i, id := range sectionIDs {
		args[i] = id
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
func UpdateSectionsWithTasks(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")

		var sections []models.SectionWithTasks
		if err := c.ShouldBindJSON(&sections); err != nil {
			log.Printf("❌ Invalid input: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		tx, err := db.Begin()
		if err != nil {
			log.Printf("❌ Failed to begin transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB transaction error"})
			return
		}

		for i, s := range sections {
			// ✅ 檢查 section 是否屬於該使用者
			var ownerID int64
			err := tx.QueryRow("SELECT user_id FROM sections WHERE id = ?", s.ID).Scan(&ownerID)
			if err != nil || ownerID != userID {
				tx.Rollback()
				log.Printf("❌ Unauthorized section update or not found: section_id=%d, user_id=%d", s.ID, userID)
				c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized section update"})
				return
			}

			// ✅ 更新 section 的排序
			_, err = tx.Exec("UPDATE sections SET sort_order = ? WHERE id = ?", i+1, s.ID)
			if err != nil {
				tx.Rollback()
				log.Printf("❌ Failed to update section sort_order: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update section sort"})
				return
			}

			// ✅ 處理每個 task
			for j, t := range s.Tasks {
				// ✅ 檢查 task 是否存在，並取得原 section_id
				var originalSectionID int64
				err := tx.QueryRow("SELECT section_id FROM tasks WHERE id = ?", t.ID).Scan(&originalSectionID)
				if err != nil {
					tx.Rollback()
					log.Printf("❌ Task not found: task_id=%d", t.ID)
					c.JSON(http.StatusBadRequest, gin.H{"error": "Task not found"})
					return
				}

				// ✅ 無論是否跨 section，一律更新 section_id + sort_order
				_, err = tx.Exec("UPDATE tasks SET section_id = ?, sort_order = ? WHERE id = ?", s.ID, j+1, t.ID)
				if err != nil {
					tx.Rollback()
					log.Printf("❌ Failed to update task (id=%d) sort/section: %v", t.ID, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
					return
				}
			}
		}

		if err := tx.Commit(); err != nil {
			log.Printf("❌ Failed to commit transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
			return
		}

		log.Println("✅ Sort orders and task-section updated successfully")
		c.JSON(http.StatusOK, gin.H{"message": "Sort orders updated"})
	}
}
