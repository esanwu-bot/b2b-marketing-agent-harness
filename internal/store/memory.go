package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
)

// Memory 是内存存储实现（mock）。并发安全，进程退出即丢失。
//
// 用于无 PostgreSQL/Redis 时跑通全链路、单元测试与演示。
type Memory struct {
	mu sync.RWMutex

	agents     map[string]domain.AgentSpec
	agentOrder []string
	tools      map[string]domain.ToolInfo

	tasks     map[string]*domain.Task
	taskOrder []string

	runs      map[string]*domain.Run
	runOrder  []string
	steps     map[string][]domain.Step
	messages  map[string][]domain.Message
	toolcalls map[string][]domain.ToolCall
	decisions map[string][]domain.Decision
	llmcalls  map[string][]domain.LLMCall

	approvals     map[string]*domain.Approval
	approvalOrder []string

	topics       map[string]*domain.Topic
	topicOrder   []string
	content      map[string]*domain.Content
	contentOrder []string

	documents map[string]*domain.Document
	docOrder  []string
	memories  map[string]*domain.Memory

	workflows        map[string]domain.WorkflowDef
	workflowRuns     map[string]*domain.WorkflowRun
	workflowRunOrder []string

	publications map[string]*domain.Publication
	pubOrder     []string
	metrics      map[string][]*domain.Metric

	events []domain.Event
}

// NewMemory 创建内存存储。
func NewMemory() *Memory {
	return &Memory{
		agents:       map[string]domain.AgentSpec{},
		tools:        map[string]domain.ToolInfo{},
		tasks:        map[string]*domain.Task{},
		runs:         map[string]*domain.Run{},
		steps:        map[string][]domain.Step{},
		messages:     map[string][]domain.Message{},
		toolcalls:    map[string][]domain.ToolCall{},
		decisions:    map[string][]domain.Decision{},
		llmcalls:     map[string][]domain.LLMCall{},
		approvals:    map[string]*domain.Approval{},
		topics:       map[string]*domain.Topic{},
		content:      map[string]*domain.Content{},
		documents:    map[string]*domain.Document{},
		memories:     map[string]*domain.Memory{},
		workflows:    map[string]domain.WorkflowDef{},
		workflowRuns: map[string]*domain.WorkflowRun{},
		publications: map[string]*domain.Publication{},
		metrics:      map[string][]*domain.Metric{},
	}
}

// Name 实现 Store。
func (m *Memory) Name() string { return "memory" }

// Ping 实现 Store。
func (m *Memory) Ping(context.Context) error { return nil }

// NewID 生成带前缀的唯一 id。
func NewID(prefix string) string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%s-%s", prefix, time.Now().Format("20060102150405"), hex.EncodeToString(b))
}

// ---------------- Agent ----------------

func (m *Memory) SaveAgent(_ context.Context, spec domain.AgentSpec) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.agents[spec.ID]; !ok {
		m.agentOrder = append(m.agentOrder, spec.ID)
	}
	m.agents[spec.ID] = spec
	return nil
}

func (m *Memory) GetAgent(_ context.Context, id string) (*domain.AgentSpec, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.agents[id]
	if !ok {
		return nil, nil
	}
	return &s, nil
}

func (m *Memory) ListAgents(_ context.Context) ([]domain.AgentSpec, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]domain.AgentSpec, 0, len(m.agentOrder))
	for _, id := range m.agentOrder {
		out = append(out, m.agents[id])
	}
	return out, nil
}

// ---------------- Tool ----------------

func (m *Memory) SaveTool(_ context.Context, info domain.ToolInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tools[info.Name] = info
	return nil
}

