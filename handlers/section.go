package handlers

import (
	"database/sql"
	"log"
	"net/http"

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
