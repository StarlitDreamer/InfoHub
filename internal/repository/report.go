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
}
