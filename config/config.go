package config

import (
	"fmt"
	"os"
)

type Config struct {
	// Database configuration
	DB DBConfig

	// Server configuration  
	Server ServerConfig

	// Swagger configuration
	Swagger SwaggerConfig

	// Email configuration
	Email EmailConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type ServerConfig struct {
	Port       string
	JWTSecret  string
	FrontendOrigin string
}

type SwaggerConfig struct {
	Host   string
	Scheme string
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

func LoadConfig() *Config {
	return &Config{
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "app_db"),
		},
		Server: ServerConfig{
			Port:       getEnv("PORT", "8088"),
			JWTSecret:  getEnv("JWT_SECRET", ""),
			FrontendOrigin: getEnv("FRONTEND_ORIGIN", ""),
		},
		Swagger: SwaggerConfig{
			Host:   getEnv("SWAGGER_HOST", "localhost:8088"),
			Scheme: getEnv("SWAGGER_SCHEME", "http"),
		},
		Email: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", ""),
			SMTPPort:     getEnv("SMTP_PORT", "587"),
			SMTPUsername: getEnv("SMTP_USERNAME", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			FromEmail:    getEnv("FROM_EMAIL", ""),
			FromName:     getEnv("FROM_NAME", ""),
		},
	}
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", 
		c.DB.User, c.DB.Password, c.DB.Host, c.DB.Port, c.DB.Name)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}