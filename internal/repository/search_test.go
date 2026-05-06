package repository

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"InfoHub-agent/internal/model"
)

func TestFileSearchRepositorySaveAndLoad(t *testing.T) {
	repo := NewFileSearchRepository(filepath.Join(t.TempDir(), "searches"))
	record := SearchRecord{
		Query:       "agent",
		GeneratedAt: time.Date(2026, 5, 6, 10, 0, 0, 0, time.UTC),
		Markdown:    "# search",
		Items:       []model.NewsItem{{Title: "Agent post", Score: 4}},
		Warnings:    []string{"reddit: timeout"},
	}

	if err := repo.Save(context.Background(), record); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := repo.Get(context.Background(), "20260506-100000")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if loaded.Query != "agent" || len(loaded.Warnings) != 1 {
		t.Fatalf("unexpected loaded record: %+v", loaded)
	}

	list, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 1 || list[0].Query != "agent" {
		t.Fatalf("unexpected metadata: %+v", list)
	}
}
