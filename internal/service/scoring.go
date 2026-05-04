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

// LimitItemsBalancedBySource 在保留原有优先级顺序的前提下，尽量让展示结果覆盖更多来源。
func LimitItemsBalancedBySource(items []model.NewsItem, n int) []model.NewsItem {
	if n <= 0 || len(items) <= n {
		return append([]model.NewsItem(nil), items...)
	}

	grouped := make(map[string][]model.NewsItem)
	order := make([]string, 0, len(items))
	for _, item := range items {
		key := displaySourceKey(item)
		if _, ok := grouped[key]; !ok {
			order = append(order, key)
		}
		grouped[key] = append(grouped[key], item)
	}

	result := make([]model.NewsItem, 0, minInt(n, len(items)))
	for len(result) < n {
		picked := false
		for _, key := range order {
			queue := grouped[key]
			if len(queue) == 0 {
				continue
			}

			result = append(result, queue[0])
			grouped[key] = queue[1:]
			picked = true
			if len(result) == n {
				break
			}
		}

		if !picked {
			break
		}
	}

	return result
}

func decisionScore(item model.NewsItem, now time.Time) float64 {
	freshness := 1 / (1 + now.Sub(item.PublishTime).Hours())
	tagHeat := float64(len(item.Tags)) * 0.2
	return item.Score + freshness + tagHeat
}

func displaySourceKey(item model.NewsItem) string {
	if value := strings.TrimSpace(item.SourceName); value != "" {
		return value
	}
	if value := strings.TrimSpace(item.Source); value != "" {
		return value
	}

	return "__unknown__"
}

func minInt(left, right int) int {
	if left < right {
		return left
	}

	return right
}
