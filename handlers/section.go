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

		// ✅ 取得目前最大的 sort_order
		var maxSort sql.NullInt64
		err := db.QueryRow("SELECT MAX(sort_order) FROM sections").Scan(&maxSort)
		if err != nil {
			log.Printf("❌ Failed to query max sort: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get max sort"})
			return
		}

		newSort := 1
		if maxSort.Valid {
			newSort = int(maxSort.Int64) + 1
		}

		// ✅ 插入資料
		res, err := db.Exec("INSERT INTO sections (title, sort_order) VALUES (?, ?)", input.Title, newSort)
		if err != nil {
			log.Printf("❌ Failed to insert section: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create section"})
			return
		}

		insertedID, _ := res.LastInsertId()
		log.Printf("✅ Section created: ID=%d, Title=%s, Sort=%d", insertedID, input.Title, newSort)
		c.JSON(http.StatusOK, gin.H{
			"id":    insertedID,
			"title": input.Title,
			"sort":  newSort,
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
		rows, err := db.Query("SELECT id, title, sort_order, created_at, updated_at FROM sections ORDER BY sort_order ASC")
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
// @Description  根據 ID 刪除一個區塊，並重新排序其他區塊
// @Tags         Plans
// @Security     BearerAuth
// @Param        id  path  int  true  "Section ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /plans/sections/{id} [delete]
func DeleteSection(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// 先刪除指定 section
		_, err := db.Exec("DELETE FROM sections WHERE id = ?", id)
		if err != nil {
			log.Printf("❌ Failed to delete section %s: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete section"})
			return
		}

		// 重新初始化排序變數
		_, err = db.Exec("SET @rank := 0")
		if err != nil {
			log.Printf("❌ Failed to reset rank variable: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Section deleted, but failed to reorder"})
			return
		}

		// 重新排序 sort_order 欄位
		_, err = db.Exec("UPDATE sections SET sort_order = (@rank := @rank + 1) ORDER BY sort_order")
		if err != nil {
			log.Printf("❌ Failed to reorder sections: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Section deleted, but failed to reorder"})
			return
		}

		log.Printf("✅ Section deleted and reordered: ID=%s", id)
		c.JSON(http.StatusOK, gin.H{"message": "Section deleted and reordered"})
	}
}

// UpdateSection godoc
// @Summary      更新區塊（Section 標題）
// @Description  根據 ID 修改區塊的標題
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

		var input models.UpdateSectionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			log.Printf("❌ Invalid input: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		_, err := db.Exec("UPDATE sections SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", input.Title, id)
		if err != nil {
			log.Printf("❌ Failed to update section title: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update section"})
			return
		}

		log.Printf("✅ Section updated: ID=%s, Title=%s", id, input.Title)
		c.JSON(http.StatusOK, gin.H{
			"message": "Section updated",
			"id":      id,
			"title":   input.Title,
		})
	}
}

// GetSectionsWithTasks godoc
// @Summary      取得所有區塊（含任務）
// @Description  回傳每個區塊與其所屬任務，依照排序排列
// @Tags         Plans
// @Security     BearerAuth
// @Success      200  {array}  models.SectionWithTasks
// @Failure      500  {object}  map[string]string
// @Router       /plans/sections-with-tasks [get]
func GetSectionsWithTasks(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1️⃣ 查詢所有 sections
		sectionRows, err := db.Query(`
			SELECT id, title, sort_order, created_at, updated_at
			FROM sections
			ORDER BY sort_order ASC`)
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

		// 更新 sections 的 sort_order
		for i, s := range sections {
			_, err := tx.Exec("UPDATE sections SET sort_order = ? WHERE id = ?", i+1, s.ID)
			if err != nil {
				tx.Rollback()
				log.Printf("❌ Failed to update section sort_order: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update section sort"})
				return
			}

			// 更新 tasks 的 sort_order
			for j, t := range s.Tasks {
				_, err := tx.Exec("UPDATE tasks SET sort_order = ? WHERE id = ?", j+1, t.ID)
				if err != nil {
					tx.Rollback()
					log.Printf("❌ Failed to update task sort_order: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task sort"})
					return
				}
			}
		}

		if err := tx.Commit(); err != nil {
			log.Printf("❌ Failed to commit transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
			return
		}

		log.Println("✅ Sort orders updated successfully")
		c.JSON(http.StatusOK, gin.H{"message": "Sort orders updated"})
	}
}
