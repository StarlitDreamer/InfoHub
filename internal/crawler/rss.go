package crawler

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"InfoHub-agent/internal/model"
	"InfoHub-agent/internal/processor"
)

// RSSCrawler 从 RSS 源采集真实信息。
type RSSCrawler struct {
	url    string
	client *http.Client
}

// NewRSSCrawler 创建 RSS 采集器。
func NewRSSCrawler(url string, client *http.Client) *RSSCrawler {
	if client == nil {
		client = http.DefaultClient
	}

	return &RSSCrawler{url: url, client: client}
}

// Fetch 拉取并解析 RSS 数据。
func (c *RSSCrawler) Fetch() ([]model.NewsItem, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, c.url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("rss 请求失败，状态码：%d", resp.StatusCode)
	}

	var feed rssFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, err
	}

	items := make([]model.NewsItem, 0, len(feed.Channel.Items))
	for index, item := range feed.Channel.Items {
		title := processor.CleanText(item.Title, 200)
		content := processor.CleanText(firstNonEmpty(item.Description, item.Title), 2000)
		source := processor.CleanText(feed.Channel.Title, 100)

		items = append(items, model.NewsItem{
			ID:          int64(index + 1),
			Title:       title,
			Content:     content,
			Source:      source,
			URL:         item.Link,
			PublishTime: parseRSSDate(item.PubDate),
		})
	}

	return items, nil
}

type rssFeed struct {
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title string    `xml:"title"`
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}

func parseRSSDate(value string) time.Time {
	for _, layout := range []string{time.RFC1123Z, time.RFC1123, time.RFC822Z, time.RFC822} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed
		}
	}

	return time.Now()
}
