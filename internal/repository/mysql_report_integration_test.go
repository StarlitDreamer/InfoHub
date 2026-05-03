package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"InfoHub-agent/internal/model"

	_ "github.com/go-sql-driver/mysql"
)

func TestMySQLReportRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skip mysql integration test in short mode")
	}

	dsn := strings.TrimSpace(os.Getenv("INFOHUB_TEST_MYSQL_DSN"))
	if dsn == "" {
		t.Skip("skip mysql integration test without INFOHUB_TEST_MYSQL_DSN")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open mysql failed: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping mysql failed: %v", err)
	}

	table := fmt.Sprintf("reports_it_%d", time.Now().UnixNano())
	repo, err := NewMySQLReportRepository(db, table)
	if err != nil {
		t.Fatalf("create mysql repository failed: %v", err)
	}
	defer dropIntegrationTable(t, db, table)

	records := []ReportRecord{
		{
			GeneratedAt: time.Date(2026, 5, 3, 15, 30, 0, 0, time.UTC),
			Markdown:    "# report\n\n## item one\n- summary: one\n",
			Items: []model.NewsItem{
				{Title: "item one"},
			},
		},
		{
			GeneratedAt: time.Date(2026, 5, 3, 16, 30, 0, 0, time.UTC),
			Markdown:    "# report\n\n## item two\n- summary: two\n\n## item three\n- summary: three\n",
			Items: []model.NewsItem{
				{Title: "item two", URL: "https://example.com/two", Score: 5},
				{Title: "item three", URL: "https://example.com/three", Score: 4},
				{Title: "stored only", URL: "https://example.com/stored", Score: 1},
			},
		},
	}

	for _, record := range records {
		if err := repo.Save(ctx, record); err != nil {
			t.Fatalf("save record failed: %v", err)
		}
	}

	latest, err := repo.Latest(ctx)
	if err != nil {
		t.Fatalf("load latest report failed: %v", err)
	}
	if latest.GeneratedAt.UTC() != records[1].GeneratedAt.UTC() {
		t.Fatalf("expected latest generated_at %s, got %s", records[1].GeneratedAt.UTC(), latest.GeneratedAt.UTC())
	}
	if latest.Markdown != records[1].Markdown {
		t.Fatalf("expected latest markdown to match saved value, got %s", latest.Markdown)
	}
	if len(latest.Items) != 3 || latest.Items[2].Title != "stored only" {
		t.Fatalf("expected latest items to preserve stored payload, got %+v", latest.Items)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list reports failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 report records, got %d", len(list))
	}
	if list[0].Name != "20260503-163000" {
		t.Fatalf("expected latest record name first, got %s", list[0].Name)
	}
	if list[0].ItemCount != 3 || list[0].DisplayCount != 2 {
		t.Fatalf("expected latest summary counts 3/2, got %+v", list[0])
	}
	if list[1].ItemCount != 1 || list[1].DisplayCount != 1 {
		t.Fatalf("expected earlier summary counts 1/1, got %+v", list[1])
	}
}

func dropIntegrationTable(t *testing.T, db *sql.DB, table string) {
	t.Helper()

	query := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", table)
	if _, err := db.Exec(query); err != nil {
		t.Fatalf("drop integration table failed: %v", err)
	}
}