func (m *Memory) ListTools(_ context.Context) ([]domain.ToolInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]domain.ToolInfo, 0, len(m.tools))
	for _, t := range m.tools {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// ---------------- Task ----------------

func (m *Memory) CreateTask(_ context.Context, t *domain.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t.ID == "" {
		t.ID = NewID("task")
	}
	now := time.Now()
	t.CreatedAt, t.UpdatedAt = now, now
	if t.Status == "" {
		t.Status = domain.TaskPending
	}
	m.tasks[t.ID] = t
	m.taskOrder = append(m.taskOrder, t.ID)
	return nil
}

func (m *Memory) GetTask(_ context.Context, id string) (*domain.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tasks[id]
	if !ok {
		return nil, nil
	}
	cp := *t
	return &cp, nil
}

func (m *Memory) UpdateTask(_ context.Context, t *domain.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tasks[t.ID]; !ok {
		return fmt.Errorf("task 不存在: %s", t.ID)
	}
	t.UpdatedAt = time.Now()
	m.tasks[t.ID] = t
	return nil
}

func (m *Memory) ListTasks(_ context.Context, limit int) ([]*domain.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*domain.Task, 0, len(m.taskOrder))
	for i := len(m.taskOrder) - 1; i >= 0; i-- {
		cp := *m.tasks[m.taskOrder[i]]
		out = append(out, &cp)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

// ---------------- Run ----------------

func (m *Memory) CreateRun(_ context.Context, r *domain.Run) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r.ID == "" {
		r.ID = NewID("run")
	}
	if r.StartedAt.IsZero() {
		r.StartedAt = time.Now()
	}
	if r.Status == "" {
		r.Status = domain.RunRunning
	}
	m.runs[r.ID] = r
	m.runOrder = append(m.runOrder, r.ID)
	return nil
}

func (m *Memory) UpdateRun(_ context.Context, r *domain.Run) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runs[r.ID] = r
	return nil
}

func (m *Memory) GetRun(_ context.Context, id string) (*domain.Run, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.runs[id]
	if !ok {
		return nil, nil
	}
	cp := *r
	cp.Steps = append([]domain.Step(nil), m.steps[id]...)
	return &cp, nil
}

func (m *Memory) ListRuns(_ context.Context, limit int) ([]*domain.Run, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*domain.Run, 0, len(m.runOrder))
	for i := len(m.runOrder) - 1; i >= 0; i-- {
		cp := *m.runs[m.runOrder[i]]
		cp.Steps = append([]domain.Step(nil), m.steps[cp.ID]...)
		out = append(out, &cp)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (m *Memory) AppendStep(_ context.Context, runID string, step domain.Step) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if step.At.IsZero() {
		step.At = time.Now()
	}
	m.steps[runID] = append(m.steps[runID], step)
	return nil
}

func (m *Memory) AppendMessage(_ context.Context, runID string, msg domain.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages[runID] = append(m.messages[runID], msg)
	return nil
}

func (m *Memory) AppendToolCall(_ context.Context, runID string, call domain.ToolCall) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.toolcalls[runID] = append(m.toolcalls[runID], call)
	return nil
}

func (m *Memory) AppendDecision(_ context.Context, runID string, d domain.Decision) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if d.At.IsZero() {
		d.At = time.Now()
	}
	m.decisions[runID] = append(m.decisions[runID], d)
	return nil
}

func (m *Memory) AppendLLMCall(_ context.Context, runID string, c domain.LLMCall) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c.At.IsZero() {
		c.At = time.Now()
	}
	m.llmcalls[runID] = append(m.llmcalls[runID], c)
	return nil
}

// ---------------- Approval ----------------

func (m *Memory) CreateApproval(_ context.Context, a *domain.Approval) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if a.ID == "" {
		a.ID = NewID("apr")
	}
	if a.Status == "" {
		a.Status = "pending"
	}
	if a.RequestedAt.IsZero() {
		a.RequestedAt = time.Now()
	}
	m.approvals[a.ID] = a
	m.approvalOrder = append(m.approvalOrder, a.ID)
	return nil
}

func (m *Memory) GetApproval(_ context.Context, id string) (*domain.Approval, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.approvals[id]
	if !ok {
		return nil, nil
	}
	cp := *a
	return &cp, nil
}

func (m *Memory) UpdateApproval(_ context.Context, a *domain.Approval) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.approvals[a.ID] = a
	return nil
}

