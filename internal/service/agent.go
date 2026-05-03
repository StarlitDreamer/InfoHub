package service

import (
	"context"
	"time"

	"InfoHub-agent/internal/delivery"
	"InfoHub-agent/internal/model"
	"InfoHub-agent/internal/repository"
)

// Agent 负责统一编排采集、处理、存储和输出流程。
type Agent struct {
	pipeline        PipelineRunner
	repo            repository.ReportRepository
	webhookSender   MarkdownSender
	sendEmptyReport bool
	reportMaxItems  int
	now             func() time.Time
}

// PipelineRunner 定义 Agent 依赖的信息处理管道。
type PipelineRunner interface {
	RunContext(ctx context.Context) ([]model.NewsItem, error)
}

// MarkdownSender 定义 Markdown 输出的外部投递能力。
type MarkdownSender interface {
	Send(markdown string) error
}

// Result 表示一次 Agent 执行结果。
type Result struct {
	ItemCount    int
	DisplayCount int
	GeneratedAt  time.Time
	Markdown     string
	Items        []model.NewsItem
	Request      AgentRequest
}

// Source 定义 Agent 本次执行使用的数据源。
type Source struct {
	Name     string
	Kind     string
	Location string
}

// ExecutionContext 描述一次 Agent 执行上下文。
type ExecutionContext struct {
	Trigger     string
	RequestedAt time.Time
	Sources     []Source
}

// AgentRequest 表示一次 Agent 任务输入。
type AgentRequest struct {
	Context ExecutionContext
}

// AgentOptions 保存 Agent 运行选项。
type AgentOptions struct {
	WebhookSender   MarkdownSender
	SendEmptyReport bool
	ReportMaxItems  int
	Now             func() time.Time
}

// NewAgent 创建 Agent 编排服务。
func NewAgent(pipeline PipelineRunner, repo repository.ReportRepository, options AgentOptions) *Agent {
	now := options.Now
	if now == nil {
		now = time.Now
	}

	return &Agent{
		pipeline:        pipeline,
		repo:            repo,
		webhookSender:   options.WebhookSender,
		sendEmptyReport: options.SendEmptyReport,
		reportMaxItems:  options.ReportMaxItems,
		now:             now,
	}
}

// Run 执行 Agent 主流程。
func (a *Agent) Run(ctx context.Context) (Result, error) {
	return a.RunWithRequest(ctx, AgentRequest{})
}

// RunWithRequest 执行带上下文输入的 Agent 主流程。
func (a *Agent) RunWithRequest(ctx context.Context, request AgentRequest) (Result, error) {
	items, err := a.pipeline.RunContext(ctx)
	if err != nil {
		return Result{}, err
	}

	sortedItems := SortByDecisionScore(items, a.now())
	displayItems := LimitItems(sortedItems, a.reportMaxItems)
	report := delivery.RenderMarkdown(displayItems)
	generatedAt := a.now()

	record := repository.ReportRecord{
		GeneratedAt: generatedAt,
		Markdown:    report,
		Items:       sortedItems,
	}
	if err := a.repo.Save(ctx, record); err != nil {
		return Result{}, err
	}

	if a.webhookSender != nil && (len(sortedItems) > 0 || a.sendEmptyReport) {
		if err := a.webhookSender.Send(report); err != nil {
			return Result{}, err
		}
	}

	return Result{
		ItemCount:    len(sortedItems),
		DisplayCount: len(displayItems),
		GeneratedAt:  generatedAt,
		Markdown:     report,
		Items:        sortedItems,
		Request:      normalizeAgentRequest(request, generatedAt),
	}, nil
}

func normalizeAgentRequest(request AgentRequest, requestedAt time.Time) AgentRequest {
	if request.Context.RequestedAt.IsZero() {
		request.Context.RequestedAt = requestedAt
	}
	if request.Context.Trigger == "" {
		request.Context.Trigger = "manual"
	}
	if request.Context.Sources == nil {
		request.Context.Sources = []Source{}
	}

	return request
}
