// Package domain 定义 Agent Harness 的领域契约。
//
// 本层零外部依赖（仅标准库），所有实现（Runtime/Tool/Store/LLM）都面向这些接口，
// 以保证可替换、可 mock、可测试。
package domain

import (
	"context"
	"time"
)

// RunPolicy 限定单次 Agent Run 的资源与风险边界，禁止无限循环。
type RunPolicy struct {
	MaxSteps        int           `json:"max_steps"`
	MaxTokens       int           `json:"max_tokens"`
	Timeout         time.Duration `json:"timeout"`
	MaxToolCalls    int           `json:"max_tool_calls"`
	MaxCost         float64       `json:"max_cost"`
	RequireApproval bool          `json:"require_approval"`
}

// DefaultRunPolicy 返回保守的默认策略。
func DefaultRunPolicy() RunPolicy {
	return RunPolicy{
		MaxSteps:     12,
		MaxTokens:    16000,
		Timeout:      5 * time.Minute,
		MaxToolCalls: 20,
		MaxCost:      1.0,
	}
}

// PlanStep 是 Rule Planner 的声明式步骤（也是 LLM Plan 的输出单元）。
type PlanStep struct {
	ID          string         `json:"id"`
	Action      string         `json:"action"` // tool | output
	Tool        string         `json:"tool,omitempty"`
	Args        map[string]any `json:"args,omitempty"`
	Description string         `json:"description,omitempty"`
}

// Plan 是 Planner 的产出，Executor 才负责执行。
type Plan struct {
	Goal  string     `json:"goal"`
	Steps []PlanStep `json:"steps"`
}

// AgentSpec 是一个 Agent 的可持久化规格（对应 agents 表）。
type AgentSpec struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Role         string         `json:"role"`
	Description  string         `json:"description"`
	SystemPrompt string         `json:"system_prompt"`
	Tools        []string       `json:"tools"`
	Steps        []PlanStep     `json:"steps"`
	Planner      string         `json:"planner"` // rule | llm
	Policy       RunPolicy      `json:"policy"`
	Config       map[string]any `json:"config,omitempty"`
}

// Agent 是所有 Agent 的统一契约。
type Agent interface {
	ID() string
	Name() string
	Spec() AgentSpec
	Run(ctx context.Context, task *Task) (*Result, error)
}

// Result 是一次 Agent Run 的产出。
type Result struct {
	RunID  string         `json:"run_id"`
	Status string         `json:"status"`
	Output map[string]any `json:"output,omitempty"`
	Error  string         `json:"error,omitempty"`
}
