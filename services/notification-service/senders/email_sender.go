package email

import (
	"context"
	"gopkg.in/gomail.v2"
	"internet-shop/services/notification-service/handlers"
	"log"
)

type EmailSender struct {
	from     string
	username string
	password string
	smtpHost string
	smtpPort int
	auth     *gomail.Dialer
}

func NewEmailSender(from, username, password, smtpHost string, smtpPort int) *EmailSender {
	auth := gomail.NewDialer(smtpHost, smtpPort, username, password)
	return &EmailSender{
		username: username,
		password: password,
		from:     from,
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		auth:     auth,
	}
}

func (s *EmailSender) SendEmail(userID, orderID int64, subject, content string, h handlers.NotificationHandler) error {
	email, err := h.GetEmailById(context.Background(), userID)
	if err != nil {
		log.Fatal("email_sender: sendemail: getemailbyid: error: %v", err)
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", content)

	if err := s.auth.DialAndSend(m); err != nil {
		log.Fatalf("email_sender: sendemail: dialandsend: Error sending email: %v", err)
		return err
	}

	return nil
}
