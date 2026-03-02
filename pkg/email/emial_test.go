package email

import (
	"fmt"
	"log"
	"testing"
)

// go test -v -run=Test_email .
func Test_email(t *testing.T) {
	emailSvc := NewEmailClient(
		"xx@qq.com",
		"xxxx",
		"smtp.qq.com",
		587,
	)

	to := []string{"xxx@qq.com"}
	subject := "Hello from Golang"
	body := "Hello!\nThis is a test email sent using Golang."

	err := emailSvc.SendMail(to, subject, body)
	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}

	fmt.Println("Email sent successfully!")
}
