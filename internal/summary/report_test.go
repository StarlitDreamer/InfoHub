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
