// Package processor 提供信息清洗、去重和聚合能力。
package processor

import (
	"slices"
	"strings"
	"time"
	"unicode"

	"InfoHub-agent/internal/model"
)

const similarEventWindow = 48 * time.Hour

// DeduplicateItems 按精确指纹去重，并对相似事件做轻量聚合。
func DeduplicateItems(items []model.NewsItem) []model.NewsItem {
	seen := make(map[string]struct{}, len(items)*3)
	result := make([]model.NewsItem, 0, len(items))

	for _, item := range items {
		keys := DedupKeys(item)
		if hasAnyDedupKey(seen, keys) {
			continue
		}

		merged := false
		for index := range result {
			if !IsSimilarEvent(result[index], item) {
				continue
			}

			result[index] = mergeSimilarItem(result[index], item)
			markDedupKeys(seen, keys)
			merged = true
			break
		}
		if merged {
			continue
		}

		markDedupKeys(seen, keys)
		result = append(result, item)
	}

	return result
}

// IsSimilarEvent 判断两条信息是否更像是同一事件的不同来源报道。
func IsSimilarEvent(left, right model.NewsItem) bool {
	if !publishTimeClose(left.PublishTime, right.PublishTime, similarEventWindow) {
		return false
	}

	titleSimilarity := titleSimilarityScore(left.Title, right.Title)
	if titleSimilarity < 0.55 {
		return false
	}

	if titleSimilarity >= 0.72 {
		return true
	}

	contentShingleSimilarity := shingleSimilarity(left.Content, right.Content)
	contentTokenSimilarity := tokenSimilarity(left.Content, right.Content)
	contentSimilarity := maxFloat(contentShingleSimilarity, contentTokenSimilarity)

	return titleSimilarity >= 0.58 && contentSimilarity >= 0.5
}

func titleSimilarityScore(left, right string) float64 {
	leftVariants := titleVariants(left)
	rightVariants := titleVariants(right)
	best := 0.0

	for _, leftValue := range leftVariants {
		for _, rightValue := range rightVariants {
			score := maxFloat(shingleSimilarity(leftValue, rightValue), tokenSimilarity(leftValue, rightValue))
			if score > best {
				best = score
			}
		}
	}

	return best
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

func mergeSimilarItem(base, candidate model.NewsItem) model.NewsItem {
	result := base

	if len(strings.TrimSpace(candidate.Title)) > len(strings.TrimSpace(result.Title)) {
		result.Title = candidate.Title
	}
	if len(strings.TrimSpace(candidate.Content)) > len(strings.TrimSpace(result.Content)) {
		result.Content = candidate.Content
	}
	if result.Source == "" {
		result.Source = candidate.Source
	}
	if result.SourceName == "" {
		result.SourceName = candidate.SourceName
	}
	if result.URL == "" {
		result.URL = candidate.URL
	}
	if result.PublishTime.IsZero() || (!candidate.PublishTime.IsZero() && candidate.PublishTime.After(result.PublishTime)) {
		result.PublishTime = candidate.PublishTime
	}
	if candidate.Score > result.Score {
		result.Score = candidate.Score
	}
	result.Tags = mergeTags(result.Tags, candidate.Tags)

	return result
}

func mergeTags(left, right []string) []string {
	result := make([]string, 0, len(left)+len(right))
	seen := make(map[string]struct{}, len(left)+len(right))

	for _, tag := range append(append([]string(nil), left...), right...) {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		key := strings.ToLower(tag)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, tag)
	}

	return result
}

func publishTimeClose(left, right time.Time, window time.Duration) bool {
	if left.IsZero() || right.IsZero() {
		return true
	}

	diff := left.Sub(right)
	if diff < 0 {
		diff = -diff
	}

	return diff <= window
}

func shingleSimilarity(left, right string) float64 {
	leftSet := makeShingles(left)
	rightSet := makeShingles(right)
	return setSimilarity(leftSet, rightSet)
}

func tokenSimilarity(left, right string) float64 {
	leftSet := makeTokenSet(left)
	rightSet := makeTokenSet(right)
	return setSimilarity(leftSet, rightSet)
}

func setSimilarity(leftSet, rightSet map[string]struct{}) float64 {
	if len(leftSet) == 0 || len(rightSet) == 0 {
		return 0
	}

	intersection := 0
	for value := range leftSet {
		if _, ok := rightSet[value]; ok {
			intersection++
		}
	}

	union := len(leftSet) + len(rightSet) - intersection
	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

func makeShingles(value string) map[string]struct{} {
	normalized := normalizeSimilarityText(value)
	if normalized == "" {
		return nil
	}

	tokens := strings.Fields(normalized)
	if len(tokens) > 1 {
		return tokenShingles(tokens)
	}

	runes := []rune(normalized)
	if len(runes) == 1 {
		return map[string]struct{}{normalized: {}}
	}

	result := make(map[string]struct{}, len(runes)-1)
	for index := 0; index < len(runes)-1; index++ {
		result[string(runes[index:index+2])] = struct{}{}
	}

	return result
}

func makeTokenSet(value string) map[string]struct{} {
	normalized := normalizeSimilarityText(value)
	if normalized == "" {
		return nil
	}

	tokens := strings.Fields(normalized)
	if len(tokens) > 1 {
		result := make(map[string]struct{}, len(tokens))
		for _, token := range tokens {
			if token == "" {
				continue
			}
			result[token] = struct{}{}
		}
		return result
	}

	runes := []rune(normalized)
	if len(runes) == 0 {
		return nil
	}

	result := make(map[string]struct{}, len(runes))
	for _, r := range runes {
		result[string(r)] = struct{}{}
	}

	return result
}

func tokenShingles(tokens []string) map[string]struct{} {
	result := make(map[string]struct{}, len(tokens)*2)
	compact := slices.DeleteFunc(append([]string(nil), tokens...), func(token string) bool {
		return token == ""
	})
	for _, token := range compact {
		result[token] = struct{}{}
	}
	for index := 0; index < len(compact)-1; index++ {
		result[compact[index]+" "+compact[index+1]] = struct{}{}
	}

	return result
}

func normalizeSimilarityText(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(value))
	lastSpace := false
	for _, r := range value {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Han, r):
			builder.WriteRune(r)
			lastSpace = false
		case unicode.IsSpace(r):
			if !lastSpace {
				builder.WriteByte(' ')
				lastSpace = true
			}
		default:
			if !lastSpace {
				builder.WriteByte(' ')
				lastSpace = true
			}
		}
	}

	return strings.TrimSpace(builder.String())
}

func titleVariants(value string) []string {
	candidates := []string{value}
	for _, separator := range []string{" - ", " | ", " -- ", " — ", ": "} {
		parts := strings.Split(value, separator)
		if len(parts) != 2 {
			continue
		}
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" || len([]rune(part)) < 8 {
				continue
			}
			candidates = append(candidates, part)
		}
	}

	result := make([]string, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		normalized := strings.TrimSpace(candidate)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}

	return result
}

func maxFloat(left, right float64) float64 {
	if left > right {
		return left
	}
	return right
}
