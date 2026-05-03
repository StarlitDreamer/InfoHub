package repository

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
)

// FileReportRepository 将日报和信息条目保存到本地文件。
type FileReportRepository struct {
	root string
}

// NewFileReportRepository 创建文件型日报存储。
func NewFileReportRepository(root string) *FileReportRepository {
	return &FileReportRepository{root: root}
}

// Save 保存 Markdown 日报和 NewsItem JSON 快照。
func (r *FileReportRepository) Save(ctx context.Context, record ReportRecord) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(r.root, "reports"), 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(r.root, "items"), 0755); err != nil {
		return err
	}

	name := record.GeneratedAt.Format("20060102-150405")
	if err := os.WriteFile(filepath.Join(r.root, "reports", name+".md"), []byte(record.Markdown), 0644); err != nil {
		return err
	}

	items, err := json.MarshalIndent(record.Items, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(r.root, "items", name+".json"), items, 0644)
}
