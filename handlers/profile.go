package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Profile() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
			return
		}

		// ✅ 實務上可以從 DB 查 user 資料，這邊簡化直接回傳 ID
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"message": "You are authenticated!",
		})
	}
}
