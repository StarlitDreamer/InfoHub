package delivery

import (
	"fmt"
	"mime"
	"net/smtp"
	"strings"
)

// SMTPSender 定义发送 SMTP 邮件所需的最小能力，便于测试替换。
type SMTPSender func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error

// EmailSender 将 Markdown 日报发送到邮件收件人。
type EmailSender struct {
	host     string
	port     int
	username string
	password string
	from     string
	to       []string
	subject  string
	sendMail SMTPSender
}

// NewEmailSender 创建邮件发送器。
func NewEmailSender(host string, port int, username, password, from string, to []string, subject string) *EmailSender {
	if port <= 0 {
		port = 25
	}
	subject = strings.TrimSpace(subject)
	if subject == "" {
		subject = "InfoHub 每日报告"
	}

	normalizedTo := make([]string, 0, len(to))
	for _, recipient := range to {
		recipient = strings.TrimSpace(recipient)
		if recipient != "" {
			normalizedTo = append(normalizedTo, recipient)
		}
	}

	return &EmailSender{
		host:     strings.TrimSpace(host),
		port:     port,
		username: strings.TrimSpace(username),
		password: password,
		from:     strings.TrimSpace(from),
		to:       normalizedTo,
		subject:  subject,
		sendMail: smtp.SendMail,
	}
}

// Send 发送 Markdown 日报邮件。
func (s *EmailSender) Send(markdown string) error {
	if s.host == "" || s.from == "" || len(s.to) == 0 {
		return fmt.Errorf("邮件发送配置不完整")
	}

	var auth smtp.Auth
	if s.username != "" || s.password != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	message := s.buildMessage(markdown)
	return s.sendMail(s.address(), auth, s.from, s.to, []byte(message))
}

func (s *EmailSender) address() string {
	return fmt.Sprintf("%s:%d", s.host, s.port)
}

func (s *EmailSender) buildMessage(markdown string) string {
	headers := []string{
		fmt.Sprintf("From: %s", s.from),
		fmt.Sprintf("To: %s", strings.Join(s.to, ", ")),
		fmt.Sprintf("Subject: %s", mime.BEncoding.Encode("UTF-8", s.subject)),
		"MIME-Version: 1.0",
		"Content-Type: text/markdown; charset=UTF-8",
	}

	return strings.Join(headers, "\r\n") + "\r\n\r\n" + markdown
}
