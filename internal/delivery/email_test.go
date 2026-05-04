package delivery

import (
	"net/smtp"
	"strings"
	"testing"
)

func TestEmailSenderSend(t *testing.T) {
	sender := NewEmailSender(
		"smtp.example.com",
		587,
		"user",
		"secret",
		"robot@example.com",
		[]string{"alice@example.com", "bob@example.com"},
		"每日报告",
	)

	called := false
	sender.sendMail = func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
		called = true
		if addr != "smtp.example.com:587" {
			t.Fatalf("unexpected smtp addr: %s", addr)
		}
		if auth == nil {
			t.Fatal("expected smtp auth to be configured")
		}
		if from != "robot@example.com" {
			t.Fatalf("unexpected from: %s", from)
		}
		if len(to) != 2 || to[0] != "alice@example.com" || to[1] != "bob@example.com" {
			t.Fatalf("unexpected recipients: %+v", to)
		}
		content := string(msg)
		if !strings.Contains(content, "Content-Type: text/markdown; charset=UTF-8") {
			t.Fatalf("expected markdown content type, got %s", content)
		}
		if !strings.Contains(content, "# 今日信息") {
			t.Fatalf("expected markdown body, got %s", content)
		}
		return nil
	}

	if err := sender.Send("# 今日信息"); err != nil {
		t.Fatalf("邮件发送失败：%v", err)
	}
	if !called {
		t.Fatal("期望邮件发送器被调用")
	}
}

func TestEmailSenderSendWithoutCredentials(t *testing.T) {
	sender := NewEmailSender(
		"smtp.example.com",
		25,
		"",
		"",
		"robot@example.com",
		[]string{"alice@example.com"},
		"",
	)

	sender.sendMail = func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
		if auth != nil {
			t.Fatal("did not expect smtp auth")
		}
		if !strings.Contains(string(msg), "Subject: =?UTF-8?b?SW5mb0h1YiDmr4/ml6XmiqXlkYo=?=") {
			t.Fatalf("expected default subject, got %s", string(msg))
		}
		return nil
	}

	if err := sender.Send("# 今日信息"); err != nil {
		t.Fatalf("邮件发送失败：%v", err)
	}
}

func TestEmailSenderSendRequiresCompleteConfig(t *testing.T) {
	sender := NewEmailSender("", 25, "", "", "", nil, "")

	if err := sender.Send("# 今日信息"); err == nil {
		t.Fatal("expected incomplete email config to fail")
	}
}
