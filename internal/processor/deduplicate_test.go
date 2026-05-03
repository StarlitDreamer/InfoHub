package processor

import (
	"testing"

	"InfoHub-agent/internal/model"
)

func TestDeduplicateItemsKeepsFirstItemWhenTitlesMatch(t *testing.T) {
	items := []model.NewsItem{
		{ID: 1, Title: "same title", Content: "first content"},
		{ID: 2, Title: "same title", Content: "second content"},
		{ID: 3, Title: "unique title", Content: "third content"},
	}

	result := DeduplicateItems(items)

	if len(result) != 2 {
		t.Fatalf("expected 2 items after title deduplication, got %d", len(result))
	}
	if result[0].ID != 1 {
		t.Fatalf("expected first duplicate item to be kept, got ID %d", result[0].ID)
	}
}

func TestDeduplicateItemsKeepsFirstItemWhenURLsMatch(t *testing.T) {
	items := []model.NewsItem{
		{ID: 1, Title: "alpha", URL: "https://example.com/post"},
		{ID: 2, Title: "beta", URL: "https://example.com/post/"},
		{ID: 3, Title: "gamma", URL: "https://example.com/other"},
	}

	result := DeduplicateItems(items)

	if len(result) != 2 {
		t.Fatalf("expected 2 items after url deduplication, got %d", len(result))
	}
	if result[0].ID != 1 || result[1].ID != 3 {
		t.Fatalf("unexpected url deduplication result: %+v", result)
	}
}

func TestDeduplicateItemsKeepsFirstItemWhenContentsMatch(t *testing.T) {
	items := []model.NewsItem{
		{ID: 1, Title: "alpha", Content: "Same   body text"},
		{ID: 2, Title: "beta", Content: " same body   text "},
		{ID: 3, Title: "gamma", Content: "different body"},
	}

	result := DeduplicateItems(items)

	if len(result) != 2 {
		t.Fatalf("expected 2 items after content deduplication, got %d", len(result))
	}
	if result[0].ID != 1 || result[1].ID != 3 {
		t.Fatalf("unexpected content deduplication result: %+v", result)
	}
}
