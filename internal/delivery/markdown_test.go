package delivery

import (
	"strings"
	"testing"
	"time"

	"InfoHub-agent/internal/model"
)

func TestRenderMarkdownIncludesDecisionFields(t *testing.T) {
	report := RenderMarkdown([]model.NewsItem{
		{
			Title:       "test title",
			Content:     "【标题】test title\n【发生了什么】summary\n【为什么重要】important\n【影响】impact\n【评分】4",
			Source:      "OpenAI News",
			URL:         "https://example.com/item",
			PublishTime: time.Date(2026, 5, 4, 9, 30, 0, 0, time.UTC),
			Tags:        []string{"AI", "Agent"},
			Score:       4,
		},
	})

	if !strings.Contains(report, "# 今日信息") {
		t.Fatal("expected markdown to include report heading")
	}
	if !strings.Contains(report, "## ⭐⭐⭐⭐ test title") {
		t.Fatal("expected markdown to include scored item heading")
	}
	if !strings.Contains(report, "- 为什么重要：important") {
		t.Fatal("expected markdown to include importance field")
	}
	if !strings.Contains(report, "- 原文链接：https://example.com/item") {
		t.Fatal("expected markdown to include source url")
	}
}

func TestRenderMarkdownEmptyItems(t *testing.T) {
	report := RenderMarkdown(nil)

	if !strings.Contains(report, "今日暂无新增信息。") {
		t.Fatal("expected empty report message")
	}
}

func TestRenderMarkdownBySourceGroupsItems(t *testing.T) {
	report := RenderMarkdownBySource([]model.NewsItem{
		{Title: "openai item", Content: "summary a", Source: "OpenAI News", Score: 5},
		{Title: "google item", Content: "summary b", Source: "Google Blog", Score: 4},
		{Title: "second openai item", Content: "summary c", Source: "OpenAI News", Score: 3},
	})

	if !strings.Contains(report, "### 来源：Google Blog") || !strings.Contains(report, "### 来源：OpenAI News") {
		t.Fatalf("expected source headings, got %s", report)
	}
	if countMarkdownItemHeadings(report) != 3 {
		t.Fatalf("expected grouped report to preserve item headings, got %s", report)
	}
}

func TestRenderMarkdownFallsBackForPlainSummary(t *testing.T) {
	report := RenderMarkdown([]model.NewsItem{
		{Title: "plain title", Content: "plain summary", Score: 3},
	})

	if !strings.Contains(report, "- 发生了什么：plain summary") {
		t.Fatalf("expected plain summary fallback, got %s", report)
	}
	if !strings.Contains(report, "- 为什么重要：该信息可能影响后续判断") {
		t.Fatalf("expected default importance fallback, got %s", report)
	}
}

func countMarkdownItemHeadings(report string) int {
	count := 0
	for _, line := range strings.Split(report, "\n") {
		if strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
			count++
		}
	}

	return count
}
