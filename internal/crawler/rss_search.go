package crawler

import (
	"context"
	"net/http"
	"strings"

	"InfoHub-agent/internal/model"
)

// RSSSearchCrawler 对 RSS 源做关键词过滤搜索。
type RSSSearchCrawler struct {
	feedURL  string
	maxItems int
	rss      *RSSCrawler
}

// NewRSSSearchCrawler 创建 RSS 搜索采集器。
func NewRSSSearchCrawler(feedURL string, maxItems int, client *http.Client, headers map[string]string, options RSSOptions) *RSSSearchCrawler {
	if maxItems <= 0 {
		maxItems = 8
	}
	return &RSSSearchCrawler{
		feedURL:  strings.TrimSpace(feedURL),
		maxItems: maxItems,
		rss:      NewRSSCrawler(feedURL, client, headers, options),
	}
}

func (c *RSSSearchCrawler) String() string {
	return c.feedURL
}

// Search 拉取 RSS 内容后按关键词过滤。
func (c *RSSSearchCrawler) Search(ctx context.Context, query string) ([]model.NewsItem, error) {
	items, err := c.rss.Fetch(ctx)
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return items, nil
	}

	limit := c.maxItems
	if len(items) < limit {
		limit = len(items)
	}
	result := make([]model.NewsItem, 0, limit)
	for _, item := range items {
		text := strings.ToLower(item.Title + "\n" + item.Content)
		if !strings.Contains(text, query) {
			continue
		}
		item.Channel = "rss_search"
		result = append(result, item)
		if len(result) == c.maxItems {
			break
		}
	}

	return result, nil
}
