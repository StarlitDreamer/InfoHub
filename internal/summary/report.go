// Package summary 提供 AI 结构化摘要与日报概览能力。
package summary

import (
	"fmt"
	"sort"
	"strings"

	"InfoHub-agent/internal/model"
)

// Overview 表示日报顶部概览信息。
type Overview struct {
	ItemCount         int
	HighPriorityCount int
	SourceSummary     string
	TopTitles         []string
	PriorityActions   []string
}

// BuildOverview 构建日报概览。
func BuildOverview(items []model.NewsItem, titleLimit, actionLimit int) Overview {
	return Overview{
		ItemCount:         len(items),
		HighPriorityCount: countHighPriority(items),
		SourceSummary:     summarizeSources(items),
		TopTitles:         topTitles(items, titleLimit),
		PriorityActions:   summarizePriorityActions(items, actionLimit),
	}
}

func summarizePriorityActions(items []model.NewsItem, limit int) []string {
	if limit <= 0 {
		limit = 1
	}

	actions := make([]string, 0, limit)
	seen := make(map[string]struct{})
	for _, item := range items {
		action := RecommendAction(item, Parse(item)).Description
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

	if len(parts) == 0 {
		return "暂无"
	}

	return strings.Join(parts, "；")
}

func topTitles(items []model.NewsItem, limit int) []string {
	if limit <= 0 {
		limit = 1
	}

	result := make([]string, 0, limit)
	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		if title == "" {
			continue
		}
		result = append(result, title)
		if len(result) == limit {
			break
		}
	}

	return result
}

func normalizeSource(source string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return "未知来源"
	}

	return source
}
