package repository

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"InfoHub-agent/internal/model"
)

// ErrReportNotFound 表示当前还没有可读取的日报。
var ErrReportNotFound = errors.New("日报不存在")

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

// Latest 读取最近一次生成的日报。
func (r *FileReportRepository) Latest(ctx context.Context) (ReportRecord, error) {
	records, err := r.List(ctx)
	if err != nil {
		return ReportRecord{}, err
	}

	if len(records) == 0 {
		return ReportRecord{}, ErrReportNotFound
	}

	latest := records[0]
	markdown, err := os.ReadFile(filepath.Join(r.root, latest.Markdown))
	if err != nil {
		return ReportRecord{}, err
	}

	items, err := readItems(filepath.Join(r.root, latest.Items))
	if err != nil {
		return ReportRecord{}, err
	}

	return ReportRecord{
		GeneratedAt: latest.CreatedAt,
		Markdown:    string(markdown),
		Items:       items,
	}, nil
}

// List 返回历史日报索引，按生成时间倒序排列。
func (r *FileReportRepository) List(ctx context.Context) ([]ReportMetadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(filepath.Join(r.root, "reports"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, err
	}

	records := make([]ReportMetadata, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")
		createdAt, err := time.Parse("20060102-150405", name)
		if err != nil {
			continue
		}

		markdownPath := filepath.Join(r.root, "reports", entry.Name())
		itemsPath := filepath.Join(r.root, "items", name+".json")
		markdownContent, err := os.ReadFile(markdownPath)
		if err != nil {
			return nil, err
		}
		items, err := readItems(itemsPath)
		if err != nil {
			return nil, err
		}

		records = append(records, BuildReportMetadata(
			name,
			filepath.ToSlash(filepath.Join("reports", entry.Name())),
			filepath.ToSlash(filepath.Join("items", name+".json")),
			string(markdownContent),
			items,
			createdAt,
			2,
		))
	}

	sort.SliceStable(records, func(i, j int) bool {
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})

	return records, nil
}

func readItems(path string) ([]model.NewsItem, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var items []model.NewsItem
	if err := json.Unmarshal(content, &items); err != nil {
		return nil, err
	}

	return items, nil
}
