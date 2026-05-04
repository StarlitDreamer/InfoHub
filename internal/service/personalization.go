package service

import (
	"sort"
	"strings"
	"time"

	"InfoHub-agent/internal/model"
)

// UserPreference 描述用户感兴趣的标签、来源和关键词。
type UserPreference struct {
	Tags     []string
	Sources  []string
	Keywords []string
	Weights  PreferenceWeights
}

// PreferenceWeights 定义个性化推荐中不同偏好信号的加权系数。
type PreferenceWeights struct {
	TagMatch     float64
	SourceMatch  float64
	KeywordMatch float64
}

// IsZero 判断是否未配置任何个性化偏好。
func (p UserPreference) IsZero() bool {
	return len(p.Tags) == 0 && len(p.Sources) == 0 && len(p.Keywords) == 0
}

// FilterByPreference 按用户标签偏好过滤信息。
func FilterByPreference(items []model.NewsItem, preference UserPreference) []model.NewsItem {
	if len(preference.Tags) == 0 {
		return append([]model.NewsItem(nil), items...)
	}

	allowed := make(map[string]struct{}, len(preference.Tags))
	for _, tag := range preference.Tags {
		allowed[strings.ToLower(strings.TrimSpace(tag))] = struct{}{}
	}

	result := make([]model.NewsItem, 0, len(items))
	for _, item := range items {
		if hasAllowedTag(item.Tags, allowed) {
			result = append(result, item)
		}
	}

	return result
}

// SortByPreferenceScore 在基础决策分上叠加偏好权重排序。
func SortByPreferenceScore(items []model.NewsItem, preference UserPreference, now time.Time) []model.NewsItem {
	if preference.IsZero() {
		return SortByDecisionScore(items, now)
	}

	result := append([]model.NewsItem(nil), items...)
	sort.SliceStable(result, func(i, j int) bool {
		left := result[i]
		right := result[j]
		leftScore := decisionScore(left, now) + preferenceBoost(left, preference)
		rightScore := decisionScore(right, now) + preferenceBoost(right, preference)
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

func preferenceBoost(item model.NewsItem, preference UserPreference) float64 {
	weights := normalizePreferenceWeights(preference.Weights)
	boost := 0.0

	tagMatches := countTagMatches(item.Tags, preference.Tags)
	boost += float64(tagMatches) * weights.TagMatch

	if matchesSource(item, preference.Sources) {
		boost += weights.SourceMatch
	}

	keywordMatches := countKeywordMatches(item, preference.Keywords)
	boost += float64(keywordMatches) * weights.KeywordMatch

	return boost
}

// normalizePreferenceWeights 为未配置的权重补齐默认值。
func normalizePreferenceWeights(weights PreferenceWeights) PreferenceWeights {
	if weights.TagMatch <= 0 {
		weights.TagMatch = 1.2
	}
	if weights.SourceMatch <= 0 {
		weights.SourceMatch = 1.0
	}
	if weights.KeywordMatch <= 0 {
		weights.KeywordMatch = 0.6
	}

	return weights
}

func hasAllowedTag(tags []string, allowed map[string]struct{}) bool {
	for _, tag := range tags {
		if _, ok := allowed[strings.ToLower(strings.TrimSpace(tag))]; ok {
			return true
		}
	}

	return false
}

func countTagMatches(tags, preferred []string) int {
	allowed := make(map[string]struct{}, len(preferred))
	for _, tag := range preferred {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized != "" {
			allowed[normalized] = struct{}{}
		}
	}

	count := 0
	seen := make(map[string]struct{})
	for _, tag := range tags {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized == "" {
			continue
		}
		if _, ok := allowed[normalized]; !ok {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		count++
	}

	return count
}

func matchesSource(item model.NewsItem, preferred []string) bool {
	sourceValues := []string{item.SourceName, item.Source}
	for _, preferredSource := range preferred {
		normalizedPreferred := strings.ToLower(strings.TrimSpace(preferredSource))
		if normalizedPreferred == "" {
			continue
		}
		for _, value := range sourceValues {
			if strings.ToLower(strings.TrimSpace(value)) == normalizedPreferred {
				return true
			}
		}
	}

	return false
}

func countKeywordMatches(item model.NewsItem, keywords []string) int {
	text := strings.ToLower(item.Title + "\n" + item.Content)
	count := 0
	seen := make(map[string]struct{})
	for _, keyword := range keywords {
		normalized := strings.ToLower(strings.TrimSpace(keyword))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		if strings.Contains(text, normalized) {
			seen[normalized] = struct{}{}
			count++
		}
	}

	return count
}
