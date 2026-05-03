package ai

import (
	"fmt"

	"InfoHub-agent/internal/model"
)

// MockProcessor 提供无需外部依赖的 AI 模拟处理能力。
type MockProcessor struct{}

// NewMockProcessor 创建模拟 AI 处理器。
func NewMockProcessor() *MockProcessor {
	return &MockProcessor{}
}

// Summarize 生成结构化摘要并补充重要性评分。
func (p *MockProcessor) Summarize(item model.NewsItem) (model.NewsItem, error) {
	score := float64((len(item.Title) % 5) + 1)
	item.Score = score
	item.Content = fmt.Sprintf("【标题】%s\n【发生了什么】%s\n【为什么重要】该信息可能影响后续技术选型或行动判断。\n【影响】建议持续关注相关进展。\n【评分】%.0f", item.Title, item.Content, score)

	return item, nil
}
