package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"InfoHub-agent/internal/model"
)

func TestNewMySQLReportRepositoryEnsuresTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new failed: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS `reports`")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo, err := NewMySQLReportRepository(db, "reports")
	if err != nil {
		t.Fatalf("create repository failed: %v", err)
	}
	if repo.table != "reports" {
		t.Fatalf("expected table name reports, got %s", repo.table)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestMySQLReportRepositorySave(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new failed: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS `reports`")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo, err := NewMySQLReportRepository(db, "reports")
	if err != nil {
		t.Fatalf("create repository failed: %v", err)
	}

	generatedAt := time.Date(2026, 5, 3, 16, 0, 0, 0, time.FixedZone("CST", 8*3600))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `reports` (generated_at, markdown, items_json) VALUES (?, ?, ?)")).
		WithArgs(generatedAt.UTC(), "# report", `[{"ID":0,"SourceName":"","Title":"test title","Content":"","Source":"","URL":"","PublishTime":"0001-01-01T00:00:00Z","Tags":null,"Score":0}]`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Save(context.Background(), ReportRecord{
		GeneratedAt: generatedAt,
		Markdown:    "# report",
		Items:       []model.NewsItem{{Title: "test title"}},
	})
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestMySQLReportRepositoryLatestAndList(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new failed: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS `reports`")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo, err := NewMySQLReportRepository(db, "reports")
	if err != nil {
		t.Fatalf("create repository failed: %v", err)
	}

	latestTime := time.Date(2026, 5, 3, 16, 30, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT generated_at, markdown, items_json FROM `reports` ORDER BY generated_at DESC, id DESC LIMIT 1")).
		WillReturnRows(sqlmock.NewRows([]string{"generated_at", "markdown", "items_json"}).
			AddRow(latestTime, "# report\n\n## item\n- summary\n", `[{"ID":0,"SourceName":"","Title":"second","Content":"","Source":"","URL":"","PublishTime":"0001-01-01T00:00:00Z","Tags":null,"Score":0}]`))

	record, err := repo.Latest(context.Background())
	if err != nil {
		t.Fatalf("latest failed: %v", err)
	}
	if record.Markdown == "" || len(record.Items) != 1 || record.Items[0].Title != "second" {
		t.Fatalf("unexpected latest record: %+v", record)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT generated_at, markdown, items_json FROM `reports` ORDER BY generated_at DESC, id DESC")).
		WillReturnRows(sqlmock.NewRows([]string{"generated_at", "markdown", "items_json"}).
			AddRow(latestTime, "# report\n\n## item one\n- summary\n\n## item two\n- summary\n", `[{"ID":0,"SourceName":"","Title":"second","Content":"","Source":"","URL":"","PublishTime":"0001-01-01T00:00:00Z","Tags":null,"Score":0},{"ID":0,"SourceName":"","Title":"third","Content":"","Source":"","URL":"","PublishTime":"0001-01-01T00:00:00Z","Tags":null,"Score":0},{"ID":0,"SourceName":"","Title":"stored only","Content":"","Source":"","URL":"","PublishTime":"0001-01-01T00:00:00Z","Tags":null,"Score":0}]`).
			AddRow(time.Date(2026, 5, 3, 15, 30, 0, 0, time.UTC), "# report\n\n## older item\n- summary\n", `[{"ID":0,"SourceName":"","Title":"first","Content":"","Source":"","URL":"","PublishTime":"0001-01-01T00:00:00Z","Tags":null,"Score":0}]`))

	records, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(records) != 2 || records[0].Name != "20260503-163000" {
		t.Fatalf("unexpected list order: %+v", records)
	}
	if records[0].ItemCount != 3 || records[0].DisplayCount != 2 {
		t.Fatalf("unexpected list summary: %+v", records[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestMySQLReportRepositoryGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new failed: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS `reports`")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo, err := NewMySQLReportRepository(db, "reports")
	if err != nil {
		t.Fatalf("create repository failed: %v", err)
	}

	reportTime := time.Date(2026, 5, 3, 16, 30, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT generated_at, markdown, items_json FROM `reports` WHERE generated_at = ? ORDER BY id DESC LIMIT 1")).
		WithArgs(reportTime).
		WillReturnRows(sqlmock.NewRows([]string{"generated_at", "markdown", "items_json"}).
			AddRow(reportTime, "# report\n\n## item\n- summary\n", `[{"ID":0,"SourceName":"","Title":"second","Content":"","Source":"","URL":"","PublishTime":"0001-01-01T00:00:00Z","Tags":null,"Score":0}]`))

	record, err := repo.Get(context.Background(), "20260503-163000")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if record.Markdown == "" || len(record.Items) != 1 || record.Items[0].Title != "second" {
		t.Fatalf("unexpected report: %+v", record)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestMySQLReportRepositoryGetReturnsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new failed: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS `reports`")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo, err := NewMySQLReportRepository(db, "reports")
	if err != nil {
		t.Fatalf("create repository failed: %v", err)
	}

	reportTime := time.Date(2026, 5, 3, 16, 30, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT generated_at, markdown, items_json FROM `reports` WHERE generated_at = ? ORDER BY id DESC LIMIT 1")).
		WithArgs(reportTime).
		WillReturnRows(sqlmock.NewRows([]string{"generated_at", "markdown", "items_json"}))

	_, err = repo.Get(context.Background(), "20260503-163000")
	if err != ErrReportNotFound {
		t.Fatalf("expected ErrReportNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestMySQLReportRepositoryLatestReturnsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new failed: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS `reports`")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo, err := NewMySQLReportRepository(db, "reports")
	if err != nil {
		t.Fatalf("create repository failed: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT generated_at, markdown, items_json FROM `reports` ORDER BY generated_at DESC, id DESC LIMIT 1")).
		WillReturnRows(sqlmock.NewRows([]string{"generated_at", "markdown", "items_json"}))

	_, err = repo.Latest(context.Background())
	if err != ErrReportNotFound {
		t.Fatalf("expected ErrReportNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
