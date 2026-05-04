// Package delivery 提供日报输出和推送能力。
package delivery

import (
	"fmt"
	"sort"
	"strings"

	"InfoHub-agent/internal/model"
	"InfoHub-agent/internal/summary"
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
	builder.WriteString("# 今日信息日报\n\n")

	if len(items) == 0 {
		builder.WriteString("今日暂无新增信息。\n")
		return builder.String()
	}

	renderOverview(&builder, items)

	if groupBySource {
		renderGroupedItems(&builder, items)
		return builder.String()
	}

	for _, item := range items {
		renderItem(&builder, item)
	}

	return builder.String()
}

func renderOverview(builder *strings.Builder, items []model.NewsItem) {
	builder.WriteString("## 今日概览\n")
	builder.WriteString(fmt.Sprintf("- 收录条目：%d\n", len(items)))
	builder.WriteString(fmt.Sprintf("- 高优先级：%d\n", countHighPriority(items)))
	builder.WriteString(fmt.Sprintf("- 来源分布：%s\n", summarizeSources(items)))
	builder.WriteString(fmt.Sprintf("- 重点关注：%s\n", summarizeTopTitles(items, 3)))

	actions := summarizePriorityActions(items, 3)
	if len(actions) > 0 {
		builder.WriteString("- 今日优先动作：\n")
		for _, action := range actions {
			builder.WriteString(fmt.Sprintf("  - %s\n", action))
		}
	}

	builder.WriteString("\n")
}

func renderGroupedItems(builder *strings.Builder, items []model.NewsItem) {
	grouped := make(map[string][]model.NewsItem)
	order := make([]string, 0)

	for _, item := range items {
		source := normalizeSource(item.Source)
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
	parsed := summary.Parse(item)
	action := summary.RecommendAction(item, parsed)

	builder.WriteString(fmt.Sprintf("## %s %s\n", scoreStars(item.Score), parsed.Title))
	builder.WriteString(fmt.Sprintf("- 标题：%s\n", parsed.Title))
	builder.WriteString(fmt.Sprintf("- 来源：%s\n", normalizeSource(item.Source)))
	if !item.PublishTime.IsZero() {
		builder.WriteString(fmt.Sprintf("- 时间：%s\n", item.PublishTime.Format("2006-01-02 15:04")))
	}
	if len(item.Tags) > 0 {
		builder.WriteString(fmt.Sprintf("- 标签：%s\n", strings.Join(item.Tags, "、")))
	}
	builder.WriteString(fmt.Sprintf("- 发生了什么：%s\n", parsed.WhatHappened))
	builder.WriteString(fmt.Sprintf("- 为什么重要：%s\n", parsed.WhyImportant))
	builder.WriteString(fmt.Sprintf("- 影响：%s\n", parsed.Impact))
	builder.WriteString(fmt.Sprintf("- 建议动作：%s\n", action.Description))
	builder.WriteString(fmt.Sprintf("- 评分：%.0f/5\n", clampScore(item.Score)))
	if item.URL != "" {
		builder.WriteString(fmt.Sprintf("- 原文链接：%s\n", item.URL))
	}
	builder.WriteString("\n")
}

func summarizePriorityActions(items []model.NewsItem, limit int) []string {
	if limit <= 0 {
		limit = 1
	}

	actions := make([]string, 0, limit)
	seen := make(map[string]struct{})
	for _, item := range items {
		action := summary.RecommendAction(item, summary.Parse(item)).Description
		if _, ok := seen[action]; ok {
			continue
		}
		seen[action] = struct{}{}
		actions = append(actions, action)
		if len(actions) == limit {
			break
		}
	}

	return actions
}

func countHighPriority(items []model.NewsItem) int {
	count := 0
	for _, item := range items {
		if clampScore(item.Score) >= 4 {
			count++
		}
	}

	return count
}

func summarizeSources(items []model.NewsItem) string {
	counts := map[string]int{}
	order := make([]string, 0)

	for _, item := range items {
		source := normalizeSource(item.Source)
		if _, ok := counts[source]; !ok {
			order = append(order, source)
		}
		counts[source]++
	}

	sort.Strings(order)
	parts := make([]string, 0, len(order))
	for _, source := range order {
		parts = append(parts, fmt.Sprintf("%s %d", source, counts[source]))
	}

	return strings.Join(parts, "；")
}

func summarizeTopTitles(items []model.NewsItem, limit int) string {
	if limit <= 0 {
		limit = 1
	}

	parts := make([]string, 0, limit)
	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		if title == "" {
			continue
		}
		parts = append(parts, title)
		if len(parts) == limit {
			break
		}
	}

	if len(parts) == 0 {
		return "暂无"
	}

	return strings.Join(parts, "；")
}

func normalizeSource(source string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return "未知来源"
	}

	return source
}

// scoreStars 将 1-5 分评分转换为星级展示。
func scoreStars(score float64) string {
	return strings.Repeat("⭐", int(clampScore(score)))
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
