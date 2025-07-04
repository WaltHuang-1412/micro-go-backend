package handlers

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/Walter1412/micro-backend/models"
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
