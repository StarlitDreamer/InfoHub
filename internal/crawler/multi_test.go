package crawler

import (
	"errors"
	"testing"

	"InfoHub-agent/internal/model"
)

func TestMultiCrawlerFetchesAllSuccessfulSources(t *testing.T) {
	crawler := NewMultiCrawler([]Crawler{
		staticCrawler{items: []model.NewsItem{{Title: "第一条"}}},
		staticCrawler{items: []model.NewsItem{{Title: "第二条"}}},
	})

	items, err := crawler.Fetch()
	if err != nil {
		t.Fatalf("多源采集失败：%v", err)
	}

	if len(items) != 2 {
		t.Fatalf("期望采集 2 条数据，实际为 %d", len(items))
	}
}

func TestMultiCrawlerKeepsSuccessfulSourcesWhenSomeFail(t *testing.T) {
	crawler := NewMultiCrawler([]Crawler{
		errorCrawler{err: errors.New("采集失败")},
		staticCrawler{items: []model.NewsItem{{Title: "成功数据"}}},
	})

	items, err := crawler.Fetch()
	if err != nil {
		t.Fatalf("部分失败时不应中断全部采集：%v", err)
	}

	if len(items) != 1 || items[0].Title != "成功数据" {
		t.Fatalf("部分失败采集结果不符合预期：%+v", items)
	}
}

func TestMultiCrawlerReturnsErrorWhenAllSourcesFail(t *testing.T) {
	crawler := NewMultiCrawler([]Crawler{
		errorCrawler{err: errors.New("采集失败")},
	})

	if _, err := crawler.Fetch(); err == nil {
		t.Fatal("期望所有数据源失败时返回错误")
	}
}

type staticCrawler struct {
	items []model.NewsItem
}

func (c staticCrawler) Fetch() ([]model.NewsItem, error) {
	return c.items, nil
}

type errorCrawler struct {
	err error
}

func (c errorCrawler) Fetch() ([]model.NewsItem, error) {
	return nil, c.err
}
