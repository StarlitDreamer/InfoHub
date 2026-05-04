package delivery

import (
	"strings"
	"testing"
	"time"

	"InfoHub-agent/internal/model"
	"InfoHub-agent/internal/summary"
)

func TestRenderMarkdownIncludesOverviewAndDecisionFields(t *testing.T) {
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

	if !strings.Contains(report, "# 今日信息日报") {
		t.Fatal("expected markdown to include report heading")
	}
	if !strings.Contains(report, "## 今日概览") {
		t.Fatal("expected markdown to include overview section")
	}
	if !strings.Contains(report, "- 收录条目：1") {
		t.Fatal("expected markdown to include item count")
	}
	if !strings.Contains(report, "- 今日优先动作：") {
		t.Fatal("expected markdown to include priority actions section")
	}
	if !strings.Contains(report, "## ⭐⭐⭐⭐ test title") {
		t.Fatal("expected markdown to include scored item heading")
	}
	if !strings.Contains(report, "- 为什么重要：important") {
		t.Fatal("expected markdown to include importance field")
	}
	if !strings.Contains(report, "- 建议动作：加入近期待办，指定负责人跟进原文和后续进展。") {
		t.Fatal("expected markdown to include suggested action")
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
	if !strings.Contains(report, "- 为什么重要：该信息可能影响后续判断，建议结合业务上下文继续关注。") {
		t.Fatalf("expected default importance fallback, got %s", report)
	}
}

func TestRenderMarkdownOverviewSummarizesSourcesAndTopItems(t *testing.T) {
	report := RenderMarkdown([]model.NewsItem{
		{Title: "alpha", Content: "a", Source: "OpenAI News", Score: 5},
		{Title: "beta", Content: "b", Source: "Google Blog", Score: 4},
		{Title: "gamma", Content: "c", Source: "OpenAI News", Score: 2},
		{Title: "delta", Content: "d", Source: "", Score: 1},
	})

	if !strings.Contains(report, "- 高优先级：2") {
		t.Fatalf("expected high priority count, got %s", report)
	}
	if !strings.Contains(report, "- 来源分布：Google Blog 1；OpenAI News 2；未知来源 1") {
		t.Fatalf("expected source distribution, got %s", report)
	}
	if !strings.Contains(report, "- 重点关注：alpha；beta；gamma") {
		t.Fatalf("expected top item summary, got %s", report)
	}
}

func TestSummarizePriorityActionsDeduplicates(t *testing.T) {
	actions := summarizePriorityActions([]model.NewsItem{
		{Title: "alpha", Score: 5},
		{Title: "beta", Score: 5},
		{Title: "gamma", Score: 3, Tags: []string{"AI"}},
	}, 3)

	if len(actions) != 2 {
		t.Fatalf("expected deduplicated actions, got %+v", actions)
	}
	if !strings.Contains(actions[0], "优先安排评审") {
		t.Fatalf("expected highest priority action first, got %+v", actions)
	}
}

func TestRecommendActionUsesScoreAndTopicSignals(t *testing.T) {
	high := summary.RecommendAction(model.NewsItem{Score: 5}, summary.Structured{})
	if !strings.Contains(high.Description, "优先安排评审") {
		t.Fatalf("expected top score action, got %+v", high)
	}

	database := summary.RecommendAction(
		model.NewsItem{Score: 3, Tags: []string{"Database"}},
		summary.Structured{WhyImportant: "数据库性能值得关注"},
	)
	if !strings.Contains(database.Description, "专项验证") {
		t.Fatalf("expected database action, got %+v", database)
	}
}

func countMarkdownItemHeadings(report string) int {
	count := 0
	for _, line := range strings.Split(report, "\n") {
		if strings.HasPrefix(line, "## ⭐") {
			count++
		}
	}

	return count
}
