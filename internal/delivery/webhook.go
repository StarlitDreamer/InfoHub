package delivery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// WebhookSender 将日报推送到 Webhook。
type WebhookSender struct {
	url    string
	client *http.Client
}

// NewWebhookSender 创建 Webhook 推送器。
func NewWebhookSender(url string, client *http.Client) *WebhookSender {
	if client == nil {
		client = http.DefaultClient
	}

	return &WebhookSender{url: url, client: client}
}

// Send 发送 Markdown 内容。
func (s *WebhookSender) Send(markdown string) error {
	body, err := json.Marshal(map[string]string{"text": markdown})
	if err != nil {
		return err
	}

	resp, err := s.client.Post(s.url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("Webhook 推送失败，状态码：%d", resp.StatusCode)
	}

	return nil
}
