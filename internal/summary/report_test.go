package summary

import (
	"strings"
	"testing"

	"InfoHub-agent/internal/model"
)

func TestBuildOverview(t *testing.T) {
	overview := BuildOverview([]model.NewsItem{
		{Title: "alpha", Source: "OpenAI News", Score: 5},
		{Title: "beta", Source: "Google Blog", Score: 4},
		{Title: "gamma", Source: "OpenAI News", Score: 2, Tags: []string{"AI"}},
		{Title: "delta", Source: "", Score: 1},
	}, 3, 3)

	if overview.ItemCount != 4 || overview.HighPriorityCount != 2 {
		t.Fatalf("unexpected overview counts: %+v", overview)
	}
	if !strings.Contains(overview.SourceSummary, "Google Blog 1") || !strings.Contains(overview.SourceSummary, "未知来源 1") {
		t.Fatalf("unexpected source summary: %+v", overview)
	}
	if len(overview.TopTitles) != 3 || overview.TopTitles[0] != "alpha" {
		t.Fatalf("unexpected top titles: %+v", overview.TopTitles)
	}
	if len(overview.PriorityActions) == 0 {
		t.Fatalf("expected priority actions, got %+v", overview.PriorityActions)
	}
}

func TestGroupBySource(t *testing.T) {
	groups := GroupBySource([]model.NewsItem{
		{Title: "b1", Source: "Beta"},
		{Title: "a1", Source: "Alpha"},
		{Title: "a2", Source: "Alpha"},
		{Title: "u1", Source: ""},
	})

	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %+v", groups)
	}
	if groups[0].Source != "Alpha" || len(groups[0].Items) != 2 {
		t.Fatalf("unexpected first group: %+v", groups[0])
	}
	if groups[2].Source != "未知来源" || len(groups[2].Items) != 1 {
		t.Fatalf("unexpected unknown source group: %+v", groups[2])
	}
}
