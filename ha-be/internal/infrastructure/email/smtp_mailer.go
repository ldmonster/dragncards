package email

import (
	"fmt"
	"net/smtp"
	"strings"
)

type Mailer interface {
	SendConfirmation(to, token string) error
	SendReset(to, token string) error
}

type SMTPMailer struct {
	from     string
	host     string
	port     int
	username string
	password string
}

func NewSMTPMailer(from, host string, port int, username, password string) *SMTPMailer {
	return &SMTPMailer{from: from, host: host, port: port, username: username, password: password}
}

func (m *SMTPMailer) isConfigured() bool {
	return m.host != "" && m.port > 0 && m.from != ""
}

func (m *SMTPMailer) send(to, subject, body string) error {
	if !m.isConfigured() {
		return nil
	}
	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	msg := []byte("From: " + m.from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"\r\n" +
		body)
	return smtp.SendMail(addr, auth, m.from, []string{to}, msg)
}

func (m *SMTPMailer) SendConfirmation(to, token string) error {
	if token == "" || to == "" {
		return fmt.Errorf("invalid email token")
	}
	body := strings.Builder{}
	body.WriteString("Please confirm your account by using this token:\n")
	body.WriteString(token)
	return m.send(to, "Please confirm your DragnCards account", body.String())
}

func (m *SMTPMailer) SendReset(to, token string) error {
	if token == "" || to == "" {
		return fmt.Errorf("invalid email token")
	}
	body := strings.Builder{}
	body.WriteString("Use this password reset token:\n")
	body.WriteString(token)
	return m.send(to, "DragnCards password reset", body.String())
}
