package ai

import (
	"testing"

	"InfoHub-agent/internal/model"
)

func TestAnalyzeItemPrefersAnalyzer(t *testing.T) {
	processor := &analyzerStub{
		analysis: Analysis{
			Tags:    []string{"AI"},
			Summary: "summary",
			Score:   5,
		},
	}

	analysis, err := AnalyzeItem(processor, model.NewsItem{Title: "test"})
	if err != nil {
		t.Fatalf("analyze item failed: %v", err)
	}
	if processor.analyzeCalls != 1 {
		t.Fatalf("expected Analyze to be called once, got %d", processor.analyzeCalls)
	}
	if analysis.Score != 5 || analysis.Summary != "summary" {
		t.Fatalf("unexpected analysis: %+v", analysis)
	}
}

func TestMockProcessorImplementsSplitCapabilities(t *testing.T) {
	processor := NewMockProcessor()
	item := model.NewsItem{
		Title:   "OpenAI model update",
		Content: "new model release",
	}

	tags, err := processor.Classify(item)
	if err != nil {
		t.Fatalf("classify failed: %v", err)
	}
	summary, err := processor.Summarize(item)
	if err != nil {
		t.Fatalf("summarize failed: %v", err)
	}
	score, err := processor.Score(item)
	if err != nil {
		t.Fatalf("score failed: %v", err)
	}

	if len(tags) == 0 || tags[0] != "AI" {
		t.Fatalf("unexpected tags: %+v", tags)
	}
	if summary == "" {
		t.Fatal("expected non-empty summary")
	}
	if score < 1 || score > 5 {
		t.Fatalf("expected score in range 1..5, got %v", score)
	}
}

type analyzerStub struct {
	analysis     Analysis
	analyzeCalls int
}

func (s *analyzerStub) Classify(item model.NewsItem) ([]string, error) {
	return []string{"fallback"}, nil
}

func (s *analyzerStub) Summarize(item model.NewsItem) (string, error) {
	return "fallback", nil
}

func (s *analyzerStub) Score(item model.NewsItem) (float64, error) {
	return 1, nil
}

func (s *analyzerStub) Analyze(item model.NewsItem) (Analysis, error) {
	s.analyzeCalls++
	return s.analysis, nil
}
