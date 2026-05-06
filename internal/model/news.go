// Package model 定义信息汇总流程中的核心数据结构。
package model

import "time"

// NewsItem 表示从数据源采集并经过处理的一条信息。
type NewsItem struct {
	ID          int64
	SourceName  string
	Channel     string `json:",omitempty"`
	Title       string
	Content     string
	Source      string
	URL         string
	PublishTime time.Time
	Tags        []string
	Score       float64
	Query       string  `json:",omitempty"`
	SourceScore float64 `json:",omitempty"`
}
