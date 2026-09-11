// Package tool 提供 Tool Registry 与内置 Native Tool。
//
// Tool 是 Agent 与外部世界交互的唯一出口；所有副作用（采集/发布/生成/通知）
// 都封装为 Tool，便于 mock、审批与观测。
package tool

import (
	"context"
	"sort"
	"sync"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
)

// FuncTool 用函数适配 domain.Tool，减少样板。
type FuncTool struct {
	ToolName        string
	ToolDescription string
	Schema          map[string]any
	ToolRisk        string
	Approval        bool
	Fn              func(ctx context.Context, input map[string]any) (*domain.ToolResult, error)
}

// Name 实现 domain.Tool。
func (t *FuncTool) Name() string { return t.ToolName }

// Description 实现 domain.Tool。
func (t *FuncTool) Description() string { return t.ToolDescription }

// InputSchema 实现 domain.Tool。
func (t *FuncTool) InputSchema() map[string]any {
	if t.Schema == nil {
		return map[string]any{"type": "object", "properties": map[string]any{}}
	}
	return t.Schema
}

// Risk 实现 domain.Tool。
func (t *FuncTool) Risk() string {
	if t.ToolRisk == "" {
		return domain.RiskRead
	}
	return t.ToolRisk
}

// RequiresApproval 实现 domain.Tool。
func (t *FuncTool) RequiresApproval() bool { return t.Approval }

// Execute 实现 domain.Tool。
func (t *FuncTool) Execute(ctx context.Context, input map[string]any) (*domain.ToolResult, error) {
	if t.Fn == nil {
		return &domain.ToolResult{Output: map[string]any{}}, nil
	}
	return t.Fn(ctx, input)
}

// Registry 是 Tool 注册表（并发安全）。
type Registry struct {
	mu    sync.RWMutex
	tools map[string]domain.Tool
}

// NewRegistry 创建空注册表。
func NewRegistry() *Registry {
	return &Registry{tools: map[string]domain.Tool{}}
}

// Register 注册 Tool。
func (r *Registry) Register(t domain.Tool) {
	if t == nil {
		return
	}
	r.mu.Lock()
	r.tools[t.Name()] = t
	r.mu.Unlock()
}

// RegisterAll 批量注册。
func (r *Registry) RegisterAll(ts []domain.Tool) {
	for _, t := range ts {
		r.Register(t)
	}
}

// Get 取 Tool。
func (r *Registry) Get(name string) (domain.Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

// Names 返回全部 Tool 名（有序）。
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.tools))
	for k := range r.tools {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Infos 返回全部 Tool 元数据。
func (r *Registry) Infos() []domain.ToolInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.ToolInfo, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, domain.Info(t))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
