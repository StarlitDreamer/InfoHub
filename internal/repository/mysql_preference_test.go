package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestNewMySQLUserPreferenceRepositoryEnsuresTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new failed: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS `user_preferences`")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo, err := NewMySQLUserPreferenceRepository(db, "user_preferences")
	if err != nil {
		t.Fatalf("create preference repository failed: %v", err)
	}
	if repo.table != "user_preferences" {
		t.Fatalf("expected table user_preferences, got %s", repo.table)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestMySQLUserPreferenceRepositorySaveAndGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new failed: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS `user_preferences`")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo, err := NewMySQLUserPreferenceRepository(db, "user_preferences")
	if err != nil {
		t.Fatalf("create preference repository failed: %v", err)
	}

	updatedAt := time.Date(2026, 5, 4, 9, 0, 0, 0, time.UTC)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `user_preferences` (user_id, tags_json, sources_json, keywords_json, tag_weight, source_weight, keyword_weight, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE tags_json = VALUES(tags_json), sources_json = VALUES(sources_json), keywords_json = VALUES(keywords_json), tag_weight = VALUES(tag_weight), source_weight = VALUES(source_weight), keyword_weight = VALUES(keyword_weight), updated_at = VALUES(updated_at)")).
		WithArgs("alice", `["AI"]`, `["openai-news"]`, `["agent"]`, 1.5, 1.1, 0.7, updatedAt.UTC()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Save(context.Background(), UserPreferenceRecord{
		UserID:    "alice",
		Tags:      []string{"AI"},
		Sources:   []string{"openai-news"},
		Keywords:  []string{"agent"},
		Weights:   PreferenceWeightValue{Tag: 1.5, Source: 1.1, Keyword: 0.7},
		UpdatedAt: updatedAt,
	})
	if err != nil {
		t.Fatalf("save preference failed: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, tags_json, sources_json, keywords_json, tag_weight, source_weight, keyword_weight, updated_at FROM `user_preferences` WHERE user_id = ? LIMIT 1")).
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "tags_json", "sources_json", "keywords_json", "tag_weight", "source_weight", "keyword_weight", "updated_at"}).
			AddRow("alice", `["AI"]`, `["openai-news"]`, `["agent"]`, 1.5, 1.1, 0.7, updatedAt))

	record, err := repo.Get(context.Background(), "alice")
	if err != nil {
		t.Fatalf("get preference failed: %v", err)
	}
	if record.UserID != "alice" || len(record.Tags) != 1 || record.Weights.Tag != 1.5 {
		t.Fatalf("unexpected preference record: %+v", record)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestMySQLUserPreferenceRepositoryGetReturnsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new failed: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS `user_preferences`")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo, err := NewMySQLUserPreferenceRepository(db, "user_preferences")
	if err != nil {
		t.Fatalf("create preference repository failed: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, tags_json, sources_json, keywords_json, tag_weight, source_weight, keyword_weight, updated_at FROM `user_preferences` WHERE user_id = ? LIMIT 1")).
		WithArgs("missing").
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "tags_json", "sources_json", "keywords_json", "tag_weight", "source_weight", "keyword_weight", "updated_at"}))

	_, err = repo.Get(context.Background(), "missing")
	if err != ErrUserPreferenceNotFound {
		t.Fatalf("expected ErrUserPreferenceNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
