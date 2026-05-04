package service

import (
	"testing"
	"time"

	"InfoHub-agent/internal/model"
)

func TestFilterByPreference(t *testing.T) {
	items := []model.NewsItem{
		{Title: "AI 信息", Tags: []string{"AI"}},
		{Title: "数据库信息", Tags: []string{"数据库"}},
	}

	result := FilterByPreference(items, UserPreference{Tags: []string{"AI"}})

	if len(result) != 1 || result[0].Title != "AI 信息" {
		t.Fatalf("偏好过滤结果不符合预期：%+v", result)
	}
}

func TestSortByPreferenceScoreBoostsPreferredTagsAndSources(t *testing.T) {
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	items := []model.NewsItem{
		{
			Title:       "一般数据库更新",
			Content:     "常规变更",
			Tags:        []string{"数据库"},
			SourceName:  "db-feed",
			Score:       4.5,
			PublishTime: now.Add(-1 * time.Hour),
		},
		{
			Title:       "AI Agent 框架发布",
			Content:     "新的 agent 编排能力上线",
			Tags:        []string{"AI", "Agent"},
			SourceName:  "ai-feed",
			Score:       3.2,
			PublishTime: now.Add(-2 * time.Hour),
		},
	}

	result := SortByPreferenceScore(items, UserPreference{
		Tags:     []string{"AI", "Agent"},
		Sources:  []string{"ai-feed"},
		Keywords: []string{"编排"},
	}, now)

	if len(result) != 2 || result[0].Title != "AI Agent 框架发布" {
		t.Fatalf("期望偏好条目排在前面，得到：%+v", result)
	}
}

func TestSortByPreferenceScoreFallsBackWhenPreferenceEmpty(t *testing.T) {
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	items := []model.NewsItem{
		{Title: "较新", Score: 3, PublishTime: now.Add(-1 * time.Hour)},
		{Title: "较旧", Score: 3, PublishTime: now.Add(-2 * time.Hour)},
	}

	result := SortByPreferenceScore(items, UserPreference{}, now)

	if len(result) != 2 || result[0].Title != "较新" {
		t.Fatalf("期望空偏好时回退基础排序，得到：%+v", result)
	}
}
