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
	builder.WriteString(fmt.Sprintf("- 重点关注：%s\n\n", summarizeTopTitles(items, 3)))
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
	summary := parseStructuredSummary(item)

	builder.WriteString(fmt.Sprintf("## %s %s\n", scoreStars(item.Score), summary.Title))
	builder.WriteString(fmt.Sprintf("- 标题：%s\n", summary.Title))
	builder.WriteString(fmt.Sprintf("- 来源：%s\n", normalizeSource(item.Source)))
	if !item.PublishTime.IsZero() {
		builder.WriteString(fmt.Sprintf("- 时间：%s\n", item.PublishTime.Format("2006-01-02 15:04")))
	}
	if len(item.Tags) > 0 {
		builder.WriteString(fmt.Sprintf("- 标签：%s\n", strings.Join(item.Tags, "、")))
	}
	builder.WriteString(fmt.Sprintf("- 发生了什么：%s\n", summary.WhatHappened))
	builder.WriteString(fmt.Sprintf("- 为什么重要：%s\n", summary.WhyImportant))
	builder.WriteString(fmt.Sprintf("- 影响：%s\n", summary.Impact))
	builder.WriteString(fmt.Sprintf("- 建议动作：%s\n", recommendAction(item, summary)))
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

	for _, rawLine := range strings.Split(item.Content, "\n") {
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

func recommendAction(item model.NewsItem, summary structuredSummary) string {
	score := clampScore(item.Score)
	text := strings.ToLower(strings.Join(item.Tags, " ") + " " + item.Title + " " + summary.WhyImportant + " " + summary.Impact)

	switch {
	case score >= 5:
		return "优先安排评审，判断是否需要立即纳入本周行动或技术路线。"
	case score >= 4:
		return "加入近期待办，指定负责人跟进原文和后续进展。"
	case strings.Contains(text, "security") || strings.Contains(text, "安全") || strings.Contains(text, "cyber"):
		return "转给安全相关负责人评估影响面，并确认是否需要额外检查。"
	case strings.Contains(text, "database") || strings.Contains(text, "数据库") || strings.Contains(text, "index"):
		return "结合当前系统瓶颈评估可借鉴点，必要时安排一次专项验证。"
	case strings.Contains(text, "ai") || strings.Contains(text, "agent") || strings.Contains(text, "模型"):
		return "记录到 AI 能力跟踪清单，评估是否值得做小范围试用。"
	default:
		return "先纳入观察列表，等待更多上下文后再决定是否升级处理。"
	}
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
