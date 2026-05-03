// Package server 提供 Gin HTTP API。
package server

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"InfoHub-agent/internal/repository"
)

// ReportRunner 表示可被 HTTP 触发的日报生成任务。
type ReportRunner func(context.Context) (ReportResult, error)

// ReportResult 表示一次日报生成结果摘要。
type ReportResult struct {
	ItemCount int `json:"item_count"`
}

// Options 保存 HTTP 服务选项。
type Options struct {
	AuthToken string
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

	protected.POST("/reports/run", func(ctx *gin.Context) {
		result, err := runner(ctx.Request.Context())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"status":     "generated",
			"item_count": result.ItemCount,
		})
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
			"generated_at": record.GeneratedAt,
			"markdown":     record.Markdown,
			"items":        record.Items,
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
