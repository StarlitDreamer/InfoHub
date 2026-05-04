package processor

import (
	"testing"
	"time"

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

func TestDeduplicateItemsMergesSimilarEvents(t *testing.T) {
	baseTime := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	items := []model.NewsItem{
		{
			ID:          1,
			Title:       "OpenAI releases GPT-5 enterprise coding model",
			Content:     "OpenAI released GPT-5 for enterprise teams. The model improves coding and analysis workflows.",
			Source:      "feed-a",
			SourceName:  "Feed A",
			URL:         "https://example.com/a",
			PublishTime: baseTime,
			Tags:        []string{"AI"},
			Score:       3,
		},
		{
			ID:          2,
			Title:       "OpenAI releases GPT-5 enterprise coding model with stronger analysis",
			Content:     "OpenAI released GPT-5 for enterprise teams with stronger coding, analysis, and deployment support.",
			Source:      "feed-b",
			SourceName:  "Feed B",
			URL:         "https://example.com/b",
			PublishTime: baseTime.Add(2 * time.Hour),
			Tags:        []string{"Enterprise", "AI"},
			Score:       5,
		},
	}

	result := DeduplicateItems(items)

	if len(result) != 1 {
		t.Fatalf("expected 1 merged item, got %d", len(result))
	}

	merged := result[0]
	if merged.ID != 1 {
		t.Fatalf("expected merged item to keep base ID, got %d", merged.ID)
	}
	if merged.Title != items[1].Title {
		t.Fatalf("expected longer title to be kept, got %q", merged.Title)
	}
	if merged.Content != items[1].Content {
		t.Fatalf("expected richer content to be kept, got %q", merged.Content)
	}
	if merged.PublishTime != items[1].PublishTime {
		t.Fatalf("expected newer publish time to be kept, got %v", merged.PublishTime)
	}
	if merged.Score != 5 {
		t.Fatalf("expected higher score to be kept, got %v", merged.Score)
	}
	if len(merged.Tags) != 2 || merged.Tags[0] != "AI" || merged.Tags[1] != "Enterprise" {
		t.Fatalf("expected merged tags, got %+v", merged.Tags)
	}
	if merged.Source != items[0].Source {
		t.Fatalf("expected original non-empty source to be kept, got %q", merged.Source)
	}
	if merged.URL != items[0].URL {
		t.Fatalf("expected original non-empty url to be kept, got %q", merged.URL)
	}
}

func TestDeduplicateItemsDoesNotMergeDistantSimilarEvents(t *testing.T) {
	baseTime := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	items := []model.NewsItem{
		{
			ID:          1,
			Title:       "OpenAI releases GPT-5 enterprise coding model",
			Content:     "OpenAI released GPT-5 for enterprise teams.",
			PublishTime: baseTime,
		},
		{
			ID:          2,
			Title:       "OpenAI releases GPT-5 enterprise coding model with stronger analysis",
			Content:     "OpenAI released GPT-5 for enterprise teams with stronger analysis.",
			PublishTime: baseTime.Add(72 * time.Hour),
		},
	}

	result := DeduplicateItems(items)

	if len(result) != 2 {
		t.Fatalf("expected distant items to stay separate, got %d", len(result))
	}
}

func TestDeduplicateItemsDoesNotMergeUnrelatedItems(t *testing.T) {
	baseTime := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	items := []model.NewsItem{
		{
			ID:          1,
			Title:       "OpenAI releases GPT-5 enterprise coding model",
			Content:     "OpenAI released GPT-5 for enterprise teams.",
			PublishTime: baseTime,
		},
		{
			ID:          2,
			Title:       "Apple reports record iPhone revenue in Q2",
			Content:     "Apple reported record iPhone revenue and expanded services growth.",
			PublishTime: baseTime.Add(2 * time.Hour),
		},
	}

	result := DeduplicateItems(items)

	if len(result) != 2 {
		t.Fatalf("expected unrelated items to stay separate, got %d", len(result))
	}
}

func TestDeduplicateItemsMergesTitlesWithSourceSuffixes(t *testing.T) {
	baseTime := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	items := []model.NewsItem{
		{
			ID:          1,
			Title:       "OpenAI releases GPT-5 enterprise coding model - OpenAI News",
			Content:     "OpenAI released GPT-5 for enterprise teams with stronger coding and analysis workflows.",
			PublishTime: baseTime,
		},
		{
			ID:          2,
			Title:       "OpenAI releases GPT-5 enterprise coding model",
			Content:     "OpenAI released GPT-5 for enterprise teams with stronger coding and analysis workflows.",
			PublishTime: baseTime.Add(90 * time.Minute),
		},
	}

	result := DeduplicateItems(items)

	if len(result) != 1 {
		t.Fatalf("expected source suffix variants to merge, got %d items", len(result))
	}
}
