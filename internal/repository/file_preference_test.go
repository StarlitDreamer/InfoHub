package repository

import (
	"context"
	"path/filepath"
	"testing"
)

func TestFileUserPreferenceRepositorySaveAndGet(t *testing.T) {
	repo := NewFileUserPreferenceRepository(filepath.Join(t.TempDir(), "preferences", "users.json"))
	record := UserPreferenceRecord{
		UserID:   "alice",
		Tags:     []string{"AI", "Agent"},
		Sources:  []string{"openai-news"},
		Keywords: []string{"workflow"},
		Weights: PreferenceWeightValue{
			Tag:     1.4,
			Source:  1.1,
			Keyword: 0.8,
		},
	}

	if err := repo.Save(context.Background(), record); err != nil {
		t.Fatalf("save preference failed: %v", err)
	}

	saved, err := repo.Get(context.Background(), "alice")
	if err != nil {
		t.Fatalf("get preference failed: %v", err)
	}
	if saved.UserID != "alice" || len(saved.Tags) != 2 || saved.Weights.Tag != 1.4 {
		t.Fatalf("unexpected saved preference: %+v", saved)
	}
}

func TestFileUserPreferenceRepositoryReturnsNotFound(t *testing.T) {
	repo := NewFileUserPreferenceRepository(filepath.Join(t.TempDir(), "preferences.json"))

	_, err := repo.Get(context.Background(), "missing")
	if err != ErrUserPreferenceNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}
