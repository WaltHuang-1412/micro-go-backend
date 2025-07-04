package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Profile godoc
// @Summary      取得個人資訊
// @Description  使用 JWT 取得當前登入者資訊
// @Tags         user
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]string
// @Router       /profile [get]
func Profile() gin.HandlerFunc {
	return func(context *gin.Context) {
		userIdentifier, exists := context.Get("user_id")
		if !exists {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "user_id not found"})
			return
		}
		username, exists := context.Get("username")
		if !exists {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "username not found"})
			return
		}

		// ✅ 實務上可以從 DB 查 user 資料，這邊簡化直接回傳 ID
		context.JSON(http.StatusOK, gin.H{
			"user_id":  userIdentifier,
			"username": username,
			"message":  "You are authenticated!",
		})
	}
}
