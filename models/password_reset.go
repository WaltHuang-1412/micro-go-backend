package models

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"
)

type PasswordReset struct {
	ID        int
	UserID    int
	Token     string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}

func CreatePasswordReset(database *sql.DB, userID int) (*PasswordReset, error) {
	token, err := generateResetToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Hour * 1) // 1 hour expiration

	_, err = database.Exec(
		"INSERT INTO password_resets (user_id, token, expires_at) VALUES (?, ?, ?)",
		userID, token, expiresAt,
	)
	if err != nil {
		return nil, err
	}

	return &PasswordReset{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		Used:      false,
		CreatedAt: time.Now(),
	}, nil
}

func GetPasswordResetByToken(database *sql.DB, token string) (*PasswordReset, error) {
	row := database.QueryRow(
		"SELECT id, user_id, token, expires_at, used, created_at FROM password_resets WHERE token = ? AND used = FALSE AND expires_at > NOW()",
		token,
	)

	var reset PasswordReset
	err := row.Scan(&reset.ID, &reset.UserID, &reset.Token, &reset.ExpiresAt, &reset.Used, &reset.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &reset, nil
}

func MarkPasswordResetAsUsed(database *sql.DB, token string) error {
	_, err := database.Exec(
		"UPDATE password_resets SET used = TRUE WHERE token = ?",
		token,
	)
	return err
}

func CleanupExpiredPasswordResets(database *sql.DB) error {
	_, err := database.Exec(
		"DELETE FROM password_resets WHERE expires_at < NOW() OR used = TRUE",
	)
	return err
}

func generateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}