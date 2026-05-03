package crawler

import (
	"time"

	"InfoHub-agent/internal/model"
)

// DemoCrawler 返回固定的模拟数据，用于验证 MVP 流程。
type DemoCrawler struct{}

// NewDemoCrawler 创建演示采集器实例。
func NewDemoCrawler() *DemoCrawler {
	return &DemoCrawler{}
}

// Fetch 采集模拟信息，暂不接入真实外部 API。
func (c *DemoCrawler) Fetch() ([]model.NewsItem, error) {
	now := time.Now()

	return []model.NewsItem{
		{
			ID:          1,
			Title:       "开源模型发布新版本",
			Content:     "某开源模型发布新版本，提升了推理速度和上下文处理能力。",
			Source:      "DemoSource",
			URL:         "https://example.com/open-model",
			PublishTime: now.Add(-2 * time.Hour),
			Tags:        []string{"AI", "开源"},
		},
		{
			ID:          2,
			Title:       "云厂商推出开发者工具更新",
			Content:     "云厂商更新开发者工具链，重点改善部署体验和监控能力。",
			Source:      "DemoSource",
			URL:         "https://example.com/cloud-tooling",
			PublishTime: now.Add(-90 * time.Minute),
			Tags:        []string{"云计算", "开发者工具"},
		},
		{
			ID:          3,
			Title:       "数据库社区讨论新索引策略",
			Content:     "数据库社区正在讨论新的索引策略，以改善高并发查询场景的性能。",
			Source:      "DemoSource",
			URL:         "https://example.com/database-index",
			PublishTime: now.Add(-45 * time.Minute),
			Tags:        []string{"数据库", "性能"},
		},
	}, nil
}
