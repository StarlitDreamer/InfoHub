// Package processor 提供信息清洗、去重和聚合能力。
package processor

import "InfoHub-agent/internal/model"

// DeduplicateItems 按标题、URL 和正文指纹组合去重，并保留第一次出现的信息。
func DeduplicateItems(items []model.NewsItem) []model.NewsItem {
	seen := make(map[string]struct{}, len(items)*3)
	result := make([]model.NewsItem, 0, len(items))

	for _, item := range items {
		keys := DedupKeys(item)
		if hasAnyDedupKey(seen, keys) {
			continue
		}

		markDedupKeys(seen, keys)
		result = append(result, item)
	}

	return result
}

func hasAnyDedupKey(seen map[string]struct{}, keys []string) bool {
	for _, key := range keys {
		if _, ok := seen[key]; ok {
			return true
		}
	}

	return false
}

func markDedupKeys(seen map[string]struct{}, keys []string) {
	for _, key := range keys {
		seen[key] = struct{}{}
	}
}
