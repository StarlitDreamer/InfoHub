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
		t.Fatalf("创建 sqlmock 失败：%v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS `reports`")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo, err := NewMySQLReportRepository(db, "reports")
	if err != nil {
		t.Fatalf("创建 MySQL 仓储失败：%v", err)
	}

	if repo.table != "reports" {
		t.Fatalf("期望表名为 reports，实际为 %s", repo.table)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("未满足的 SQL 预期：%v", err)
	}
}

func TestMySQLReportRepositorySave(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建 sqlmock 失败：%v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS `reports`")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo, err := NewMySQLReportRepository(db, "reports")
	if err != nil {
		t.Fatalf("创建 MySQL 仓储失败：%v", err)
	}

	generatedAt := time.Date(2026, 5, 3, 16, 0, 0, 0, time.FixedZone("CST", 8*3600))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `reports` (generated_at, markdown, items_json) VALUES (?, ?, ?)")).
		WithArgs(generatedAt.UTC(), "# 今日信息", `[{"ID":0,"Title":"测试标题","Content":"","Source":"","URL":"","PublishTime":"0001-01-01T00:00:00Z","Tags":null,"Score":0}]`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Save(context.Background(), ReportRecord{
		GeneratedAt: generatedAt,
		Markdown:    "# 今日信息",
		Items:       []model.NewsItem{{Title: "测试标题"}},
	})
	if err != nil {
		t.Fatalf("保存日报失败：%v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("未满足的 SQL 预期：%v", err)
	}
}

func TestMySQLReportRepositoryLatestAndList(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建 sqlmock 失败：%v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS `reports`")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo, err := NewMySQLReportRepository(db, "reports")
	if err != nil {
		t.Fatalf("创建 MySQL 仓储失败：%v", err)
	}

	latestTime := time.Date(2026, 5, 3, 16, 30, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT generated_at, markdown, items_json FROM `reports` ORDER BY generated_at DESC, id DESC LIMIT 1")).
		WillReturnRows(sqlmock.NewRows([]string{"generated_at", "markdown", "items_json"}).
			AddRow(latestTime, "# 第二份日报", `[{"ID":0,"Title":"第二条","Content":"","Source":"","URL":"","PublishTime":"0001-01-01T00:00:00Z","Tags":null,"Score":0}]`))

	record, err := repo.Latest(context.Background())
	if err != nil {
		t.Fatalf("读取最新日报失败：%v", err)
	}

	if record.Markdown != "# 第二份日报" || len(record.Items) != 1 || record.Items[0].Title != "第二条" {
		t.Fatalf("最新日报内容不符合预期：%+v", record)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT generated_at FROM `reports` ORDER BY generated_at DESC, id DESC")).
		WillReturnRows(sqlmock.NewRows([]string{"generated_at"}).
			AddRow(latestTime).
			AddRow(time.Date(2026, 5, 3, 15, 30, 0, 0, time.UTC)))

	records, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("读取日报列表失败：%v", err)
	}

	if len(records) != 2 || records[0].Name != "20260503-163000" {
		t.Fatalf("日报列表不符合预期：%+v", records)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("未满足的 SQL 预期：%v", err)
	}
}

func TestMySQLReportRepositoryLatestReturnsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建 sqlmock 失败：%v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS `reports`")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	repo, err := NewMySQLReportRepository(db, "reports")
	if err != nil {
		t.Fatalf("创建 MySQL 仓储失败：%v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT generated_at, markdown, items_json FROM `reports` ORDER BY generated_at DESC, id DESC LIMIT 1")).
		WillReturnRows(sqlmock.NewRows([]string{"generated_at", "markdown", "items_json"}))

	_, err = repo.Latest(context.Background())
	if err != ErrReportNotFound {
		t.Fatalf("期望返回 ErrReportNotFound，实际为 %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("未满足的 SQL 预期：%v", err)
	}
}
