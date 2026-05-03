package service

import (
	"sort"
	"time"

	"InfoHub-agent/internal/model"
)

// SortByDecisionScore 按热度、时间和 AI 评分综合排序。
func SortByDecisionScore(items []model.NewsItem, now time.Time) []model.NewsItem {
	result := append([]model.NewsItem(nil), items...)
	sort.SliceStable(result, func(i, j int) bool {
		return decisionScore(result[i], now) > decisionScore(result[j], now)
	})

	return result
}

// LimitItems 保留前 n 条，n 小于等于 0 时返回全部结果。
func LimitItems(items []model.NewsItem, n int) []model.NewsItem {
	if n <= 0 || len(items) <= n {
		return append([]model.NewsItem(nil), items...)
	}

	result := make([]model.NewsItem, n)
	copy(result, items[:n])
	return result
}

func decisionScore(item model.NewsItem, now time.Time) float64 {
	freshness := 1 / (1 + now.Sub(item.PublishTime).Hours())
	tagHeat := float64(len(item.Tags)) * 0.2
	return item.Score + freshness + tagHeat
}
