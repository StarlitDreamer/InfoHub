package crawler

import (
	"fmt"
	"strings"

	"InfoHub-agent/internal/model"
)

// MultiCrawler 聚合多个数据采集器。
type MultiCrawler struct {
	crawlers []Crawler
}

// NewMultiCrawler 创建多源聚合采集器。
func NewMultiCrawler(crawlers []Crawler) *MultiCrawler {
	return &MultiCrawler{crawlers: crawlers}
}

// Fetch 依次采集多个来源，部分失败时保留成功来源的数据。
func (c *MultiCrawler) Fetch() ([]model.NewsItem, error) {
	var messages []string
	items := make([]model.NewsItem, 0)

	for index, crawler := range c.crawlers {
		fetched, err := crawler.Fetch()
		if err != nil {
			messages = append(messages, fmt.Sprintf("%s: %v", sourceLabel(crawler, index), err))
			continue
		}

		items = append(items, fetched...)
	}

	if len(items) == 0 && len(messages) > 0 {
		return nil, fmt.Errorf("all sources failed: %s", strings.Join(messages, "; "))
	}

	return items, nil
}

func sourceLabel(crawler Crawler, index int) string {
	if named, ok := crawler.(interface{ String() string }); ok {
		label := strings.TrimSpace(named.String())
		if label != "" {
			return label
		}
	}

	return fmt.Sprintf("source-%d", index+1)
}
