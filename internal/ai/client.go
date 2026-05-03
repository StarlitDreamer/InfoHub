package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"InfoHub-agent/internal/model"
)

// HTTPClient 调用 OpenAI 兼容接口生成真实摘要。
type HTTPClient struct {
	endpoint string
	apiKey   string
	model    string
	client   *http.Client
}

// NewHTTPClient 创建真实 AI 客户端。
func NewHTTPClient(endpoint, apiKey, modelName string, client *http.Client) *HTTPClient {
	if client == nil {
		client = http.DefaultClient
	}

	return &HTTPClient{
		endpoint: endpoint,
		apiKey:   apiKey,
		model:    modelName,
		client:   client,
	}
}

// Summarize 调用真实模型并按项目要求回填摘要。
func (c *HTTPClient) Summarize(item model.NewsItem) (model.NewsItem, error) {
	if c.endpoint == "" || c.apiKey == "" || c.model == "" {
		return item, errors.New("AI 客户端缺少 endpoint、apiKey 或 model")
	}

	body, err := json.Marshal(chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: "你是信息汇总 Agent，请输出结构化摘要。"},
			{Role: "user", Content: promptFor(item)},
		},
	})
	if err != nil {
		return item, err
	}

	req, err := http.NewRequest(http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return item, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return item, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return item, fmt.Errorf("AI 请求失败，状态码：%d", resp.StatusCode)
	}

	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return item, err
	}
	if len(result.Choices) == 0 {
		return item, errors.New("AI 响应缺少 choices")
	}

	item.Content = strings.TrimSpace(result.Choices[0].Message.Content)
	item.Score = extractScore(item.Content)
	return item, nil
}

func promptFor(item model.NewsItem) string {
	return fmt.Sprintf("总结以下内容：\n标题：%s\n内容：%s\n输出格式：\n【标题】\n【发生了什么】\n【为什么重要】\n【影响】\n【评分】1-5", item.Title, item.Content)
}

func extractScore(content string) float64 {
	for _, value := range []string{"5", "4", "3", "2", "1"} {
		if strings.Contains(content, "【评分】"+value) || strings.Contains(content, "评分】"+value) {
			return float64(value[0] - '0')
		}
	}

	return 1
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}
