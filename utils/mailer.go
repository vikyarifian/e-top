package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendEmail(to, subject, body string) error {
	// Ganti dengan email Gmail kamu
	from := os.Getenv("EMAIL_FROM")

	// App Password 16 digit (bukan password Gmail biasa!)
	password := os.Getenv("EMAIL_PASS")

	// Gmail SMTP server
	smtpHost := os.Getenv("EMAIL_HOST")
	smtpPort := os.Getenv("EMAIL_PORT")

	// Header email
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	// Gabungkan header + body
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Auth
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Send
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(message))
	if err != nil {
		return err
	}
	return nil
}

// Email verifikasi akun
func SendVerificationEmail(to, verifyURL string) error {
	subject := `[[etop]] Verify Your Account`

	body := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
	  <meta charset="UTF-8">
	  <style>
		body { font-family: Arial, sans-serif; background:#f9fafb; color:#111; }
		.container { max-width:600px; margin:auto; background:white; padding:20px; border-radius:8px; }
		.btn {
		  display:inline-block; 
		  padding:10px 20px; 
		  background:#2563eb; 
		  color:white; 
		  text-decoration:none; 
		  border-radius:6px;
		}
		.footer { font-size:12px; color:#555; margin-top:20px; }
	  </style>
	</head>
	<body>
	  <div class="container">
		<h2>Welcome to etop 🚀</h2>
		<p>Hi there,</p>
		<p>Thank you for signing up. Please confirm your email address to activate your account.</p>
		<p style="text-align:center;">
		  <a href="%s" class="btn">Verify My Account</a>
		</p>
		<p>If the button doesn’t work, copy and paste this link into your browser:</p>
		<p><a href="%s">%s</a></p>
		<div class="footer">
		  <p>© 2025 EBER IT Department. All rights reserved.</p>
		  <p>If you didn’t sign up, please ignore this email.</p>
		</div>
	  </div>
	</body>
	</html>
	`, verifyURL, verifyURL, verifyURL)

	return SendEmail(to, subject, body)
}
