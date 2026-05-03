package processor

import (
	"html"
	"regexp"
	"strings"
)

const defaultCleanTextLimit = 2000

var (
	htmlTagPattern = regexp.MustCompile(`<[^>]+>`)
	spacePattern   = regexp.MustCompile(`\s+`)
)

// CleanText 清洗 RSS 和网页文本，适合在进入 AI 处理前使用。
func CleanText(value string, limit int) string {
	if limit <= 0 {
		limit = defaultCleanTextLimit
	}

	value = html.UnescapeString(value)
	value = strings.ReplaceAll(value, "\u00a0", " ")
	value = htmlTagPattern.ReplaceAllString(value, " ")
	value = html.UnescapeString(value)
	value = strings.ReplaceAll(value, "\u00a0", " ")
	value = spacePattern.ReplaceAllString(value, " ")
	value = strings.TrimSpace(value)

	return truncateRunes(value, limit)
}

func truncateRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}

	return string(runes[:limit])
}
