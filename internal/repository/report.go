// Package repository 提供信息和日报的持久化能力。
package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"InfoHub-agent/internal/model"
)

// ErrUserPreferenceNotFound 表示未找到指定用户偏好。
var ErrUserPreferenceNotFound = errors.New("用户偏好不存在")

// ReportRecord 表示一次日报生成结果。
type ReportRecord struct {
	GeneratedAt time.Time
	Markdown    string
	Items       []model.NewsItem
}

// ReportRepository 定义日报存储接口。
type ReportRepository interface {
	Save(ctx context.Context, record ReportRecord) error
	Latest(ctx context.Context) (ReportRecord, error)
	List(ctx context.Context) ([]ReportMetadata, error)
}

// UserPreferenceRecord 表示一个用户的个性化偏好配置。
type UserPreferenceRecord struct {
	UserID    string                `json:"user_id"`
	Tags      []string              `json:"tags"`
	Sources   []string              `json:"sources"`
	Keywords  []string              `json:"keywords"`
	Weights   PreferenceWeightValue `json:"weights"`
	UpdatedAt time.Time             `json:"updated_at"`
}

// PreferenceWeightValue 表示持久化的偏好权重配置。
type PreferenceWeightValue struct {
	Tag     float64 `json:"tag"`
	Source  float64 `json:"source"`
	Keyword float64 `json:"keyword"`
}

// UserPreferenceRepository 定义用户偏好存储接口。
type UserPreferenceRepository interface {
	Save(ctx context.Context, record UserPreferenceRecord) error
	Get(ctx context.Context, userID string) (UserPreferenceRecord, error)
}

// ReportMetadata 表示历史日报的索引信息。
type ReportMetadata struct {
	Name              string    `json:"name"`
	Markdown          string    `json:"markdown"`
	Items             string    `json:"items"`
	ItemCount         int       `json:"item_count"`
	DisplayCount      int       `json:"display_count"`
	HighPriorityCount int       `json:"high_priority_count"`
	TopTitles         []string  `json:"top_titles"`
	CreatedAt         time.Time `json:"created_at"`
}

// CountDisplayItems 统计日报中实际展示的信息条目数量。
func CountDisplayItems(markdown string) int {
	count := 0
	for _, line := range strings.Split(markdown, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "## ") {
			continue
		}
		if strings.HasPrefix(line, "### ") {
			continue
		}
		switch line {
		case "## 今日概览", "## Overview", "## Summary", "## Today's Overview", "## Today Overview", "## Highlights":
			continue
		}
		count++
	}

	return count
}

// CountHighPriorityItems 统计高优先级条目数量。
func CountHighPriorityItems(items []model.NewsItem) int {
	count := 0
	for _, item := range items {
		if item.Score >= 4 {
			count++
		}
	}

	return count
}

// TopTitles 返回前 n 个重点标题。
func TopTitles(items []model.NewsItem, limit int) []string {
	if limit <= 0 {
		limit = 1
	}

	result := make([]string, 0, limit)
	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		if title == "" {
			continue
		}
		result = append(result, title)
		if len(result) == limit {
			break
		}
	}

	return result
}
