package handlers

import (
	"testing"
	"net/http/httptest"
	"strings"
	"github.com/gin-gonic/gin"
)

func TestProfile(t *testing.T) {
	// TODO: 實作 Profile 的測試
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// 設置測試數據
	// req := httptest.NewRequest("POST", "/api/v1/test", strings.NewReader(`{}`))
	// w := httptest.NewRecorder()
	// router.ServeHTTP(w, req)
	
	// 驗證結果
	// if w.Code != http.StatusOK {
	//     t.Errorf("Expected status 200, got %d", w.Code)
	// }
}

