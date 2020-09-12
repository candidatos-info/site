package email

import (
	"fmt"
	"net/smtp"
	"strings"
)

var (
	mimeHeaders = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
)

// Client defines the state of this object
type Client struct {
	Email    string
	password string
}

// New returns a new email instance
func New(email, password string) *Client {
	return &Client{
		Email:    email,
		password: password,
	}
}

// Send sends an email
func (e *Client) Send(from string, to []string, subject, body string) error {
	var emailBodyBuilder strings.Builder
	emailBodyBuilder.WriteString(fmt.Sprintf("Subject: %s \n%s\n\n", subject, mimeHeaders))
	emailBodyBuilder.WriteString(body)
	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", e.Email, e.password, "smtp.gmail.com"),
		e.Email, to, []byte(emailBodyBuilder.String()))
	if err != nil {
		return fmt.Errorf("falha ao enviar email, erro %q", err)
	}
	return nil
}
