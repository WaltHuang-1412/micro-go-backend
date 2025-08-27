package middlewares

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var (
	// 全域限制器：每秒100個請求，突發200個（適合小型網站100-500用戶）
	globalLimiter = rate.NewLimiter(rate.Limit(100), 200)
)

// RateLimitMiddleware 全域請求頻率限制中間件
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !globalLimiter.Allow() {
			// 計算下次允許請求的等待時間
			reservation := globalLimiter.Reserve()
			delay := reservation.Delay()
			reservation.Cancel() // 取消預約，不實際等待
			
			retryAfterSeconds := int(delay.Seconds()) + 1 // 向上取整並加1秒緩衝
			
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfterSeconds))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": fmt.Sprintf("%ds", retryAfterSeconds),
				"message":     "Too many requests, please try again later",
			})
			return
		}
		c.Next()
	}
}