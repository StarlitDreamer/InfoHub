package delivery

import (
	"fmt"
	"strings"

	"InfoHub-agent/internal/model"
	"InfoHub-agent/internal/summary"
)

// RenderMarkdown 将信息列表渲染为 Markdown 日报。
func RenderMarkdown(items []model.NewsItem) string {
	return renderMarkdown(items, nil, false)
}

// RenderMarkdownWithWarnings 渲染带抓取告警的 Markdown 日报。
func RenderMarkdownWithWarnings(items []model.NewsItem, warnings []string) string {
	return renderMarkdown(items, warnings, false)
}

// RenderMarkdownBySource 按来源分组渲染 Markdown 日报。
func RenderMarkdownBySource(items []model.NewsItem) string {
	return renderMarkdown(items, nil, true)
}

// RenderMarkdownBySourceWithWarnings 渲染按来源分组且带抓取告警的日报。
func RenderMarkdownBySourceWithWarnings(items []model.NewsItem, warnings []string) string {
	return renderMarkdown(items, warnings, true)
}

func renderMarkdown(items []model.NewsItem, warnings []string, groupBySource bool) string {
	var builder strings.Builder
	builder.WriteString("# 今日信息日报\n\n")

	if len(items) == 0 {
		builder.WriteString("今日暂无新增信息。\n")
		return builder.String()
	}

	renderOverview(&builder, items, warnings)

	if groupBySource {
		renderGroupedItems(&builder, items)
		return builder.String()
	}

	for _, item := range items {
		renderItem(&builder, item)
	}

	return builder.String()
}

func renderOverview(builder *strings.Builder, items []model.NewsItem, warnings []string) {
	overview := summary.BuildOverview(items, 3, 3)

	builder.WriteString("## 今日概览\n")
	builder.WriteString(fmt.Sprintf("- 收录条目：%d\n", overview.ItemCount))
	builder.WriteString(fmt.Sprintf("- 高优先级：%d\n", overview.HighPriorityCount))
	builder.WriteString(fmt.Sprintf("- 来源分布：%s\n", overview.SourceSummary))
	builder.WriteString(fmt.Sprintf("- 重点关注：%s\n", joinOrDefault(overview.TopTitles, "；", "暂无")))

	if len(overview.PriorityActions) > 0 {
		builder.WriteString("- 今日优先动作：\n")
		for _, action := range overview.PriorityActions {
			builder.WriteString(fmt.Sprintf("  - %s\n", action))
		}
	}
	if len(warnings) > 0 {
		builder.WriteString(fmt.Sprintf("- 抓取异常：%s\n", strings.Join(warnings, "；")))
	}

	builder.WriteString("\n")
}

func renderGroupedItems(builder *strings.Builder, items []model.NewsItem) {
	for _, group := range summary.GroupBySource(items) {
		builder.WriteString(fmt.Sprintf("### 来源：%s\n\n", group.Source))
		for _, item := range group.Items {
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

func normalizeSource(source string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return "未知来源"
	}

	return source
}

func joinOrDefault(values []string, sep, fallback string) string {
	if len(values) == 0 {
		return fallback
	}
	return strings.Join(values, sep)
}

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
