// Package repository 提供信息和日报的持久化能力。
package repository

import (
	"context"
	"time"

	"InfoHub-agent/internal/model"
)

// ReportRecord 表示一次日报生成结果。
type ReportRecord struct {
	GeneratedAt time.Time
	Markdown    string
	Items       []model.NewsItem
}

// ReportRepository 定义日报存储接口。
type ReportRepository interface {
	Save(ctx context.Context, record ReportRecord) error
	Latest(ctx context.Context) (ReportRecord, error)
	List(ctx context.Context) ([]ReportMetadata, error)
}

// ReportMetadata 表示历史日报的文件索引信息。
type ReportMetadata struct {
	Name      string    `json:"name"`
	Markdown  string    `json:"markdown"`
	Items     string    `json:"items"`
	CreatedAt time.Time `json:"created_at"`
}
