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
	tags := mockTags(item)
	score := mockScore(item, tags)

	return Analysis{
		Tags:    tags,
		Summary: mockSummary(item, tags, score),
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

	text := strings.ToLower(item.Title + " " + item.Content + " " + item.Source)
	tags := make([]string, 0, 3)
	if containsAny(text, "ai", "model", "openai", "gemini", "agent", "codex", "gpt") {
		tags = append(tags, "AI")
	}
	if containsAny(text, "security", "cyber", "phishing", "safeguard", "account takeover", "fedramp") {
		tags = append(tags, "Security")
	}
	if containsAny(text, "cloud", "aws", "api", "deploy", "infrastructure", "compute", "enterprise") {
		tags = append(tags, "Cloud")
	}
	if containsAny(text, "database", "index", "storage", "mysql", "redis") {
		tags = append(tags, "Database")
	}
	if len(tags) == 0 {
		tags = append(tags, "General")
	}

	return tags
}

func mockScore(item model.NewsItem, tags []string) float64 {
	text := strings.ToLower(item.Title + " " + item.Content + " " + item.Source)
	score := 2.0

	if containsTag(tags, "Security") {
		score += 1.4
	}
	if containsTag(tags, "AI") {
		score += 0.9
	}
	if containsTag(tags, "Cloud") || containsTag(tags, "Database") {
		score += 0.7
	}

	if containsAny(text, "launch", "release", "available", "introducing", "new feature", "now available") {
		score += 0.4
	}
	if containsAny(text, "enterprise", "api", "aws", "infrastructure", "compute", "security", "cyber", "account") {
		score += 0.5
	}
	if containsAny(text, "shopping", "ads", "campaign", "marketing", "photos", "wardrobe", "route 66", "translate", "time 100", "anniversary", "travelers") {
		score -= 0.8
	}
	if containsAny(text, "earnings", "remarks from our ceo", "cover", "fun facts") {
		score -= 0.5
	}

	return clampMockScore(score)
}

func mockSummary(item model.NewsItem, tags []string, score float64) string {
	what := strings.TrimSpace(item.Content)
	if what == "" {
		what = item.Title
	}
	what = truncateText(what, 180)

	whyImportant, impact := mockDecisionContext(item, tags)
	return fmt.Sprintf(
		"【标题】%s\n【发生了什么】%s\n【为什么重要】%s\n【影响】%s\n【评分】%.0f",
		item.Title,
		what,
		whyImportant,
		impact,
		score,
	)
}

func mockDecisionContext(item model.NewsItem, tags []string) (string, string) {
	text := strings.ToLower(item.Title + " " + item.Content + " " + item.Source)

	switch {
	case containsTag(tags, "Security"):
		return "这类安全与合规更新可能影响账号保护、风险控制或对外合规判断。", "建议评估是否需要纳入安全基线、访问控制或合规跟进事项。"
	case containsAny(text, "api", "aws", "enterprise", "compute", "infrastructure", "cloud"):
		return "这类平台能力更新可能影响技术选型、集成路径或基础设施规划。", "建议确认是否需要技术验证、供应商评估或架构预研。"
	case containsTag(tags, "AI"):
		return "这类 AI 能力更新可能影响产品路线、能力接入和近期试点优先级。", "建议结合现有需求评估是否进入小范围试用或能力跟踪清单。"
	default:
		return "该信息可作为行业动态参考，但对主线决策的直接影响相对有限。", "建议纳入观察列表，等待更多上下文后再决定是否升级处理。"
	}
}

func containsTag(tags []string, target string) bool {
	for _, tag := range tags {
		if strings.EqualFold(strings.TrimSpace(tag), target) {
			return true
		}
	}

	return false
}

func containsAny(text string, values ...string) bool {
	for _, value := range values {
		if strings.Contains(text, value) {
			return true
		}
	}

	return false
}

func truncateText(value string, limit int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= limit {
		return string(runes)
	}

	return string(runes[:limit]) + "..."
}

func clampMockScore(score float64) float64 {
	if score < 1 {
		return 1
	}
	if score > 5 {
		return 5
	}

	return score
}
