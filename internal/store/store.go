// Package store 定义 Harness 的持久化契约，并提供内存实现（mock）。
//
// PostgreSQL 实现将落在 store/postgres；V1 先用内存实现保证离线可跑。
package store

import (
	"context"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
)

// Store 是 Harness 的持久化契约。
type Store interface {
	// Name 存储实现名（memory / postgres）。
	Name() string
	// Ping 健康检查。
	Ping(ctx context.Context) error

	// ---- Agent ----
	SaveAgent(ctx context.Context, spec domain.AgentSpec) error
	GetAgent(ctx context.Context, id string) (*domain.AgentSpec, error)
	ListAgents(ctx context.Context) ([]domain.AgentSpec, error)

	// ---- Tool 元数据 ----
	SaveTool(ctx context.Context, info domain.ToolInfo) error
	ListTools(ctx context.Context) ([]domain.ToolInfo, error)

	// ---- Task ----
	CreateTask(ctx context.Context, t *domain.Task) error
	GetTask(ctx context.Context, id string) (*domain.Task, error)
	UpdateTask(ctx context.Context, t *domain.Task) error
	ListTasks(ctx context.Context, limit int) ([]*domain.Task, error)

	// ---- Run ----
	CreateRun(ctx context.Context, r *domain.Run) error
	UpdateRun(ctx context.Context, r *domain.Run) error
	GetRun(ctx context.Context, id string) (*domain.Run, error)
	ListRuns(ctx context.Context, limit int) ([]*domain.Run, error)
	AppendStep(ctx context.Context, runID string, step domain.Step) error
	AppendMessage(ctx context.Context, runID string, msg domain.Message) error
	AppendToolCall(ctx context.Context, runID string, call domain.ToolCall) error
	AppendDecision(ctx context.Context, runID string, d domain.Decision) error
	AppendLLMCall(ctx context.Context, runID string, c domain.LLMCall) error

	// ---- Approval ----
	CreateApproval(ctx context.Context, a *domain.Approval) error
	GetApproval(ctx context.Context, id string) (*domain.Approval, error)
	UpdateApproval(ctx context.Context, a *domain.Approval) error
	ListApprovals(ctx context.Context, status string, limit int) ([]*domain.Approval, error)

	// ---- Topic / Content ----
	SaveTopic(ctx context.Context, t *domain.Topic) error
	GetTopic(ctx context.Context, id string) (*domain.Topic, error)
	ListTopics(ctx context.Context, limit int) ([]*domain.Topic, error)
	SaveContent(ctx context.Context, c *domain.Content) error
	GetContent(ctx context.Context, id string) (*domain.Content, error)
	ListContent(ctx context.Context, limit int) ([]*domain.Content, error)
	UpdateContent(ctx context.Context, c *domain.Content) error

	// ---- Knowledge ----
	SaveDocument(ctx context.Context, d *domain.Document) error
	ListDocuments(ctx context.Context, limit int) ([]*domain.Document, error)
	SearchDocuments(ctx context.Context, query string, limit int) ([]domain.SearchHit, error)
	CountDocuments(ctx context.Context) (int, error)
	SaveMemory(ctx context.Context, m *domain.Memory) error
	SearchMemories(ctx context.Context, scope domain.MemoryScope, scopeID string, limit int) ([]*domain.Memory, error)

	// ---- Workflow ----
	SaveWorkflow(ctx context.Context, def domain.WorkflowDef) error
	GetWorkflow(ctx context.Context, key string) (*domain.WorkflowDef, error)
	ListWorkflows(ctx context.Context) ([]domain.WorkflowDef, error)
	CreateWorkflowRun(ctx context.Context, r *domain.WorkflowRun) error
	UpdateWorkflowRun(ctx context.Context, r *domain.WorkflowRun) error
	GetWorkflowRun(ctx context.Context, id string) (*domain.WorkflowRun, error)
	ListWorkflowRuns(ctx context.Context, limit int) ([]*domain.WorkflowRun, error)

	// ---- Publication / Metric ----
	SavePublication(ctx context.Context, p *domain.Publication) error
	ListPublications(ctx context.Context, limit int) ([]*domain.Publication, error)
	SaveMetric(ctx context.Context, m *domain.Metric) error
	ListMetrics(ctx context.Context, publicationID string, limit int) ([]*domain.Metric, error)

	// ---- Event ----
	AppendEvent(ctx context.Context, e domain.Event) error
	ListEvents(ctx context.Context, limit int) ([]domain.Event, error)
}
