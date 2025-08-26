package services

import (
	"fmt"
	"net/smtp"

	"github.com/Walter1412/micro-backend/config"
)

type EmailService struct {
	config config.EmailConfig
}

func NewEmailService(cfg config.EmailConfig) *EmailService {
	return &EmailService{
		config: cfg,
	}
}

func (e *EmailService) SendPasswordResetEmail(toEmail, token string) error {
	if e.config.SMTPHost == "" || e.config.SMTPUsername == "" {
		// 開發模式：只是記錄 token，不真的發送郵件
		fmt.Printf("🔧 [DEV MODE] Password reset token for %s: %s\n", toEmail, token)
		fmt.Printf("🔧 [DEV MODE] Reset URL: http://localhost:3000/reset-password?token=%s\n", token)
		return nil // 開發環境下不返回錯誤
	}

	resetURL := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", token)
	
	subject := "Password Reset Request"
	body := fmt.Sprintf(`
Dear User,

You have requested to reset your password. Please click the link below to reset your password:

%s

This link will expire in 1 hour.

If you did not request this password reset, please ignore this email.

Best regards,
Your App Team
`, resetURL)

	message := fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body)

	auth := smtp.PlainAuth("", e.config.SMTPUsername, e.config.SMTPPassword, e.config.SMTPHost)
	
	err := smtp.SendMail(
		e.config.SMTPHost+":"+e.config.SMTPPort,
		auth,
		e.config.FromEmail,
		[]string{toEmail},
		[]byte(message),
	)

	return err
}

func (e *EmailService) SendWelcomeEmail(toEmail, username string) error {
	if e.config.SMTPHost == "" || e.config.SMTPUsername == "" {
		return fmt.Errorf("email configuration not set")
	}
	
	subject := "Welcome to Our Platform"
	body := fmt.Sprintf(`
Dear %s,

Welcome to our platform! Your account has been successfully created.

If you have any questions, feel free to contact our support team.

Best regards,
Your App Team
`, username)

	message := fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body)

	auth := smtp.PlainAuth("", e.config.SMTPUsername, e.config.SMTPPassword, e.config.SMTPHost)
	
	err := smtp.SendMail(
		e.config.SMTPHost+":"+e.config.SMTPPort,
		auth,
		e.config.FromEmail,
		[]string{toEmail},
		[]byte(message),
	)

	return err
}