func (m *Memory) ListApprovals(_ context.Context, status string, limit int) ([]*domain.Approval, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*domain.Approval, 0)
	for i := len(m.approvalOrder) - 1; i >= 0; i-- {
		a := m.approvals[m.approvalOrder[i]]
		if status != "" && a.Status != status {
			continue
		}
		cp := *a
		out = append(out, &cp)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

// ---------------- Topic / Content ----------------

func (m *Memory) SaveTopic(_ context.Context, t *domain.Topic) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t.ID == "" {
		t.ID = NewID("topic")
	}
	now := time.Now()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	t.UpdatedAt = now
	if t.Status == "" {
		t.Status = "candidate"
	}
	if _, ok := m.topics[t.ID]; !ok {
		m.topicOrder = append(m.topicOrder, t.ID)
	}
	m.topics[t.ID] = t
	return nil
}

func (m *Memory) GetTopic(_ context.Context, id string) (*domain.Topic, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.topics[id]
	if !ok {
		return nil, nil
	}
	cp := *t
	return &cp, nil
}

func (m *Memory) ListTopics(_ context.Context, limit int) ([]*domain.Topic, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*domain.Topic, 0, len(m.topics))
	for _, t := range m.topics {
		cp := *t
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (m *Memory) SaveContent(_ context.Context, c *domain.Content) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c.ID == "" {
		c.ID = NewID("content")
	}
	now := time.Now()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	c.UpdatedAt = now
	if c.Status == "" {
		c.Status = domain.ContentDraft
	}
	if _, ok := m.content[c.ID]; !ok {
		m.contentOrder = append(m.contentOrder, c.ID)
	}
	m.content[c.ID] = c
	return nil
}

func (m *Memory) GetContent(_ context.Context, id string) (*domain.Content, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.content[id]
	if !ok {
		return nil, nil
	}
	cp := *c
	return &cp, nil
}

func (m *Memory) ListContent(_ context.Context, limit int) ([]*domain.Content, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*domain.Content, 0, len(m.contentOrder))
	for i := len(m.contentOrder) - 1; i >= 0; i-- {
		cp := *m.content[m.contentOrder[i]]
		out = append(out, &cp)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (m *Memory) UpdateContent(_ context.Context, c *domain.Content) error {
	return m.SaveContent(context.Background(), c)
}

// ---------------- Knowledge ----------------

func (m *Memory) SaveDocument(_ context.Context, d *domain.Document) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if d.ID == "" {
		d.ID = NewID("doc")
	}
	if d.FetchedAt.IsZero() {
		d.FetchedAt = time.Now()
	}
	// 以 title+content 指纹去重
	for _, existing := range m.documents {
		if existing.Title == d.Title && existing.Content == d.Content {
			return nil
		}
	}
	if _, ok := m.documents[d.ID]; !ok {
		m.docOrder = append(m.docOrder, d.ID)
	}
	m.documents[d.ID] = d
	return nil
}

func (m *Memory) ListDocuments(_ context.Context, limit int) ([]*domain.Document, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*domain.Document, 0, len(m.docOrder))
	for i := len(m.docOrder) - 1; i >= 0; i-- {
		cp := *m.documents[m.docOrder[i]]
		out = append(out, &cp)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (m *Memory) SearchDocuments(_ context.Context, query string, limit int) ([]domain.SearchHit, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	q := strings.ToLower(strings.TrimSpace(query))
	terms := strings.Fields(q)
	out := make([]domain.SearchHit, 0)
	for _, d := range m.documents {
		hay := strings.ToLower(d.Title + " " + d.Content + " " + d.Summary)
		score := 0.0
		if q != "" && strings.Contains(hay, q) {
			score += 1.0
		}
		for _, t := range terms {
			if strings.Contains(hay, t) {
				score += 0.3
			}
		}
		if score == 0 && q != "" {
			continue
		}
		snippet := d.Summary
		if snippet == "" {
			snippet = truncate(d.Content, 160)
		}
		out = append(out, domain.SearchHit{DocumentID: d.ID, Title: d.Title, URL: d.URL, Snippet: snippet, Score: score})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (m *Memory) CountDocuments(context.Context) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.documents), nil
}

func (m *Memory) SaveMemory(_ context.Context, mem *domain.Memory) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if mem.ID == "" {
		mem.ID = NewID("mem")
	}
	now := time.Now()
	if mem.CreatedAt.IsZero() {
		mem.CreatedAt = now
	}
	mem.UpdatedAt = now
	m.memories[mem.ID] = mem
	return nil
}

