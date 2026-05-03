package delivery

import (
	"strings"
	"testing"

	"InfoHub-agent/internal/model"
)

func TestRenderMarkdownIncludesTitleAndSummary(t *testing.T) {
	report := RenderMarkdown([]model.NewsItem{
		{Title: "test title", Content: "test summary", Score: 5},
	})

	if !strings.Contains(report, "# 今日信息") {
		t.Fatal("expected markdown to include report heading")
	}
	if !strings.Contains(report, "- 标题：test title") {
		t.Fatal("expected markdown to include item title")
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

func countMarkdownItemHeadings(report string) int {
	count := 0
	for _, line := range strings.Split(report, "\n") {
		if strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
			count++
		}
	}

	return count
}
