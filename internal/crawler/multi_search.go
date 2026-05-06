package crawler

import (
	"context"
	"fmt"

	"InfoHub-agent/internal/model"
)

// MultiSearchCrawler 聚合多个搜索型采集器。
type MultiSearchCrawler struct {
	crawlers []SearchCrawler
	warnings []string
}

// NewMultiSearchCrawler 创建聚合搜索采集器。
func NewMultiSearchCrawler(crawlers []SearchCrawler) *MultiSearchCrawler {
	return &MultiSearchCrawler{crawlers: crawlers}
}

// Search 依次执行多个来源搜索，允许部分来源失败。
func (c *MultiSearchCrawler) Search(ctx context.Context, query string) ([]model.NewsItem, error) {
	c.warnings = c.warnings[:0]
	result := make([]model.NewsItem, 0)

	for index, crawler := range c.crawlers {
		items, err := crawler.Search(ctx, query)
		if err != nil {
			c.warnings = append(c.warnings, fmt.Sprintf("%s: %v", searchSourceLabel(crawler, index), err))
			continue
		}
		result = append(result, items...)
	}

	return result, nil
}

// Warnings 返回最近一次搜索的部分失败信息。
func (c *MultiSearchCrawler) Warnings() []string {
	return append([]string(nil), c.warnings...)
}

func searchSourceLabel(crawler SearchCrawler, index int) string {
	if named, ok := crawler.(interface{ String() string }); ok {
		return named.String()
	}

	return fmt.Sprintf("search-source-%d", index+1)
}