func (m *Memory) SearchMemories(_ context.Context, scope domain.MemoryScope, scopeID string, limit int) ([]*domain.Memory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*domain.Memory, 0)
	for _, mem := range m.memories {
		if scope != "" && mem.Scope != scope {
			continue
		}
		if scopeID != "" && mem.ScopeID != scopeID {
			continue
		}
		cp := *mem
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Importance > out[j].Importance })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// ---------------- Workflow ----------------

func (m *Memory) SaveWorkflow(_ context.Context, def domain.WorkflowDef) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workflows[def.Key] = def
	return nil
}

func (m *Memory) GetWorkflow(_ context.Context, key string) (*domain.WorkflowDef, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	def, ok := m.workflows[key]
	if !ok {
		return nil, nil
	}
	return &def, nil
}

func (m *Memory) ListWorkflows(context.Context) ([]domain.WorkflowDef, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]domain.WorkflowDef, 0, len(m.workflows))
	for _, def := range m.workflows {
		out = append(out, def)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out, nil
}

func (m *Memory) CreateWorkflowRun(_ context.Context, r *domain.WorkflowRun) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r.ID == "" {
		r.ID = NewID("wfr")
	}
	if r.StartedAt.IsZero() {
		r.StartedAt = time.Now()
	}
	if r.Status == "" {
		r.Status = domain.WfRunning
	}
	m.workflowRuns[r.ID] = r
	m.workflowRunOrder = append(m.workflowRunOrder, r.ID)
	return nil
}

func (m *Memory) UpdateWorkflowRun(_ context.Context, r *domain.WorkflowRun) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workflowRuns[r.ID] = r
	return nil
}

func (m *Memory) GetWorkflowRun(_ context.Context, id string) (*domain.WorkflowRun, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.workflowRuns[id]
	if !ok {
		return nil, nil
	}
	cp := *r
	cp.Steps = append([]domain.WorkflowStepRun(nil), r.Steps...)
	return &cp, nil
}

func (m *Memory) ListWorkflowRuns(_ context.Context, limit int) ([]*domain.WorkflowRun, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*domain.WorkflowRun, 0, len(m.workflowRunOrder))
	for i := len(m.workflowRunOrder) - 1; i >= 0; i-- {
		r := m.workflowRuns[m.workflowRunOrder[i]]
		cp := *r
		cp.Steps = append([]domain.WorkflowStepRun(nil), r.Steps...)
		out = append(out, &cp)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

// ---------------- Publication / Metric ----------------

func (m *Memory) SavePublication(_ context.Context, p *domain.Publication) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p.ID == "" {
		p.ID = NewID("pub")
	}
	if _, ok := m.publications[p.ID]; !ok {
		m.pubOrder = append(m.pubOrder, p.ID)
	}
	m.publications[p.ID] = p
	return nil
}

func (m *Memory) ListPublications(_ context.Context, limit int) ([]*domain.Publication, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*domain.Publication, 0, len(m.pubOrder))
	for i := len(m.pubOrder) - 1; i >= 0; i-- {
		cp := *m.publications[m.pubOrder[i]]
		out = append(out, &cp)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (m *Memory) SaveMetric(_ context.Context, metric *domain.Metric) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics[metric.PublicationID] = append(m.metrics[metric.PublicationID], metric)
	return nil
}

func (m *Memory) ListMetrics(_ context.Context, publicationID string, limit int) ([]*domain.Metric, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*domain.Metric, 0)
	for _, metric := range m.metrics[publicationID] {
		cp := *metric
		out = append(out, &cp)
	}
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out, nil
}

// ---------------- Event ----------------

func (m *Memory) AppendEvent(_ context.Context, e domain.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if e.At.IsZero() {
		e.At = time.Now()
	}
	m.events = append(m.events, e)
	return nil
}

func (m *Memory) ListEvents(_ context.Context, limit int) ([]domain.Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := len(m.events)
	start := 0
	if limit > 0 && n > limit {
		start = n - limit
	}
	out := make([]domain.Event, 0, n-start)
	for i := n - 1; i >= start; i-- {
		out = append(out, m.events[i])
	}
	return out, nil
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
