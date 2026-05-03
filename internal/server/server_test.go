package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"InfoHub-agent/internal/model"
	"InfoHub-agent/internal/repository"
)

func TestHealth(t *testing.T) {
	router := NewRouter(newMemoryRepository(), func(context.Context) (ReportResult, error) {
		return ReportResult{}, nil
	}, Options{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
}

func TestRunReport(t *testing.T) {
	called := false
	router := NewRouter(newMemoryRepository(), func(context.Context) (ReportResult, error) {
		called = true
		return ReportResult{ItemCount: 26, DisplayCount: 12}, nil
	}, Options{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/reports/run", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !called {
		t.Fatal("expected report runner to be called")
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `"item_count":26`) {
		t.Fatalf("expected response to include item_count, got %s", body)
	}
	if !strings.Contains(body, `"display_count":12`) {
		t.Fatalf("expected response to include display_count, got %s", body)
	}
}

func TestLatestReport(t *testing.T) {
	repo := newMemoryRepository()
	_ = repo.Save(context.Background(), repository.ReportRecord{
		GeneratedAt: time.Date(2026, 5, 3, 16, 0, 0, 0, time.UTC),
		Markdown:    "# 今日信息\n\n## ⭐⭐⭐\n- 标题：测试一\n- 摘要：摘要一\n\n## ⭐⭐\n- 标题：测试二\n- 摘要：摘要二\n",
		Items:       []model.NewsItem{{Title: "测试"}},
	})
	router := NewRouter(repo, func(context.Context) (ReportResult, error) {
		return ReportResult{}, nil
	}, Options{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/reports/latest", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), `"display_count":2`) {
		t.Fatalf("expected latest report response to include display_count, got %s", recorder.Body.String())
	}
}

func TestAuthRequiredWhenTokenConfigured(t *testing.T) {
	router := NewRouter(newMemoryRepository(), func(context.Context) (ReportResult, error) {
		return ReportResult{}, nil
	}, Options{AuthToken: "secret"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/reports", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}
}

func TestAuthRejectsWrongToken(t *testing.T) {
	router := NewRouter(newMemoryRepository(), func(context.Context) (ReportResult, error) {
		return ReportResult{}, nil
	}, Options{AuthToken: "secret"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/reports", nil)
	request.Header.Set("Authorization", "Bearer wrong")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}
}

func TestAuthAcceptsBearerToken(t *testing.T) {
	router := NewRouter(newMemoryRepository(), func(context.Context) (ReportResult, error) {
		return ReportResult{}, nil
	}, Options{AuthToken: "secret"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/reports", nil)
	request.Header.Set("Authorization", "Bearer secret")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
}

func TestReportsListIncludesSummaryCounts(t *testing.T) {
	repo := newMemoryRepository()
	repo.records = []repository.ReportRecord{
		{
			GeneratedAt: time.Date(2026, 5, 3, 16, 0, 0, 0, time.UTC),
			Markdown:    "# 今日信息\n\n## ⭐⭐⭐\n- 标题：测试一\n- 摘要：摘要一\n",
			Items:       []model.NewsItem{{Title: "测试一"}, {Title: "库存条目"}},
		},
	}
	router := NewRouter(repo, func(context.Context) (ReportResult, error) {
		return ReportResult{}, nil
	}, Options{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/reports", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `"item_count":2`) || !strings.Contains(body, `"display_count":1`) {
		t.Fatalf("expected reports list to include summary counts, got %s", body)
	}
}

func TestHealthSkipsAuth(t *testing.T) {
	router := NewRouter(newMemoryRepository(), func(context.Context) (ReportResult, error) {
		return ReportResult{}, nil
	}, Options{AuthToken: "secret"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected health to skip auth, got %d", recorder.Code)
	}
}

type memoryRepository struct {
	records []repository.ReportRecord
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{}
}

func (r *memoryRepository) Save(ctx context.Context, record repository.ReportRecord) error {
	r.records = append(r.records, record)
	return nil
}

func (r *memoryRepository) Latest(ctx context.Context) (repository.ReportRecord, error) {
	if len(r.records) == 0 {
		return repository.ReportRecord{}, repository.ErrReportNotFound
	}

	return r.records[len(r.records)-1], nil
}

func (r *memoryRepository) List(ctx context.Context) ([]repository.ReportMetadata, error) {
	result := make([]repository.ReportMetadata, 0, len(r.records))
	for _, record := range r.records {
		displayCount := 0
		for _, line := range strings.Split(record.Markdown, "\n") {
			if strings.HasPrefix(line, "## ") {
				displayCount++
			}
		}

		result = append(result, repository.ReportMetadata{
			Name:         record.GeneratedAt.Format("20060102-150405"),
			ItemCount:    len(record.Items),
			DisplayCount: displayCount,
			CreatedAt:    record.GeneratedAt,
		})
	}

	return result, nil
}
