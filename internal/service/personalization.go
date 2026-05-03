package service

import "InfoHub-agent/internal/model"

// UserPreference 描述用户感兴趣的标签。
type UserPreference struct {
	Tags []string
}

// FilterByPreference 按用户标签偏好过滤信息。
func FilterByPreference(items []model.NewsItem, preference UserPreference) []model.NewsItem {
	if len(preference.Tags) == 0 {
		return items
	}

	allowed := make(map[string]struct{}, len(preference.Tags))
	for _, tag := range preference.Tags {
		allowed[tag] = struct{}{}
	}

	result := make([]model.NewsItem, 0, len(items))
	for _, item := range items {
		if hasAllowedTag(item.Tags, allowed) {
			result = append(result, item)
		}
	}

	return result
}

func hasAllowedTag(tags []string, allowed map[string]struct{}) bool {
	for _, tag := range tags {
		if _, ok := allowed[tag]; ok {
			return true
		}
	}

	return false
}
