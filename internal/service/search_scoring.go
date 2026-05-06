package service

import (
	"sort"
	"strings"
	"time"

	"InfoHub-agent/internal/model"
)

// SortSearchItems 按搜索场景的综合得分排序。
func SortSearchItems(items []model.NewsItem, preference UserPreference, now time.Time) []model.NewsItem {
	result := append([]model.NewsItem(nil), items...)
	sort.SliceStable(result, func(i, j int) bool {
		left := searchScore(result[i], preference, now)
		right := searchScore(result[j], preference, now)
		if left != right {
			return left > right
		}
		if !result[i].PublishTime.Equal(result[j].PublishTime) {
			return result[i].PublishTime.After(result[j].PublishTime)
		}
		return strings.Compare(result[i].Title, result[j].Title) < 0
	})

	return result
}

func searchScore(item model.NewsItem, preference UserPreference, now time.Time) float64 {
	score := decisionScore(item, now)
	score += preferenceBoost(item, preference)
	score += sourceBoost(item)
	score += item.SourceScore * 0.05
	score += float64(countKeywordMatches(item, []string{item.Query})) * 0.8
	return score
}

func sourceBoost(item model.NewsItem) float64 {
	switch strings.ToLower(strings.TrimSpace(item.Channel)) {
	case "stack_overflow":
		return 0.6
	case "reddit":
		return 0.4
	case "rss_search":
		return 0.2
	default:
		return 0
	}
}
