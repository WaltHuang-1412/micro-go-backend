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
func Login(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		user, err := models.GetUserByEmail(db, input.Email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect password"})
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

		tokenString, err := token.SignedString([]byte(secret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token signing failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": tokenString})
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
func Register(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Password hash failed"})
			return
		}

		user := models.User{
			Username:     input.Username,
			Email:        input.Email,
			PasswordHash: string(hashed),
		}

		if err := models.CreateUser(db, &user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User creation failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User registered"})
	}
}
