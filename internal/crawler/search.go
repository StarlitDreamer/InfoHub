package crawler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"InfoHub-agent/internal/config"
	"InfoHub-agent/internal/model"
)

// SearchCrawler 定义按关键词搜索的数据源接口。
type SearchCrawler interface {
	Search(ctx context.Context, query string) ([]model.NewsItem, error)
}

// SearchFactoryOptions 保存构建搜索采集器时的共享选项。
type SearchFactoryOptions struct {
	DefaultMaxItems int
	RSSRecentWithin time.Duration
}

// BuildSearchFromSources 根据搜索数据源配置构建采集器。
func BuildSearchFromSources(sources []config.SearchSourceConfig, options SearchFactoryOptions) (SearchCrawler, error) {
	crawlers := make([]SearchCrawler, 0, len(sources))
	for _, source := range sources {
		crawler, err := buildSearchSource(source, options)
		if err != nil {
			return nil, err
		}
		if crawler == nil {
			continue
		}
		crawlers = append(crawlers, wrapSearchSource(source, crawler))
	}

	if len(crawlers) == 0 {
		return NewMultiSearchCrawler(nil), nil
	}
	if len(crawlers) == 1 {
		return crawlers[0], nil
	}

	return NewMultiSearchCrawler(crawlers), nil
}

type sourceSearchCrawler struct {
	source  config.SearchSourceConfig
	crawler SearchCrawler
}

func wrapSearchSource(source config.SearchSourceConfig, crawler SearchCrawler) SearchCrawler {
	return sourceSearchCrawler{
		source:  source,
		crawler: crawler,
	}
}

func (c sourceSearchCrawler) Search(ctx context.Context, query string) ([]model.NewsItem, error) {
	items, err := c.crawler.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	result := make([]model.NewsItem, len(items))
	for index, item := range items {
		item.SourceName = c.source.Name
		if item.Channel == "" {
			item.Channel = c.source.Kind
		}
		item.Query = strings.TrimSpace(query)
		result[index] = item
	}

	return result, nil
}

func (c sourceSearchCrawler) String() string {
	if named, ok := c.crawler.(interface{ String() string }); ok {
		return named.String()
	}

	return c.source.Name
}

func buildSearchSource(source config.SearchSourceConfig, options SearchFactoryOptions) (SearchCrawler, error) {
	if !source.Enabled {
		return nil, nil
	}

	kind := strings.ToLower(strings.TrimSpace(source.Kind))
	location := strings.TrimSpace(source.Location)
	client := searchHTTPClientForSource(source)
	maxItems := source.MaxItems
	if maxItems <= 0 {
		maxItems = options.DefaultMaxItems
	}
	if maxItems <= 0 {
		maxItems = 8
	}

	switch kind {
	case "stackexchange":
		if location == "" {
			location = "stackoverflow"
		}
		return NewStackExchangeSearchCrawler(location, maxItems, client, source.Headers), nil
	case "reddit":
		return NewRedditSearchCrawler(location, maxItems, client, source.Headers), nil
	case "rss_search":
		if location == "" {
			return nil, fmt.Errorf("rss_search source %q requires location", source.Name)
		}
		return NewRSSSearchCrawler(location, maxItems, client, source.Headers, RSSOptions{
			MaxItems:     maxItems,
			RecentWithin: options.RSSRecentWithin,
		}), nil
	case "":
		return nil, fmt.Errorf("search source %q requires kind", source.Name)
	default:
		return nil, fmt.Errorf("unsupported search source kind: %s", source.Kind)
	}
}

func searchHTTPClientForSource(source config.SearchSourceConfig) *http.Client {
	if source.TimeoutSeconds <= 0 {
		return nil
	}

	return &http.Client{Timeout: time.Duration(source.TimeoutSeconds) * time.Second}
}
