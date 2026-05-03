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

// HTTPClient 调用 OpenAI 兼容接口生成真实分析结果。
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

// Analyze 调用真实模型并返回分类、摘要和评分结果。
func (c *HTTPClient) Analyze(item model.NewsItem) (Analysis, error) {
	if c.endpoint == "" || c.apiKey == "" || c.model == "" {
		return Analysis{}, errors.New("ai client requires endpoint, api key, and model")
	}

	body, err := json.Marshal(chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: promptFor(item)},
		},
		ResponseFormat: &responseFormat{
			Type: "json_object",
		},
	})
	if err != nil {
		return Analysis{}, err
	}

	req, err := http.NewRequest(http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return Analysis{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return Analysis{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return Analysis{}, fmt.Errorf("ai request failed with status %d", resp.StatusCode)
	}

	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Analysis{}, err
	}
	if len(result.Choices) == 0 {
		return Analysis{}, errors.New("ai response has no choices")
	}

	return parseAnalysis(result.Choices[0].Message.Content)
}

// Classify 返回标签。
func (c *HTTPClient) Classify(item model.NewsItem) ([]string, error) {
	analysis, err := c.Analyze(item)
	if err != nil {
		return nil, err
	}

	return analysis.Tags, nil
}

// Summarize 返回结构化摘要。
func (c *HTTPClient) Summarize(item model.NewsItem) (string, error) {
	analysis, err := c.Analyze(item)
	if err != nil {
		return "", err
	}

	return analysis.Summary, nil
}

// Score 返回评分。
func (c *HTTPClient) Score(item model.NewsItem) (float64, error) {
	analysis, err := c.Analyze(item)
	if err != nil {
		return 0, err
	}

	return analysis.Score, nil
}

const systemPrompt = "You are an information aggregation agent. Return JSON with keys tags, summary, and score. Summary must use the required Chinese labeled format."

func promptFor(item model.NewsItem) string {
	return fmt.Sprintf(
		"Analyze the following content.\nTitle: %s\nContent: %s\nReturn JSON like {\"tags\":[...],\"summary\":\"...\",\"score\":1-5}.\nThe summary text must use this format:\n【标题】\n【发生了什么】\n【为什么重要】\n【影响】\n【评分】1-5",
		item.Title,
		item.Content,
	)
}

func parseAnalysis(content string) (Analysis, error) {
	var payload analysisPayload
	if err := json.Unmarshal([]byte(strings.TrimSpace(content)), &payload); err != nil {
		return Analysis{}, err
	}

	return Analysis{
		Tags:    normalizeTags(payload.Tags),
		Summary: strings.TrimSpace(payload.Summary),
		Score:   clampScore(payload.Score),
	}, nil
}

func normalizeTags(tags []string) []string {
	result := make([]string, 0, len(tags))
	seen := map[string]struct{}{}

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		key := strings.ToLower(tag)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, tag)
	}

	return result
}

func clampScore(score float64) float64 {
	if score < 1 {
		return 1
	}
	if score > 5 {
		return 5
	}

	return score
}

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type responseFormat struct {
	Type string `json:"type"`
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

type analysisPayload struct {
	Tags    []string `json:"tags"`
	Summary string   `json:"summary"`
	Score   float64  `json:"score"`
}
