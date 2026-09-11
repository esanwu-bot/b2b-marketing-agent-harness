package agent

import (
	"sort"
	"sync"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
)

// Registry 是 Agent 注册表（并发安全）。
type Registry struct {
	mu     sync.RWMutex
	agents map[string]domain.Agent
}

// NewRegistry 创建注册表。
func NewRegistry() *Registry {
	return &Registry{agents: map[string]domain.Agent{}}
}

// Register 注册 Agent。
func (r *Registry) Register(a domain.Agent) {
	if a == nil {
		return
	}
	r.mu.Lock()
	r.agents[a.ID()] = a
	r.mu.Unlock()
}

// Get 取 Agent。
func (r *Registry) Get(id string) (domain.Agent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.agents[id]
	return a, ok
}

// List 返回全部 Agent（按 ID 排序）。
func (r *Registry) List() []domain.Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Agent, 0, len(r.agents))
	for _, a := range r.agents {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID() < out[j].ID() })
	return out
}

// Specs 返回全部 Agent 规格。
func (r *Registry) Specs() []domain.AgentSpec {
	out := make([]domain.AgentSpec, 0)
	for _, a := range r.List() {
		out = append(out, a.Spec())
	}
	return out
}
