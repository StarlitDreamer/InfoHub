package service

import (
	"testing"
	"time"

	"InfoHub-agent/internal/model"
)

func TestSortByDecisionScore(t *testing.T) {
	now := time.Now()
	items := []model.NewsItem{
		{Title: "low", Score: 1, PublishTime: now},
		{Title: "high", Score: 5, PublishTime: now.Add(-24 * time.Hour)},
	}

	result := SortByDecisionScore(items, now)

	if result[0].Title != "high" {
		t.Fatalf("expected high score item first, got %s", result[0].Title)
	}
}

func TestLimitItems(t *testing.T) {
	items := []model.NewsItem{
		{Title: "a"},
		{Title: "b"},
		{Title: "c"},
	}

	limited := LimitItems(items, 2)
	if len(limited) != 2 || limited[0].Title != "a" || limited[1].Title != "b" {
		t.Fatalf("unexpected limited items: %+v", limited)
	}

	limited[0].Title = "changed"
	if items[0].Title != "a" {
		t.Fatal("expected LimitItems to return a copy")
	}
}
