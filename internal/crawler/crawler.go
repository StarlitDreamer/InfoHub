// Package crawler 定义信息采集模块的接口。
package crawler

import (
	"context"

	"InfoHub-agent/internal/model"
)

// Crawler 表示一个可插拔的数据采集器。
type Crawler interface {
	Fetch(ctx context.Context) ([]model.NewsItem, error)
}
