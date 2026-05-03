// Package processor 提供信息清洗、去重和聚合能力。
package processor

import "InfoHub-agent/internal/model"

// DeduplicateByTitle 按标题去重，并保留第一次出现的信息。
func DeduplicateByTitle(items []model.NewsItem) []model.NewsItem {
	seen := make(map[string]struct{}, len(items))
	result := make([]model.NewsItem, 0, len(items))

	for _, item := range items {
		if _, ok := seen[item.Title]; ok {
			continue
		}

		seen[item.Title] = struct{}{}
		result = append(result, item)
	}

	return result
}
