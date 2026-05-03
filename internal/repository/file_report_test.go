package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"InfoHub-agent/internal/model"
)

func TestFileReportRepositorySave(t *testing.T) {
	root := t.TempDir()
	generatedAt := time.Date(2026, 5, 3, 15, 30, 0, 0, time.UTC)
	repo := NewFileReportRepository(root)

	err := repo.Save(context.Background(), ReportRecord{
		GeneratedAt: generatedAt,
		Markdown:    "# 今日信息",
		Items:       []model.NewsItem{{Title: "测试标题"}},
	})
	if err != nil {
		t.Fatalf("保存日报失败：%v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "reports", "20260503-153000.md")); err != nil {
		t.Fatalf("期望 Markdown 日报文件存在：%v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "items", "20260503-153000.json")); err != nil {
		t.Fatalf("期望 NewsItem JSON 文件存在：%v", err)
	}
}
