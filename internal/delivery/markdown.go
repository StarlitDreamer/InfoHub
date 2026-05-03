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
		renderItem(&builder, item)
	}

	return builder.String()
}

func renderGroupedItems(builder *strings.Builder, items []model.NewsItem) {
	grouped := make(map[string][]model.NewsItem)
	order := make([]string, 0)

	for _, item := range items {
		source := strings.TrimSpace(item.Source)
		if source == "" {
			source = "未知来源"
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
			renderItem(builder, item)
		}
	}
}

func renderItem(builder *strings.Builder, item model.NewsItem) {
	summary := parseStructuredSummary(item)

	builder.WriteString(fmt.Sprintf("## %s %s\n", scoreStars(item.Score), summary.Title))
	builder.WriteString(fmt.Sprintf("- 标题：%s\n", summary.Title))
	if item.Source != "" {
		builder.WriteString(fmt.Sprintf("- 来源：%s\n", item.Source))
	}
	if !item.PublishTime.IsZero() {
		builder.WriteString(fmt.Sprintf("- 时间：%s\n", item.PublishTime.Format("2006-01-02 15:04")))
	}
	if len(item.Tags) > 0 {
		builder.WriteString(fmt.Sprintf("- 标签：%s\n", strings.Join(item.Tags, "、")))
	}
	builder.WriteString(fmt.Sprintf("- 发生了什么：%s\n", summary.WhatHappened))
	builder.WriteString(fmt.Sprintf("- 为什么重要：%s\n", summary.WhyImportant))
	builder.WriteString(fmt.Sprintf("- 影响：%s\n", summary.Impact))
	builder.WriteString(fmt.Sprintf("- 评分：%.0f/5\n", clampScore(item.Score)))
	if item.URL != "" {
		builder.WriteString(fmt.Sprintf("- 原文链接：%s\n", item.URL))
	}
	builder.WriteString("\n")
}

type structuredSummary struct {
	Title        string
	WhatHappened string
	WhyImportant string
	Impact       string
}

func parseStructuredSummary(item model.NewsItem) structuredSummary {
	summary := structuredSummary{
		Title:        strings.TrimSpace(item.Title),
		WhatHappened: strings.TrimSpace(item.Content),
		WhyImportant: "该信息可能影响后续判断，建议结合业务上下文继续关注。",
		Impact:       "建议评估是否需要跟进、验证或纳入后续决策。",
	}

	lines := strings.Split(item.Content, "\n")
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		switch {
		case strings.HasPrefix(line, "【标题】"):
			summary.Title = strings.TrimSpace(strings.TrimPrefix(line, "【标题】"))
		case strings.HasPrefix(line, "【发生了什么】"):
			summary.WhatHappened = strings.TrimSpace(strings.TrimPrefix(line, "【发生了什么】"))
		case strings.HasPrefix(line, "【为什么重要】"):
			summary.WhyImportant = strings.TrimSpace(strings.TrimPrefix(line, "【为什么重要】"))
		case strings.HasPrefix(line, "【影响】"):
			summary.Impact = strings.TrimSpace(strings.TrimPrefix(line, "【影响】"))
		}
	}

	if summary.Title == "" {
		summary.Title = "未命名信息"
	}
	if summary.WhatHappened == "" {
		summary.WhatHappened = summary.Title
	}

	return summary
}

// scoreStars 将 1-5 分评分转换为星级展示。
func scoreStars(score float64) string {
	count := int(clampScore(score))
	return strings.Repeat("⭐", count)
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
