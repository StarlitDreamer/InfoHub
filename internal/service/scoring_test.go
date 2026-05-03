package service

import (
	"testing"
	"time"

	"InfoHub-agent/internal/model"
)

func TestSortByDecisionScore(t *testing.T) {
	now := time.Now()
	items := []model.NewsItem{
		{Title: "低价值", Score: 1, PublishTime: now},
		{Title: "高价值", Score: 5, PublishTime: now.Add(-24 * time.Hour)},
	}

	result := SortByDecisionScore(items, now)

	if result[0].Title != "高价值" {
		t.Fatalf("期望高价值内容排在前面，实际为 %s", result[0].Title)
	}
}
