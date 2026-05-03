package crawler

import (
	"fmt"
	"strings"
	"time"

	"InfoHub-agent/internal/config"
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

		crawlers = append(crawlers, crawler)
	}

	if len(crawlers) == 0 {
		return NewDemoCrawler(), nil
	}
	if len(crawlers) == 1 {
		return crawlers[0], nil
	}

	return NewMultiCrawler(crawlers), nil
}

func buildSingleSource(source config.SourceConfig, options FactoryOptions) (Crawler, error) {
	kind := strings.ToLower(strings.TrimSpace(source.Kind))
	location := strings.TrimSpace(source.Location)

	switch kind {
	case "rss":
		if location == "" {
			return nil, fmt.Errorf("rss source %q requires location", source.Name)
		}
		return NewRSSCrawler(location, nil, RSSOptions{
			MaxItems:     options.RSSMaxItems,
			RecentWithin: options.RSSRecentWithin,
		}), nil
	case "demo":
		return NewDemoCrawler(), nil
	case "":
		return nil, fmt.Errorf("source %q requires kind", source.Name)
	default:
		return nil, fmt.Errorf("unsupported source kind: %s", source.Kind)
	}
}
