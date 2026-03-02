package email

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

// EmailService Define the interface for the email sending service
type EmailService interface {
	SendMail(to []string, subject string, body string) error
}

// NewEmailService Create and return an EmailService instance
func NewEmailClient(username, password, host string, port int) EmailService {
	return &emailServiceImpl{
		username: username,
		password: password,
		host:     host,
		port:     port,
	}
}

type emailServiceImpl struct {
	username string
	password string
	host     string
	port     int
}

// SendMail Implement the methods of the EmailService interface for sending emails
func (e *emailServiceImpl) SendMail(to []string, subject string, body string) error {
	msg := email.NewEmail()
	msg.From = e.username
	msg.To = to
	msg.Subject = subject
	msg.Text = []byte(body)

	smtpHost := fmt.Sprintf("%s:%d", e.host, e.port)
	err := msg.Send(smtpHost, smtp.PlainAuth("", e.username, e.password, e.host))
	if err != nil {
		return err
	}

	return nil
}
