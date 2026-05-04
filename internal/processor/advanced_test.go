package processor

import (
	"testing"

	"InfoHub-agent/internal/model"
)

func TestDeduplicateByEmbeddingMergesSimilarContent(t *testing.T) {
	provider := NewKeywordEmbeddingProvider([]string{"ai", "model", "database"})
	items := []model.NewsItem{
		{ID: 1, Title: "AI model update", Content: "AI model improves speed"},
		{ID: 2, Title: "model AI release", Content: "speed improves AI model"},
		{ID: 3, Title: "database index", Content: "database performance"},
	}

	result := DeduplicateByEmbedding(items, provider, 0.9)

	if len(result) != 2 {
		t.Fatalf("expected 2 items after embedding deduplication, got %d", len(result))
	}
}
