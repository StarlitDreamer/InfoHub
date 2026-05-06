package repository

import (
	"context"
	"time"

	"InfoHub-agent/internal/model"
)

// SearchRecord 表示一次搜索结果。
type SearchRecord struct {
	Query       string
	GeneratedAt time.Time
	Markdown    string
	Items       []model.NewsItem
	Warnings    []string
}

// SearchMetadata 表示搜索历史索引信息。
type SearchMetadata struct {
	Name              string    `json:"name"`
	Query             string    `json:"query"`
	ItemCount         int       `json:"item_count"`
	DisplayCount      int       `json:"display_count"`
	HighPriorityCount int       `json:"high_priority_count"`
	TopTitles         []string  `json:"top_titles"`
	CreatedAt         time.Time `json:"created_at"`
}

// SearchRepository 定义搜索结果存储接口。
type SearchRepository interface {
	Save(ctx context.Context, record SearchRecord) error
	Latest(ctx context.Context) (SearchRecord, error)
	Get(ctx context.Context, name string) (SearchRecord, error)
	List(ctx context.Context) ([]SearchMetadata, error)
}

// BuildSearchMetadata 根据搜索结果生成索引元数据。
func BuildSearchMetadata(name, query, markdown string, items []model.NewsItem, createdAt time.Time, titleLimit int) SearchMetadata {
	overview := BuildReportOverview(markdown, items, titleLimit)
	return SearchMetadata{
		Name:              name,
		Query:             query,
		ItemCount:         len(items),
		DisplayCount:      overview.DisplayCount,
		HighPriorityCount: overview.HighPriorityCount,
		TopTitles:         overview.TopTitles,
		CreatedAt:         createdAt,
	}
}
