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

// FactoryOptions 保存构建采集器时共享的选项。
type FactoryOptions struct {
	RSSMaxItems     int
	RSSRecentWithin time.Duration
}

// BuildFromSources 根据 source 配置构建采集器。
func BuildFromSources(sources []config.SourceConfig, options FactoryOptions) (Crawler, error) {
	crawlers := make([]Crawler, 0, len(sources))

	for _, source := range sources {
		crawler, err := buildSingleSource(source, options)
		if err != nil {
			return nil, err
		}
		if crawler == nil {
			continue
		}

		crawlers = append(crawlers, wrapSourceCrawler(source, crawler))
	}

	if len(crawlers) == 0 {
		return NewDemoCrawler(), nil
	}
	if len(crawlers) == 1 {
		return crawlers[0], nil
	}

	return NewMultiCrawler(crawlers), nil
}

type sourceCrawler struct {
	source  config.SourceConfig
	crawler Crawler
}

func wrapSourceCrawler(source config.SourceConfig, crawler Crawler) Crawler {
	return sourceCrawler{
		source:  source,
		crawler: crawler,
	}
}

func (c sourceCrawler) Fetch(ctx context.Context) ([]model.NewsItem, error) {
	items, err := c.crawler.Fetch(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.NewsItem, len(items))
	for index, item := range items {
		item.SourceName = c.source.Name
		result[index] = item
	}

	return result, nil
}

func (c sourceCrawler) String() string {
	if named, ok := c.crawler.(interface{ String() string }); ok {
		return named.String()
	}

	return c.source.Name
}

func buildSingleSource(source config.SourceConfig, options FactoryOptions) (Crawler, error) {
	if !source.Enabled {
		return nil, nil
	}

	kind := strings.ToLower(strings.TrimSpace(source.Kind))
	location := strings.TrimSpace(source.Location)
	client := httpClientForSource(source)

	switch kind {
	case "rss":
		if location == "" {
			return nil, fmt.Errorf("rss source %q requires location", source.Name)
		}
		return NewRSSCrawler(location, client, source.Headers, RSSOptions{
			MaxItems:     options.RSSMaxItems,
			RecentWithin: options.RSSRecentWithin,
		}), nil
	case "http_json":
		if location == "" {
			return nil, fmt.Errorf("http_json source %q requires location", source.Name)
		}
		return NewHTTPJSONCrawler(location, client, source.Headers), nil
	case "demo":
		return NewDemoCrawler(), nil
	case "":
		return nil, fmt.Errorf("source %q requires kind", source.Name)
	default:
		return nil, fmt.Errorf("unsupported source kind: %s", source.Kind)
	}
}

func httpClientForSource(source config.SourceConfig) *http.Client {
	if source.TimeoutSeconds <= 0 {
		return nil
	}

	return &http.Client{
		Timeout: time.Duration(source.TimeoutSeconds) * time.Second,
	}
}
