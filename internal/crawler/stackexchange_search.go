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

var stackExchangeSearchBaseURL = "https://api.stackexchange.com/2.3/search/advanced"

// StackExchangeSearchCrawler 从 Stack Exchange 搜索问题。
type StackExchangeSearchCrawler struct {
	site     string
	maxItems int
	client   *http.Client
	headers  map[string]string
}

// NewStackExchangeSearchCrawler 创建 Stack Exchange 搜索采集器。
func NewStackExchangeSearchCrawler(site string, maxItems int, client *http.Client, headers map[string]string) *StackExchangeSearchCrawler {
	if client == nil {
		client = http.DefaultClient
	}
	if maxItems <= 0 {
		maxItems = 8
	}

	return &StackExchangeSearchCrawler{
		site:     strings.TrimSpace(site),
		maxItems: maxItems,
		client:   client,
		headers:  headers,
	}
}

func (c *StackExchangeSearchCrawler) String() string {
	return "stackexchange:" + c.site
}

// Search 执行关键词搜索。
func (c *StackExchangeSearchCrawler) Search(ctx context.Context, query string) ([]model.NewsItem, error) {
	params := url.Values{}
	params.Set("order", "desc")
	params.Set("sort", "relevance")
	params.Set("q", strings.TrimSpace(query))
	params.Set("site", firstNonEmptySearch(c.site, "stackoverflow"))
	params.Set("pagesize", fmt.Sprintf("%d", c.maxItems))
	params.Set("filter", "withbody")

	endpoint := stackExchangeSearchBaseURL + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build stackexchange request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "InfoHub-Agent/1.0")
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request stackexchange search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("stackexchange search returned status code %d", resp.StatusCode)
	}

	var payload stackExchangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode stackexchange search response: %w", err)
	}

	items := make([]model.NewsItem, 0, len(payload.Items))
	for index, item := range payload.Items {
		items = append(items, model.NewsItem{
			ID:          int64(index + 1),
			Channel:     "stack_overflow",
			Title:       processor.CleanText(item.Title, 200),
			Content:     processor.CleanText(item.Body, 2000),
			Source:      "Stack Overflow",
			URL:         strings.TrimSpace(item.Link),
			PublishTime: time.Unix(item.CreationDate, 0).UTC(),
			SourceScore: float64(item.Score),
		})
	}

	return items, nil
}

type stackExchangeResponse struct {
	Items []stackExchangeItem `json:"items"`
}

type stackExchangeItem struct {
	Title        string `json:"title"`
	Body         string `json:"body"`
	Link         string `json:"link"`
	Score        int    `json:"score"`
	CreationDate int64  `json:"creation_date"`
}

func firstNonEmptySearch(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}
