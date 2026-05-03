package delivery

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhookSenderSend(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if err := NewWebhookSender(server.URL, server.Client()).Send("# 今日信息"); err != nil {
		t.Fatalf("Webhook 推送失败：%v", err)
	}

	if !called {
		t.Fatal("期望 Webhook 服务被调用")
	}
}
