package utils

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

func SendVerificationEmail(toEmail, code string) error {
	smtpUser := os.Getenv("SMTP_USER")
	appPassword := os.Getenv("SMTP_APP_PASSWORD")

	if smtpUser == "" || appPassword == "" {
		return fmt.Errorf("email configuration is missing")
	}
	if strings.TrimSpace(toEmail) == "" || strings.TrimSpace(code) == "" {
		return fmt.Errorf("recipient email and verification code are required")
	}

	from := smtpUser
	host := "smtp.gmail.com"
	address := host + ":587"

	auth := smtp.PlainAuth("", smtpUser, appPassword, host)

	subject := "Your JobFlow password reset code"
	body := fmt.Sprintf(
		"Your JobFlow password reset code is: %s\n\n"+
			"This code expires in 15 minutes. "+
			"If you did not request a password reset, ignore this email.",
		code,
	)
	message := []byte(
		"From: JobFlow <" + from + ">\r\n" +
			"To: " + toEmail + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" +
			body + "\r\n",
	)
	if err := smtp.SendMail(address, auth, from, []string{toEmail}, message); err != nil {
		return fmt.Errorf("sending verification email: %w", err)
	}
	return nil
}
