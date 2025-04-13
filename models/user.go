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

func CreateUser(db *sql.DB, user *User) error {
	_, err := db.Exec(
		"INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)",
		user.Username, user.Email, user.PasswordHash,
	)
	return err
}

func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	row := db.QueryRow("SELECT id, username, email, password_hash, created_at FROM users WHERE email = ?", email)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
