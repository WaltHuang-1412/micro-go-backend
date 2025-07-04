package models

import (
	"database/sql"
	"time"
)

type UserRegisterInput struct {
	Username string `json:"username" example:"walter"`
	Email    string `json:"email" example:"w@w.com"`
	Password string `json:"password" example:"123456"`
}

type UserLoginInput struct {
	Email    string `json:"email" example:"w@w.com"`
	Password string `json:"password" example:"123456"`
}

type User struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

func CreateUser(database *sql.DB, user *User) error {
	_, error := database.Exec(
		"INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)",
		user.Username, user.Email, user.PasswordHash,
	)
	return error
}

func GetUserByEmail(database *sql.DB, email string) (*User, error) {
	row := database.QueryRow("SELECT id, username, email, password_hash, created_at FROM users WHERE email = ?", email)

	var user User
	error := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if error != nil {
		return nil, error
	}
	return &user, nil
}
