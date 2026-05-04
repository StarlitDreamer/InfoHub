package repository

import (
	"testing"
	"time"

	"InfoHub-agent/internal/model"
)

func TestBuildReportMetadataUsesSharedOverviewRules(t *testing.T) {
	items := []model.NewsItem{
		{Title: "测试一", Score: 5},
		{Title: "测试二", Score: 2},
		{Title: "测试三", Score: 4},
	}

	metadata := BuildReportMetadata(
		"20260504-120000",
		"reports/20260504-120000.md",
		"items/20260504-120000.json",
		"# 今日信息日报\n\n## 今日概览\n- 收录条目：3\n\n## ⭐⭐⭐⭐⭐ 测试一\nbody\n\n## ⭐⭐ 测试二\nbody\n",
		items,
		time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC),
		2,
	)

	if metadata.DisplayCount != 2 {
		t.Fatalf("expected display count 2, got %d", metadata.DisplayCount)
	}
	if metadata.HighPriorityCount != 2 {
		t.Fatalf("expected high priority count 2, got %d", metadata.HighPriorityCount)
	}
	if len(metadata.TopTitles) != 2 || metadata.TopTitles[0] != "测试一" || metadata.TopTitles[1] != "测试二" {
		t.Fatalf("expected top titles from shared overview, got %+v", metadata.TopTitles)
	}
}
