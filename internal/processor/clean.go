package processor

import (
	"html"
	"regexp"
	"strings"
)

const defaultCleanTextLimit = 2000

var (
	blockBoundaryPattern = regexp.MustCompile(`(?i)</?(p|div|br|li|ul|ol|section|article|h[1-6]|blockquote|tr|table)[^>]*>`)
	htmlTagPattern       = regexp.MustCompile(`<[^>]+>`)
	urlPattern           = regexp.MustCompile(`https?://\S+`)
	spacePattern         = regexp.MustCompile(`[^\S\r\n]+`)
	multiNewlinePattern  = regexp.MustCompile(`\n+`)
	dropMetaLinePattern  = regexp.MustCompile(`(?i)^(share( this article)?|sharing|print|posted by|author|source|tags?|filed under|leave a comment)\b`)
	shareMetaPattern     = regexp.MustCompile(`(?i)^(share|sharing|print|posted by|author|source|tags?|filed under)[:\s-]+`)
	noisyBlockPatterns   = []*regexp.Regexp{
		regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`),
		regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`),
		regexp.MustCompile(`(?is)<noscript[^>]*>.*?</noscript>`),
		regexp.MustCompile(`(?is)<svg[^>]*>.*?</svg>`),
		regexp.MustCompile(`(?is)<footer[^>]*>.*?</footer>`),
		regexp.MustCompile(`(?is)<nav[^>]*>.*?</nav>`),
	}
)

// CleanText 清洗 RSS 和网页文本，适合在进入 AI 处理前使用。
func CleanText(value string, limit int) string {
	if limit <= 0 {
		limit = defaultCleanTextLimit
	}

	for _, pattern := range noisyBlockPatterns {
		value = pattern.ReplaceAllString(value, " ")
	}
	value = blockBoundaryPattern.ReplaceAllString(value, "\n")
	value = html.UnescapeString(value)
	value = strings.ReplaceAll(value, "\u00a0", " ")
	value = htmlTagPattern.ReplaceAllString(value, " ")
	value = html.UnescapeString(value)
	value = strings.ReplaceAll(value, "\u00a0", " ")
	value = urlPattern.ReplaceAllString(value, " ")
	value = normalizeCleanLines(value)

	return truncateRunes(value, limit)
}

func normalizeCleanLines(value string) string {
	rawLines := strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n")
	lines := make([]string, 0, len(rawLines))
	seen := make(map[string]struct{}, len(rawLines))

	for _, rawLine := range rawLines {
		line := strings.TrimSpace(spacePattern.ReplaceAllString(rawLine, " "))
		line = strings.Trim(line, "-|:")
		if dropMetaLinePattern.MatchString(line) {
			continue
		}
		line = strings.TrimSpace(shareMetaPattern.ReplaceAllString(line, ""))
		if line == "" || isBoilerplateLine(line) {
			continue
		}

		key := strings.ToLower(line)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		lines = append(lines, line)
	}

	normalized := strings.Join(lines, "\n")
	normalized = multiNewlinePattern.ReplaceAllString(normalized, "\n")
	normalized = strings.ReplaceAll(normalized, "\n", " ")
	normalized = strings.TrimSpace(spacePattern.ReplaceAllString(normalized, " "))
	return normalized
}

func isBoilerplateLine(line string) bool {
	lower := strings.ToLower(strings.TrimSpace(line))
	for _, marker := range []string{
		"read more",
		"continue reading",
		"click here",
		"this article",
		"share this",
		"share on",
		"posted by",
		"filed under",
		"leave a comment",
		"subscribe",
		"all rights reserved",
		"copyright",
		"相关阅读",
		"延伸阅读",
		"订阅",
		"点击查看",
		"点击阅读",
		"阅读更多",
		"关注我们",
		"版权所有",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}

	switch lower {
	case "share", "print", "more", "tags", "author", "source", "this article", "阅读全文":
		return true
	}

	return false
}

func truncateRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}

	return string(runes[:limit])
}
