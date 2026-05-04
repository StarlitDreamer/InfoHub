// Package summary 提供 AI 结构化摘要解析能力。
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
