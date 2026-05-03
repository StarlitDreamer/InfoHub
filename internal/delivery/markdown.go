// Package delivery 提供日报输出和推送能力。
package delivery

import (
	"fmt"
	"strings"

	"InfoHub-agent/internal/model"
)

// RenderMarkdown 将信息列表渲染为 Markdown 日报。
func RenderMarkdown(items []model.NewsItem) string {
	var builder strings.Builder
	builder.WriteString("# 今日信息\n\n")

	if len(items) == 0 {
		builder.WriteString("今日暂无新增信息。\n")
		return builder.String()
	}

	for _, item := range items {
		builder.WriteString(fmt.Sprintf("## %s\n", scoreStars(item.Score)))
		builder.WriteString(fmt.Sprintf("- 标题：%s\n", item.Title))
		builder.WriteString(fmt.Sprintf("- 摘要：%s\n\n", item.Content))
	}

	return builder.String()
}

// scoreStars 将 1-5 分评分转换为星级展示。
func scoreStars(score float64) string {
	count := int(score)
	if count < 1 {
		count = 1
	}

	if count > 5 {
		count = 5
	}

	return strings.Repeat("⭐", count)
}
