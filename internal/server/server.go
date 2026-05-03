// Package server 提供 Gin HTTP API。
package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"InfoHub-agent/internal/repository"
)

// ReportRunner 表示可被 HTTP 触发的日报生成任务。
type ReportRunner func(context.Context) error

// NewRouter 创建 HTTP 路由。
func NewRouter(repo repository.ReportRepository, runner ReportRunner) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.POST("/reports/run", func(ctx *gin.Context) {
		if err := runner(ctx.Request.Context()); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "generated"})
	})

	router.GET("/reports/latest", func(ctx *gin.Context) {
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

	router.GET("/reports", func(ctx *gin.Context) {
		records, err := repo.List(ctx.Request.Context())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"reports": records})
	})

	return router
}
