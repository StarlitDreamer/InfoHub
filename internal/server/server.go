// Package server 提供 Gin HTTP API。
package server

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"InfoHub-agent/internal/model"
	"InfoHub-agent/internal/repository"
	"InfoHub-agent/internal/service"
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
	ItemCount    int `json:"item_count"`
	DisplayCount int `json:"display_count"`
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
			"status":        "generated",
			"item_count":    result.ItemCount,
			"display_count": result.DisplayCount,
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

		ctx.JSON(http.StatusOK, gin.H{
			"generated_at":       record.GeneratedAt,
			"markdown":           record.Markdown,
			"items":              record.Items,
			"display_count":      repository.CountDisplayItems(record.Markdown),
			"decision_summary":   buildDecisionSummary(record.Items, 3),
			"top_priority_items": repository.TopTitles(record.Items, 3),
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

func buildDecisionSummary(items []model.NewsItem, limit int) []reportDecisionSummary {
	if limit <= 0 {
		limit = 1
	}
	if len(items) > limit {
		items = items[:limit]
	}

	result := make([]reportDecisionSummary, 0, len(items))
	for _, item := range items {
		result = append(result, reportDecisionSummary{
			Title:   strings.TrimSpace(item.Title),
			Source:  normalizeSource(item),
			Score:   item.Score,
			Tags:    append([]string(nil), item.Tags...),
			Action:  summarizeAction(item),
			Summary: summarizeWhatHappened(item),
		})
	}

	return result
}

func summarizeAction(item model.NewsItem) string {
	score := item.Score
	text := strings.ToLower(strings.Join(item.Tags, " ") + " " + item.Title + " " + item.Content)

	switch {
	case score >= 5:
		return "立即评审"
	case score >= 4:
		return "近期跟进"
	case strings.Contains(text, "security") || strings.Contains(text, "安全") || strings.Contains(text, "cyber"):
		return "安全评估"
	case strings.Contains(text, "database") || strings.Contains(text, "数据库") || strings.Contains(text, "index"):
		return "专项验证"
	case strings.Contains(text, "ai") || strings.Contains(text, "agent") || strings.Contains(text, "模型"):
		return "小范围试用"
	default:
		return "持续观察"
	}
}

func summarizeWhatHappened(item model.NewsItem) string {
	for _, rawLine := range strings.Split(item.Content, "\n") {
		line := strings.TrimSpace(rawLine)
		if strings.HasPrefix(line, "【发生了什么】") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "【发生了什么】"))
			if value != "" {
				return value
			}
		}
	}

	content := strings.TrimSpace(item.Content)
	if content == "" {
		return strings.TrimSpace(item.Title)
	}

	lines := strings.Split(content, "\n")
	return strings.TrimSpace(lines[0])
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
