package repository

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileSearchRepository 将搜索结果保存到本地文件。
type FileSearchRepository struct {
	root string
}

// NewFileSearchRepository 创建文件型搜索结果存储。
func NewFileSearchRepository(root string) *FileSearchRepository {
	return &FileSearchRepository{root: root}
}

// Save 保存搜索结果快照。
func (r *FileSearchRepository) Save(ctx context.Context, record SearchRecord) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(r.root, "records"), 0755); err != nil {
		return err
	}

	name := record.GeneratedAt.Format("20060102-150405")
	payload, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(r.root, "records", name+".json"), payload, 0644)
}

// Latest 读取最近一次搜索结果。
func (r *FileSearchRepository) Latest(ctx context.Context) (SearchRecord, error) {
	records, err := r.List(ctx)
	if err != nil {
		return SearchRecord{}, err
	}
	if len(records) == 0 {
		return SearchRecord{}, ErrReportNotFound
	}

	return r.Get(ctx, records[0].Name)
}

// Get 按名称读取指定搜索结果。
func (r *FileSearchRepository) Get(ctx context.Context, name string) (SearchRecord, error) {
	if err := ctx.Err(); err != nil {
		return SearchRecord{}, err
	}

	content, err := os.ReadFile(filepath.Join(r.root, "records", name+".json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return SearchRecord{}, ErrReportNotFound
		}
		return SearchRecord{}, err
	}

	var record SearchRecord
	if err := json.Unmarshal(content, &record); err != nil {
		return SearchRecord{}, err
	}

	return record, nil
}

// List 返回搜索历史索引。
func (r *FileSearchRepository) List(ctx context.Context) ([]SearchMetadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(filepath.Join(r.root, "records"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	result := make([]SearchMetadata, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(r.root, "records", entry.Name()))
		if err != nil {
			return nil, err
		}
		var record SearchRecord
		if err := json.Unmarshal(content, &record); err != nil {
			return nil, err
		}
		name := strings.TrimSuffix(entry.Name(), ".json")
		result = append(result, BuildSearchMetadata(name, record.Query, record.Markdown, record.Items, record.GeneratedAt, 2))
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result, nil
}
