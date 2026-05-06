package server

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"InfoHub-agent/internal/model"
	"InfoHub-agent/internal/repository"
)

func TestHealth(t *testing.T) {
	router := NewRouter(newMemoryRepository(), func(context.Context, RunReportRequest) (ReportResult, error) {
		return ReportResult{}, nil
	}, Options{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
}

func TestIndexPage(t *testing.T) {
	router := NewRouter(newMemoryRepository(), func(context.Context, RunReportRequest) (ReportResult, error) {
		return ReportResult{}, nil
	}, Options{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "InfoHub Agent") || !strings.Contains(body, "信息日报工作台") {
		t.Fatalf("expected index page content, got %s", body)
	}
}

func TestRunReport(t *testing.T) {
	called := false
	router := NewRouter(newMemoryRepository(), func(context.Context, RunReportRequest) (ReportResult, error) {
		called = true
		return ReportResult{
			ItemCount:         26,
			DisplayCount:      12,
			GeneratedAt:       time.Date(2026, 5, 4, 8, 0, 0, 0, time.UTC),
			HighPriorityCount: 3,
			TopPriorityItems:  []string{"alpha"},
			DecisionSummary: []reportDecisionSummary{
				{Title: "alpha", Action: "立即评审", Summary: "summary"},
			},
		}, nil
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
	if !strings.Contains(body, `"top_priority_items":["alpha"]`) {
		t.Fatalf("expected response to include run summary, got %s", body)
	}
}

func TestSearch(t *testing.T) {
	searchRepo := newMemorySearchRepository()
	called := false
	router := NewRouter(newMemoryRepository(), func(context.Context, RunReportRequest) (ReportResult, error) {
		return ReportResult{}, nil
	}, Options{
		SearchRepo: searchRepo,
		SearchRunner: func(context.Context, SearchRequest) (SearchResult, error) {
			called = true
			return SearchResult{
				Query:             "agent",
				ItemCount:         2,
				DisplayCount:      2,
				GeneratedAt:       time.Date(2026, 5, 6, 8, 0, 0, 0, time.UTC),
				Warnings:          []string{"reddit: timeout"},
				HighPriorityCount: 1,
				TopPriorityItems:  []string{"Agent orchestration"},
				Items:             []model.NewsItem{{Title: "Agent orchestration", Score: 4}},
				Markdown:          "# search",
			}, nil
		},
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/search", bytes.NewBufferString(`{"query":"agent"}`))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !called {
		t.Fatal("expected search runner to be called")
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `"query":"agent"`) || !strings.Contains(body, `"warnings":["reddit: timeout"]`) {
		t.Fatalf("expected search response payload, got %s", body)
	}
}

func TestRunReportPassesPreferenceRequest(t *testing.T) {
	var captured RunReportRequest
	router := NewRouter(newMemoryRepository(), func(_ context.Context, request RunReportRequest) (ReportResult, error) {
		captured = request
		return ReportResult{ItemCount: 1, DisplayCount: 1}, nil
	}, Options{})
	body := bytes.NewBufferString(`{"preference":{"tags":["AI"],"sources":["openai-news"],"keywords":["agent"]}}`)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/reports/run", body)
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if len(captured.Preference.Tags) != 1 || captured.Preference.Tags[0] != "AI" {
		t.Fatalf("expected preference tags to be parsed, got %+v", captured)
	}
	if captured.UserID != "" {
		t.Fatalf("expected empty user id by default, got %+v", captured)
	}
}

func TestRunReportPassesUserID(t *testing.T) {
	var captured RunReportRequest
	router := NewRouter(newMemoryRepository(), func(_ context.Context, request RunReportRequest) (ReportResult, error) {
		captured = request
		return ReportResult{ItemCount: 1, DisplayCount: 1}, nil
	}, Options{})
	body := bytes.NewBufferString(`{"user_id":"alice"}`)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/reports/run", body)
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if captured.UserID != "alice" {
		t.Fatalf("expected user id alice, got %+v", captured)
	}
}

func TestRunReportRejectsInvalidBody(t *testing.T) {
	router := NewRouter(newMemoryRepository(), func(context.Context, RunReportRequest) (ReportResult, error) {
		return ReportResult{}, nil
	}, Options{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/reports/run", bytes.NewBufferString(`{"preference":`))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
}

func TestLatestReport(t *testing.T) {
	repo := newMemoryRepository()
	_ = repo.Save(context.Background(), repository.ReportRecord{
		GeneratedAt: time.Date(2026, 5, 3, 16, 0, 0, 0, time.UTC),
		Markdown: "# 今日信息日报\n\n## 今日概览\n- 收录条目：2\n\n## ⭐⭐ 测试一\n- 标题：测试一\n" +
			"\n## ⭐ 测试二\n- 标题：测试二\n",
		Items: []model.NewsItem{
			{Title: "测试一", Source: "OpenAI News", Score: 4, Tags: []string{"AI"}, Content: "【发生了什么】summary one"},
			{Title: "测试二", Source: "Google Blog", Score: 2, Content: "plain summary two"},
		},
	})
	router := NewRouter(repo, func(context.Context, RunReportRequest) (ReportResult, error) {
		return ReportResult{}, nil
	}, Options{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/reports/latest", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `"display_count":2`) || !strings.Contains(body, `"high_priority_count":1`) {
		t.Fatalf("expected latest report response to include overview counts, got %s", body)
	}
	if !strings.Contains(body, `"decision_summary"`) || !strings.Contains(body, `"top_priority_items"`) {
		t.Fatalf("expected latest report response to include summary fields, got %s", body)
	}
	if !strings.Contains(body, `"action":"近期跟进"`) || !strings.Contains(body, `"summary":"summary one"`) {
		t.Fatalf("expected latest report response to include decision summary details, got %s", body)
	}
}

func TestBuildReportOverviewUsesSharedSummaryRules(t *testing.T) {
	overview := buildReportOverview(
		"# 今日信息日报\n\n## 今日概览\n- 收录条目：2\n\n## ⭐⭐ 测试一\nbody\n\n## ⭐ 测试二\nbody\n",
		[]model.NewsItem{
			{Title: "测试一", Score: 5, Content: "【发生了什么】summary one"},
			{Title: "测试二", Score: 2, Content: "plain summary two"},
		},
		2,
	)

	if overview.DisplayCount != 2 || overview.HighPriorityCount != 1 {
		t.Fatalf("expected shared overview counts, got %+v", overview)
	}
	if len(overview.TopPriorityItems) != 2 || overview.TopPriorityItems[0] != "测试一" {
		t.Fatalf("expected shared top titles, got %+v", overview.TopPriorityItems)
	}
	if len(overview.DecisionSummary) != 2 || overview.DecisionSummary[0].Action != "立即评审" {
		t.Fatalf("expected shared decision summary, got %+v", overview.DecisionSummary)
	}
}

func TestLatestReportReturnsNotFoundWhenRepositoryIsEmpty(t *testing.T) {
	router := NewRouter(newMemoryRepository(), func(context.Context, RunReportRequest) (ReportResult, error) {
		return ReportResult{}, nil
	}, Options{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/reports/latest", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}
}

func TestGetReportByName(t *testing.T) {
	repo := newMemoryRepository()
	record := repository.ReportRecord{
		GeneratedAt: time.Date(2026, 5, 3, 16, 0, 0, 0, time.UTC),
		Markdown:    "# report\n\n## item\nbody\n",
		Items:       []model.NewsItem{{Title: "item", Score: 4, Content: "summary"}},
	}
	_ = repo.Save(context.Background(), record)

	router := NewRouter(repo, func(context.Context, RunReportRequest) (ReportResult, error) {
		return ReportResult{}, nil
	}, Options{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/reports/20260503-160000", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `"markdown":"# report\n\n## item\nbody\n"`) {
		t.Fatalf("expected report markdown in body, got %s", body)
	}
}

func TestGetReportByNameReturnsNotFound(t *testing.T) {
	router := NewRouter(newMemoryRepository(), func(context.Context, RunReportRequest) (ReportResult, error) {
		return ReportResult{}, nil
	}, Options{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/reports/20260503-160000", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}
}

func TestRunReportReturnsErrorWhenRunnerFails(t *testing.T) {
	router := NewRouter(newMemoryRepository(), func(context.Context, RunReportRequest) (ReportResult, error) {
		return ReportResult{}, errors.New("boom")
	}, Options{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/reports/run", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", recorder.Code)
	}
}

func TestAuthRequiredWhenTokenConfigured(t *testing.T) {
	router := NewRouter(newMemoryRepository(), func(context.Context, RunReportRequest) (ReportResult, error) {
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
	router := NewRouter(newMemoryRepository(), func(context.Context, RunReportRequest) (ReportResult, error) {
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
	router := NewRouter(newMemoryRepository(), func(context.Context, RunReportRequest) (ReportResult, error) {
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
			Markdown:    "# 今日信息日报\n\n## 今日概览\n- 收录条目：2\n\n## ⭐⭐⭐ 测试一\n- 标题：测试一\n",
			Items: []model.NewsItem{
				{Title: "测试一", Score: 4},
				{Title: "库存条目", Score: 2},
			},
		},
	}
	router := NewRouter(repo, func(context.Context, RunReportRequest) (ReportResult, error) {
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
	if !strings.Contains(body, `"high_priority_count":1`) || !strings.Contains(body, `"top_titles":["测试一","库存条目"]`) {
		t.Fatalf("expected reports list to include quick summary fields, got %s", body)
	}
}

func TestReportsListKeepsItemCountAndDisplayCountIndependent(t *testing.T) {
	repo := newMemoryRepository()
	repo.records = []repository.ReportRecord{
		{
			GeneratedAt: time.Date(2026, 5, 3, 16, 0, 0, 0, time.UTC),
			Markdown:    "# report\n\n## 今日概览\nbody\n\n## ⭐ item one\nbody\n\n## ⭐ item two\nbody\n",
			Items: []model.NewsItem{
				{Title: "one"},
				{Title: "two"},
				{Title: "three"},
				{Title: "four"},
			},
		},
	}
	router := NewRouter(repo, func(context.Context, RunReportRequest) (ReportResult, error) {
		return ReportResult{}, nil
	}, Options{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/reports", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `"item_count":4`) || !strings.Contains(body, `"display_count":2`) {
		t.Fatalf("expected list response to preserve independent counts, got %s", body)
	}
}

func TestHealthSkipsAuth(t *testing.T) {
	router := NewRouter(newMemoryRepository(), func(context.Context, RunReportRequest) (ReportResult, error) {
		return ReportResult{}, nil
	}, Options{AuthToken: "secret"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected health to skip auth, got %d", recorder.Code)
	}
}

func TestIndexSkipsAuth(t *testing.T) {
	router := NewRouter(newMemoryRepository(), func(context.Context, RunReportRequest) (ReportResult, error) {
		return ReportResult{}, nil
	}, Options{AuthToken: "secret"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected index to skip auth, got %d", recorder.Code)
	}
}

func TestPreferenceEndpointsSaveAndRead(t *testing.T) {
	preferenceRepo := newMemoryPreferenceRepository()
	router := NewRouter(newMemoryRepository(), func(context.Context, RunReportRequest) (ReportResult, error) {
		return ReportResult{}, nil
	}, Options{UserPreferenceRepo: preferenceRepo})

	putBody := bytes.NewBufferString(`{"tags":["AI"],"sources":["openai-news"],"keywords":["agent"],"weights":{"tag":1.5,"source":1.2,"keyword":0.8}}`)
	putRecorder := httptest.NewRecorder()
	putRequest := httptest.NewRequest(http.MethodPut, "/preferences/alice", putBody)
	putRequest.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(putRecorder, putRequest)

	if putRecorder.Code != http.StatusOK {
		t.Fatalf("expected put status 200, got %d", putRecorder.Code)
	}

	getRecorder := httptest.NewRecorder()
	getRequest := httptest.NewRequest(http.MethodGet, "/preferences/alice", nil)
	router.ServeHTTP(getRecorder, getRequest)

	if getRecorder.Code != http.StatusOK {
		t.Fatalf("expected get status 200, got %d", getRecorder.Code)
	}
	body := getRecorder.Body.String()
	if !strings.Contains(body, `"user_id":"alice"`) || !strings.Contains(body, `"tag":1.5`) {
		t.Fatalf("expected stored preference in response, got %s", body)
	}
}

func TestPreferenceGetReturnsNotFound(t *testing.T) {
	router := NewRouter(newMemoryRepository(), func(context.Context, RunReportRequest) (ReportResult, error) {
		return ReportResult{}, nil
	}, Options{UserPreferenceRepo: newMemoryPreferenceRepository()})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/preferences/missing", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
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

func (r *memoryRepository) Get(ctx context.Context, name string) (repository.ReportRecord, error) {
	for _, record := range r.records {
		if record.GeneratedAt.Format("20060102-150405") == name {
			return record, nil
		}
	}

	return repository.ReportRecord{}, repository.ErrReportNotFound
}

func (r *memoryRepository) List(ctx context.Context) ([]repository.ReportMetadata, error) {
	result := make([]repository.ReportMetadata, 0, len(r.records))
	for _, record := range r.records {
		overview := buildReportOverview(record.Markdown, record.Items, 2)
		result = append(result, repository.ReportMetadata{
			Name:              record.GeneratedAt.Format("20060102-150405"),
			ItemCount:         len(record.Items),
			DisplayCount:      overview.DisplayCount,
			HighPriorityCount: overview.HighPriorityCount,
			TopTitles:         overview.TopPriorityItems,
			CreatedAt:         record.GeneratedAt,
		})
	}

	return result, nil
}

type memoryPreferenceRepository struct {
	records map[string]repository.UserPreferenceRecord
}

type memorySearchRepository struct {
	records []repository.SearchRecord
}

func newMemorySearchRepository() *memorySearchRepository {
	return &memorySearchRepository{}
}

func (r *memorySearchRepository) Save(ctx context.Context, record repository.SearchRecord) error {
	r.records = append(r.records, record)
	return nil
}

func (r *memorySearchRepository) Latest(ctx context.Context) (repository.SearchRecord, error) {
	if len(r.records) == 0 {
		return repository.SearchRecord{}, repository.ErrReportNotFound
	}
	return r.records[len(r.records)-1], nil
}

func (r *memorySearchRepository) Get(ctx context.Context, name string) (repository.SearchRecord, error) {
	for _, record := range r.records {
		if record.GeneratedAt.Format("20060102-150405") == name {
			return record, nil
		}
	}
	return repository.SearchRecord{}, repository.ErrReportNotFound
}

func (r *memorySearchRepository) List(ctx context.Context) ([]repository.SearchMetadata, error) {
	result := make([]repository.SearchMetadata, 0, len(r.records))
	for _, record := range r.records {
		result = append(result, repository.BuildSearchMetadata(
			record.GeneratedAt.Format("20060102-150405"),
			record.Query,
			record.Markdown,
			record.Items,
			record.GeneratedAt,
			2,
		))
	}
	return result, nil
}

func newMemoryPreferenceRepository() *memoryPreferenceRepository {
	return &memoryPreferenceRepository{records: map[string]repository.UserPreferenceRecord{}}
}

func (r *memoryPreferenceRepository) Save(ctx context.Context, record repository.UserPreferenceRecord) error {
	r.records[record.UserID] = record
	return nil
}

func (r *memoryPreferenceRepository) Get(ctx context.Context, userID string) (repository.UserPreferenceRecord, error) {
	record, ok := r.records[userID]
	if !ok {
		return repository.UserPreferenceRecord{}, repository.ErrUserPreferenceNotFound
	}

	return record, nil
}
