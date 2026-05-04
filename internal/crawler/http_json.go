package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"InfoHub-agent/internal/model"
	"InfoHub-agent/internal/processor"
)

// HTTPJSONCrawler 从 HTTP JSON 接口采集结构化信息。
type HTTPJSONCrawler struct {
	url     string
	client  *http.Client
	headers map[string]string
}

// NewHTTPJSONCrawler 创建 HTTP JSON 采集器。
func NewHTTPJSONCrawler(url string, client *http.Client, headers map[string]string) *HTTPJSONCrawler {
	if client == nil {
		client = http.DefaultClient
	}

	return &HTTPJSONCrawler{
		url:     url,
		client:  client,
		headers: headers,
	}
}

// String 返回当前 HTTP JSON source 标识。
func (c *HTTPJSONCrawler) String() string {
	return c.url
}

// Fetch 拉取并解析 HTTP JSON 数据。
func (c *HTTPJSONCrawler) Fetch(ctx context.Context) ([]model.NewsItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return nil, fmt.Errorf("build http_json request for %s: %w", c.url, err)
	}
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request http_json source %s: %w", c.url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("http_json source %s returned status code %d", c.url, resp.StatusCode)
	}

	var payload httpJSONEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode http_json source %s: %w", c.url, err)
	}

	rawItems := payload.Items
	if len(rawItems) == 0 && len(payload.ArrayItems) > 0 {
		rawItems = payload.ArrayItems
	}

	items := make([]model.NewsItem, 0, len(rawItems))
	for index, item := range rawItems {
		publishTime := parseHTTPJSONTime(item.PublishTime)
		items = append(items, model.NewsItem{
			ID:          int64(index + 1),
			Title:       processor.CleanText(item.Title, 200),
			Content:     processor.CleanText(item.Content, 2000),
			Source:      processor.CleanText(item.Source, 100),
			URL:         strings.TrimSpace(item.URL),
			PublishTime: publishTime,
			Tags:        cleanTags(item.Tags),
			Score:       item.Score,
		})
	}

	return items, nil
}

type httpJSONEnvelope struct {
	Items      []httpJSONItem `json:"items"`
	ArrayItems []httpJSONItem `json:"-"`
}

func (e *httpJSONEnvelope) UnmarshalJSON(data []byte) error {
	type alias httpJSONEnvelope
	var object alias
	if err := json.Unmarshal(data, &object); err == nil && len(object.Items) > 0 {
		*e = httpJSONEnvelope(object)
		return nil
	}

	var arrayItems []httpJSONItem
	if err := json.Unmarshal(data, &arrayItems); err != nil {
		return err
	}

	e.ArrayItems = arrayItems
	return nil
}

type httpJSONItem struct {
	Title       string   `json:"title"`
	Content     string   `json:"content"`
	Source      string   `json:"source"`
	URL         string   `json:"url"`
	PublishTime string   `json:"publish_time"`
	Tags        []string `json:"tags"`
	Score       float64  `json:"score"`
}

func parseHTTPJSONTime(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Now()
	}

	for _, layout := range []string{time.RFC3339, time.RFC3339Nano, "2006-01-02 15:04:05"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed
		}
	}

	return time.Now()
}

func cleanTags(tags []string) []string {
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = processor.CleanText(tag, 50)
		if tag == "" {
			continue
		}
		result = append(result, tag)
	}

	return result
}
