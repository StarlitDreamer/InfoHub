package ai

import (
	"fmt"
	"strings"

	"InfoHub-agent/internal/model"
)

// MockProcessor 提供无需外部依赖的 AI 模拟处理能力。
type MockProcessor struct{}

// NewMockProcessor 创建模拟 AI 处理器。
func NewMockProcessor() *MockProcessor {
	return &MockProcessor{}
}

// Analyze 返回分类、摘要和评分结果。
func (p *MockProcessor) Analyze(item model.NewsItem) (Analysis, error) {
	score := float64((len(item.Title) % 5) + 1)
	tags := mockTags(item)

	return Analysis{
		Tags:    tags,
		Summary: mockSummary(item, score),
		Score:   score,
	}, nil
}

// Classify 返回模拟标签。
func (p *MockProcessor) Classify(item model.NewsItem) ([]string, error) {
	analysis, err := p.Analyze(item)
	if err != nil {
		return nil, err
	}

	return analysis.Tags, nil
}

// Summarize 返回模拟摘要。
func (p *MockProcessor) Summarize(item model.NewsItem) (string, error) {
	analysis, err := p.Analyze(item)
	if err != nil {
		return "", err
	}

	return analysis.Summary, nil
}

// Score 返回模拟评分。
func (p *MockProcessor) Score(item model.NewsItem) (float64, error) {
	analysis, err := p.Analyze(item)
	if err != nil {
		return 0, err
	}

	return analysis.Score, nil
}

func mockTags(item model.NewsItem) []string {
	if len(item.Tags) > 0 {
		return append([]string(nil), item.Tags...)
	}

	text := strings.ToLower(item.Title + " " + item.Content)
	tags := make([]string, 0, 2)
	if strings.Contains(text, "ai") || strings.Contains(text, "model") || strings.Contains(text, "openai") {
		tags = append(tags, "AI")
	}
	if strings.Contains(text, "cloud") || strings.Contains(text, "deploy") {
		tags = append(tags, "Cloud")
	}
	if strings.Contains(text, "database") || strings.Contains(text, "index") {
		tags = append(tags, "Database")
	}
	if len(tags) == 0 {
		tags = append(tags, "General")
	}

	return tags
}

func mockSummary(item model.NewsItem, score float64) string {
	return fmt.Sprintf(
		"【标题】%s\n【发生了什么】%s\n【为什么重要】该信息可能影响后续技术选型或行动判断。\n【影响】建议持续关注相关进展。\n【评分】%.0f",
		item.Title,
		item.Content,
		score,
	)
}
