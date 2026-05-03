package service

import (
	"testing"

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
