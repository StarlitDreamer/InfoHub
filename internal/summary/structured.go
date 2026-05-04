// Package summary 提供 AI 结构化摘要与动作建议能力。
package summary

import (
	"strings"

	"InfoHub-agent/internal/model"
)

// Structured 表示结构化摘要内容。
type Structured struct {
	Title        string
	WhatHappened string
	WhyImportant string
	Impact       string
}

// Action 表示统一的动作建议输出。
type Action struct {
	Label       string
	Description string
}

// Parse 解析 NewsItem 中的结构化摘要，并在缺失字段时提供回退值。
func Parse(item model.NewsItem) Structured {
	result := Structured{
		Title:        strings.TrimSpace(item.Title),
		WhatHappened: strings.TrimSpace(item.Content),
		WhyImportant: "该信息可能影响后续判断，建议结合业务上下文继续关注。",
		Impact:       "建议评估是否需要跟进、验证或纳入后续决策。",
	}

	for _, rawLine := range strings.Split(item.Content, "\n") {
		line := strings.TrimSpace(rawLine)
		switch {
		case strings.HasPrefix(line, "【标题】"):
			result.Title = strings.TrimSpace(strings.TrimPrefix(line, "【标题】"))
		case strings.HasPrefix(line, "【发生了什么】"):
			result.WhatHappened = strings.TrimSpace(strings.TrimPrefix(line, "【发生了什么】"))
		case strings.HasPrefix(line, "【为什么重要】"):
			result.WhyImportant = strings.TrimSpace(strings.TrimPrefix(line, "【为什么重要】"))
		case strings.HasPrefix(line, "【影响】"):
			result.Impact = strings.TrimSpace(strings.TrimPrefix(line, "【影响】"))
		}
	}

	if result.Title == "" {
		result.Title = "未命名信息"
	}
	if result.WhatHappened == "" {
		result.WhatHappened = result.Title
	}

	return result
}

// RecommendAction 基于评分、标签和摘要内容生成统一动作建议。
func RecommendAction(item model.NewsItem, parsed Structured) Action {
	score := clampScore(item.Score)
	text := strings.ToLower(strings.Join(item.Tags, " ") + " " + item.Title + " " + parsed.WhyImportant + " " + parsed.Impact)

	switch {
	case score >= 5:
		return Action{
			Label:       "立即评审",
			Description: "优先安排评审，判断是否需要立即纳入本周行动或技术路线。",
		}
	case score >= 4:
		return Action{
			Label:       "近期跟进",
			Description: "加入近期待办，指定负责人跟进原文和后续进展。",
		}
	case strings.Contains(text, "security") || strings.Contains(text, "安全") || strings.Contains(text, "cyber"):
		return Action{
			Label:       "安全评估",
			Description: "转给安全相关负责人评估影响面，并确认是否需要额外检查。",
		}
	case strings.Contains(text, "database") || strings.Contains(text, "数据库") || strings.Contains(text, "index"):
		return Action{
			Label:       "专项验证",
			Description: "结合当前系统瓶颈评估可借鉴点，必要时安排一次专项验证。",
		}
	case strings.Contains(text, "ai") || strings.Contains(text, "agent") || strings.Contains(text, "模型"):
		return Action{
			Label:       "小范围试用",
			Description: "记录到 AI 能力跟踪清单，评估是否值得做小范围试用。",
		}
	default:
		return Action{
			Label:       "持续观察",
			Description: "先纳入观察列表，等待更多上下文后再决定是否升级处理。",
		}
	}
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
