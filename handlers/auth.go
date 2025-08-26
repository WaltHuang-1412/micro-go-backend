package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Walter1412/micro-backend/models"
	"github.com/Walter1412/micro-backend/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Login godoc
// @Summary      使用者登入
// @Description  輸入 email 與密碼後登入並取得 JWT Token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        login  body  models.UserLoginInput  true  "登入資訊"
// @Success      200    {object}  map[string]string
// @Failure      400    {object}  map[string]string
// @Router       /login [post]
func Login(database *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		var input struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if error := context.ShouldBindJSON(&input); error != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		user, error := models.GetUserByEmail(database, input.Email)
		if error != nil {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		if error := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); error != nil {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect password"})
			return
		}

		// 🔐 建立 JWT token
		claims := jwt.MapClaims{
			"user_id":  user.ID,
			"username": user.Username,
			"exp":      time.Now().Add(time.Hour * 72).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "default_secret"
		}

		tokenString, error := token.SignedString([]byte(secret))
		if error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Token signing failed"})
			return
		}

		context.JSON(http.StatusOK, gin.H{"token": tokenString})
	}
}

// Register godoc
// @Summary      註冊使用者
// @Description  使用者註冊帳號
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        user  body  models.UserRegisterInput  true  "使用者資料"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Router       /register [post]
func Register(database *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		var input struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if error := context.ShouldBindJSON(&input); error != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		hashed, error := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Password hash failed"})
			return
		}

		user := models.User{
			Username:     input.Username,
			Email:        input.Email,
			PasswordHash: string(hashed),
		}

		if error := models.CreateUser(database, &user); error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "User creation failed"})
			return
		}

		context.JSON(http.StatusOK, gin.H{"message": "User registered"})
	}
}

// ForgotPassword godoc
// @Summary      忘記密碼
// @Description  發送重設密碼信件到用戶 email
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body  object{email=string}  true  "Email 地址"
// @Success      200    {object}  map[string]string
// @Failure      400    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Router       /forgot-password [post]
func ForgotPassword(database *sql.DB, emailService *services.EmailService) gin.HandlerFunc {
	return func(context *gin.Context) {
		var input struct {
			Email string `json:"email"`
		}

		if error := context.ShouldBindJSON(&input); error != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		user, error := models.GetUserByEmail(database, input.Email)
		if error != nil {
			fmt.Printf("🚨 GetUserByEmail error: %v\n", error)
			context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		fmt.Printf("✅ User found: ID=%d, Email=%s\n", user.ID, user.Email)

		passwordReset, error := models.CreatePasswordReset(database, user.ID)
		if error != nil {
			fmt.Printf("🚨 CreatePasswordReset error: %v\n", error)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reset token"})
			return
		}
		fmt.Printf("✅ Token created: %s\n", passwordReset.Token)

		error = emailService.SendPasswordResetEmail(user.Email, passwordReset.Token)
		if error != nil {
			fmt.Printf("🚨 SendPasswordResetEmail error: %v\n", error)
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
			return
		}
		fmt.Printf("✅ Email process completed\n")

		context.JSON(http.StatusOK, gin.H{"message": "Password reset email sent"})
	}
}

// ResetPassword godoc
// @Summary      重設密碼
// @Description  使用 token 重設用戶密碼
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body  object{token=string,new_password=string}  true  "重設資料"
// @Success      200    {object}  map[string]string
// @Failure      400    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Router       /reset-password [post]
func ResetPassword(database *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		var input struct {
			Token       string `json:"token"`
			NewPassword string `json:"new_password"`
		}

		if error := context.ShouldBindJSON(&input); error != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		passwordReset, error := models.GetPasswordResetByToken(database, input.Token)
		if error != nil {
			context.JSON(http.StatusNotFound, gin.H{"error": "Invalid or expired reset token"})
			return
		}

		hashed, error := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
		if error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Password hash failed"})
			return
		}

		error = models.UpdateUserPassword(database, passwordReset.UserID, string(hashed))
		if error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}

		error = models.MarkPasswordResetAsUsed(database, input.Token)
		if error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark token as used"})
			return
		}

		context.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
	}
}

// GetLatestToken godoc
// @Summary      獲取最新的重設密碼 token (僅供開發測試)
// @Description  返回最新的未使用密碼重設 token，僅供開發環境測試使用
// @Tags         Auth
// @Produce      json
// @Success      200    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Router       /dev/latest-token [get]
func GetLatestToken(database *sql.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		row := database.QueryRow("SELECT token, user_id FROM password_resets WHERE used = 0 ORDER BY created_at DESC LIMIT 1")
		
		var token string
		var userID int
		error := row.Scan(&token, &userID)
		if error != nil {
			context.JSON(http.StatusNotFound, gin.H{"error": "No unused tokens found"})
			return
		}

		context.JSON(http.StatusOK, gin.H{
			"token": token,
			"user_id": userID,
			"note": "This endpoint is for development testing only",
		})
	}
}

