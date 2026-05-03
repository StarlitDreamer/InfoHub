// Package delivery 提供日报输出和推送能力。
package delivery

import (
	"fmt"
	"sort"
	"strings"

	"InfoHub-agent/internal/model"
)

// RenderMarkdown 将信息列表渲染为 Markdown 日报。
func RenderMarkdown(items []model.NewsItem) string {
	return renderMarkdown(items, false)
}

// RenderMarkdownBySource 按来源分组渲染 Markdown 日报。
func RenderMarkdownBySource(items []model.NewsItem) string {
	return renderMarkdown(items, true)
}

func renderMarkdown(items []model.NewsItem, groupBySource bool) string {
	var builder strings.Builder
	builder.WriteString("# 今日信息\n\n")

	if len(items) == 0 {
		builder.WriteString("今日暂无新增信息。\n")
		return builder.String()
	}

	if groupBySource {
		renderGroupedItems(&builder, items)
		return builder.String()
	}

	for _, item := range items {
		builder.WriteString(fmt.Sprintf("## %s\n", scoreStars(item.Score)))
		builder.WriteString(fmt.Sprintf("- 标题：%s\n", item.Title))
		builder.WriteString(fmt.Sprintf("- 摘要：%s\n\n", item.Content))
	}

	return builder.String()
}

func renderGroupedItems(builder *strings.Builder, items []model.NewsItem) {
	grouped := make(map[string][]model.NewsItem)
	order := make([]string, 0)

	for _, item := range items {
		source := strings.TrimSpace(item.Source)
		if source == "" {
			source = "Unknown Source"
		}
		if _, ok := grouped[source]; !ok {
			order = append(order, source)
		}
		grouped[source] = append(grouped[source], item)
	}

	sort.Strings(order)
	for _, source := range order {
		builder.WriteString(fmt.Sprintf("### 来源：%s\n\n", source))
		for _, item := range grouped[source] {
			builder.WriteString(fmt.Sprintf("## %s\n", scoreStars(item.Score)))
			builder.WriteString(fmt.Sprintf("- 标题：%s\n", item.Title))
			builder.WriteString(fmt.Sprintf("- 摘要：%s\n\n", item.Content))
		}
	}
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
