package service

import (
	"sort"
	"strings"
	"time"

	"InfoHub-agent/internal/model"
)

// SortByDecisionScore 按热度、时间和 AI 评分综合排序。
func SortByDecisionScore(items []model.NewsItem, now time.Time) []model.NewsItem {
	result := append([]model.NewsItem(nil), items...)
	sort.SliceStable(result, func(i, j int) bool {
		left := result[i]
		right := result[j]
		leftScore := decisionScore(left, now)
		rightScore := decisionScore(right, now)
		if leftScore != rightScore {
			return leftScore > rightScore
		}
		if !left.PublishTime.Equal(right.PublishTime) {
			return left.PublishTime.After(right.PublishTime)
		}
		if left.Score != right.Score {
			return left.Score > right.Score
		}
		if left.Title != right.Title {
			return strings.Compare(left.Title, right.Title) < 0
		}
		return strings.Compare(left.URL, right.URL) < 0
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
