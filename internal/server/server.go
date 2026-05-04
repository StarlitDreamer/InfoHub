// Package server 提供 Gin HTTP API。
package server

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"InfoHub-agent/internal/model"
	"InfoHub-agent/internal/repository"
	"InfoHub-agent/internal/service"
	"InfoHub-agent/internal/summary"
)

// ReportRunner 表示可被 HTTP 触发的日报生成任务。
type ReportRunner func(context.Context, RunReportRequest) (ReportResult, error)

// RunReportRequest 表示 HTTP 触发日报生成时的可选参数。
type RunReportRequest struct {
	UserID     string            `json:"user_id"`
	Preference PreferenceRequest `json:"preference"`
}

// PreferenceRequest 表示一次请求的个性化偏好参数。
type PreferenceRequest struct {
	Tags     []string `json:"tags"`
	Sources  []string `json:"sources"`
	Keywords []string `json:"keywords"`
	Weights  struct {
		Tag     float64 `json:"tag"`
		Source  float64 `json:"source"`
		Keyword float64 `json:"keyword"`
	} `json:"weights"`
}

// ReportResult 表示一次日报生成结果摘要。
type ReportResult struct {
	ItemCount         int                     `json:"item_count"`
	DisplayCount      int                     `json:"display_count"`
	GeneratedAt       time.Time               `json:"generated_at"`
	HighPriorityCount int                     `json:"high_priority_count"`
	TopPriorityItems  []string                `json:"top_priority_items"`
	DecisionSummary   []reportDecisionSummary `json:"decision_summary"`
}

// Options 保存 HTTP 服务选项。
type Options struct {
	AuthToken          string
	UserPreferenceRepo repository.UserPreferenceRepository
}

type reportDecisionSummary struct {
	Title   string   `json:"title"`
	Source  string   `json:"source"`
	Score   float64  `json:"score"`
	Tags    []string `json:"tags"`
	Action  string   `json:"action"`
	Summary string   `json:"summary"`
}

type reportOverview struct {
	DisplayCount      int                     `json:"display_count"`
	HighPriorityCount int                     `json:"high_priority_count"`
	TopPriorityItems  []string                `json:"top_priority_items"`
	DecisionSummary   []reportDecisionSummary `json:"decision_summary"`
}

// NewRouter 创建 HTTP 路由。
func NewRouter(repo repository.ReportRepository, runner ReportRunner, options Options) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	protected := router.Group("")
	protected.Use(authMiddleware(options.AuthToken))
	preferenceRepo := options.UserPreferenceRepo

	protected.POST("/reports/run", func(ctx *gin.Context) {
		var request RunReportRequest
		if ctx.Request.ContentLength > 0 {
			if err := ctx.ShouldBindJSON(&request); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
				return
			}
		}

		result, err := runner(ctx.Request.Context(), request)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"status":              "generated",
			"item_count":          result.ItemCount,
			"display_count":       result.DisplayCount,
			"generated_at":        result.GeneratedAt,
			"high_priority_count": result.HighPriorityCount,
			"top_priority_items":  result.TopPriorityItems,
			"decision_summary":    result.DecisionSummary,
		})
	})

	protected.GET("/preferences/:userID", func(ctx *gin.Context) {
		if preferenceRepo == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "preference repository not configured"})
			return
		}

		record, err := preferenceRepo.Get(ctx.Request.Context(), ctx.Param("userID"))
		if err != nil {
			if errors.Is(err, repository.ErrUserPreferenceNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "user preference not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, record)
	})

	protected.PUT("/preferences/:userID", func(ctx *gin.Context) {
		if preferenceRepo == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "preference repository not configured"})
			return
		}

		var request PreferenceRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		record := repository.UserPreferenceRecord{
			UserID:   ctx.Param("userID"),
			Tags:     append([]string(nil), request.Tags...),
			Sources:  append([]string(nil), request.Sources...),
			Keywords: append([]string(nil), request.Keywords...),
			Weights: repository.PreferenceWeightValue{
				Tag:     request.Weights.Tag,
				Source:  request.Weights.Source,
				Keyword: request.Weights.Keyword,
			},
		}
		if err := preferenceRepo.Save(ctx.Request.Context(), record); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		saved, err := preferenceRepo.Get(ctx.Request.Context(), ctx.Param("userID"))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, saved)
	})

	protected.GET("/reports/latest", func(ctx *gin.Context) {
		record, err := repo.Latest(ctx.Request.Context())
		if err != nil {
			if errors.Is(err, repository.ErrReportNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		overview := buildReportOverview(record.Markdown, record.Items, 3)
		ctx.JSON(http.StatusOK, gin.H{
			"generated_at":        record.GeneratedAt,
			"markdown":            record.Markdown,
			"items":               record.Items,
			"display_count":       overview.DisplayCount,
			"high_priority_count": overview.HighPriorityCount,
			"decision_summary":    overview.DecisionSummary,
			"top_priority_items":  overview.TopPriorityItems,
		})
	})

	protected.GET("/reports", func(ctx *gin.Context) {
		records, err := repo.List(ctx.Request.Context())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"reports": records})
	})

	return router
}

