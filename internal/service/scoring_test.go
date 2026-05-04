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

func TestLimitItemsBalancedBySourceAlternatesSourcesWhenPossible(t *testing.T) {
	items := []model.NewsItem{
		{Source: "google", Title: "g1"},
		{Source: "google", Title: "g2"},
		{Source: "google", Title: "g3"},
		{Source: "openai", Title: "o1"},
		{Source: "openai", Title: "o2"},
	}

	limited := LimitItemsBalancedBySource(items, 4)

	if len(limited) != 4 {
		t.Fatalf("expected 4 items, got %d", len(limited))
	}
	expected := []string{"g1", "o1", "g2", "o2"}
	for index, title := range expected {
		if limited[index].Title != title {
			t.Fatalf("expected %s at %d, got %+v", title, index, limited)
		}
	}
}

func TestLimitItemsBalancedBySourceFallsBackToOriginalOrderForSingleSource(t *testing.T) {
	items := []model.NewsItem{
		{Source: "google", Title: "g1"},
		{Source: "google", Title: "g2"},
		{Source: "google", Title: "g3"},
	}

	limited := LimitItemsBalancedBySource(items, 2)

	if len(limited) != 2 || limited[0].Title != "g1" || limited[1].Title != "g2" {
		t.Fatalf("expected original order for single source, got %+v", limited)
	}
}

func TestSortByDecisionScoreUsesStableTieBreakers(t *testing.T) {
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	items := []model.NewsItem{
		{Title: "b", URL: "https://example.com/b", Score: 3, PublishTime: now},
		{Title: "a", URL: "https://example.com/a", Score: 3, PublishTime: now},
	}

	result := SortByDecisionScore(items, now)

	if result[0].Title != "a" {
		t.Fatalf("expected title-based tie breaker to put a first, got %s", result[0].Title)
	}
}
