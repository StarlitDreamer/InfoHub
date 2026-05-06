package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"InfoHub-agent/internal/model"
	"InfoHub-agent/internal/processor"
)

var redditSearchBaseURL = "https://www.reddit.com"

// RedditSearchCrawler 从 Reddit 搜索公开帖子。
type RedditSearchCrawler struct {
	community string
	maxItems  int
	client    *http.Client
	headers   map[string]string
}

// NewRedditSearchCrawler 创建 Reddit 搜索采集器。
func NewRedditSearchCrawler(community string, maxItems int, client *http.Client, headers map[string]string) *RedditSearchCrawler {
	if client == nil {
		client = http.DefaultClient
	}
	if maxItems <= 0 {
		maxItems = 8
	}

	return &RedditSearchCrawler{
		community: strings.TrimSpace(community),
		maxItems:  maxItems,
		client:    client,
		headers:   headers,
	}
}

func (c *RedditSearchCrawler) String() string {
	if c.community == "" {
		return "reddit"
	}
	return "reddit:" + c.community
}

// Search 执行关键词搜索。
func (c *RedditSearchCrawler) Search(ctx context.Context, query string) ([]model.NewsItem, error) {
	params := url.Values{}
	params.Set("q", strings.TrimSpace(query))
	params.Set("sort", "relevance")
	params.Set("limit", fmt.Sprintf("%d", c.maxItems))
	params.Set("restrict_sr", "on")
	params.Set("raw_json", "1")

	endpoint := strings.TrimRight(redditSearchBaseURL, "/") + "/search.json?" + params.Encode()
	if c.community != "" {
		endpoint = fmt.Sprintf("%s/r/%s/search.json?%s", strings.TrimRight(redditSearchBaseURL, "/"), url.PathEscape(c.community), params.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build reddit request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "InfoHub-Agent/1.0")
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request reddit search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("reddit search returned status code %d", resp.StatusCode)
	}

	var payload redditSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode reddit search response: %w", err)
	}

	items := make([]model.NewsItem, 0, len(payload.Data.Children))
	for index, child := range payload.Data.Children {
		items = append(items, model.NewsItem{
			ID:          int64(index + 1),
			Channel:     "reddit",
			Title:       processor.CleanText(child.Data.Title, 200),
			Content:     processor.CleanText(child.Data.SelfText, 2000),
			Source:      "Reddit",
			URL:         buildRedditURL(child.Data.Permalink),
			PublishTime: time.Unix(int64(child.Data.CreatedUTC), 0).UTC(),
			SourceScore: child.Data.Score,
		})
	}

	return items, nil
}

type redditSearchResponse struct {
	Data struct {
		Children []struct {
			Data struct {
				Title      string  `json:"title"`
				SelfText   string  `json:"selftext"`
				Permalink  string  `json:"permalink"`
				CreatedUTC float64 `json:"created_utc"`
				Score      float64 `json:"score"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

func buildRedditURL(permalink string) string {
	permalink = strings.TrimSpace(permalink)
	if permalink == "" {
		return ""
	}
	if strings.HasPrefix(permalink, "http://") || strings.HasPrefix(permalink, "https://") {
		return permalink
	}
	return "https://www.reddit.com" + permalink
}