func authMiddleware(token string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if token == "" {
			ctx.Next()
			return
		}

		header := ctx.GetHeader("Authorization")
		value, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || value != token {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		ctx.Next()
	}
}

// ToUserPreference 转换为 service 层使用的偏好结构。
func (r PreferenceRequest) ToUserPreference() service.UserPreference {
	return service.UserPreference{
		Tags:     append([]string(nil), r.Tags...),
		Sources:  append([]string(nil), r.Sources...),
		Keywords: append([]string(nil), r.Keywords...),
		Weights: service.PreferenceWeights{
			TagMatch:     r.Weights.Tag,
			SourceMatch:  r.Weights.Source,
			KeywordMatch: r.Weights.Keyword,
		},
	}
}

func buildReportOverview(markdown string, items []model.NewsItem, limit int) reportOverview {
	base := repository.BuildReportOverview(markdown, items, limit)
	return reportOverview{
		DisplayCount:      base.DisplayCount,
		HighPriorityCount: base.HighPriorityCount,
		TopPriorityItems:  base.TopTitles,
		DecisionSummary:   buildDecisionSummary(items, limit),
	}
}

// BuildReportResult 根据一次执行结果构建对外返回的轻量摘要。
func BuildReportResult(itemCount, displayCount int, generatedAt time.Time, markdown string, items []model.NewsItem, limit int) ReportResult {
	overview := buildReportOverview(markdown, items, limit)
	return ReportResult{
		ItemCount:         itemCount,
		DisplayCount:      displayCount,
		GeneratedAt:       generatedAt,
		HighPriorityCount: overview.HighPriorityCount,
		TopPriorityItems:  overview.TopPriorityItems,
		DecisionSummary:   overview.DecisionSummary,
	}
}

func buildDecisionSummary(items []model.NewsItem, limit int) []reportDecisionSummary {
	if limit <= 0 {
		limit = 1
	}
	if len(items) > limit {
		items = items[:limit]
	}

	result := make([]reportDecisionSummary, 0, len(items))
	for _, item := range items {
		parsed := summary.Parse(item)
		action := summary.RecommendAction(item, parsed)
		result = append(result, reportDecisionSummary{
			Title:   parsed.Title,
			Source:  normalizeSource(item),
			Score:   item.Score,
			Tags:    append([]string(nil), item.Tags...),
			Action:  action.Label,
			Summary: parsed.WhatHappened,
		})
	}

	return result
}

func normalizeSource(item model.NewsItem) string {
	source := strings.TrimSpace(item.Source)
	if source != "" {
		return source
	}
	sourceName := strings.TrimSpace(item.SourceName)
	if sourceName != "" {
		return sourceName
	}

	return "未知来源"
}
