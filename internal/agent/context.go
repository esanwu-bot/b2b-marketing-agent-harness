// Package agent 实现 Agent Harness 的核心：Runtime、Planner、ContextBuilder、Agent 注册表。
package agent

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/llm"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/store"
)

// ContextBuilder 组装送给模型的上下文：
// System Prompt + Agent Profile + Task + Knowledge + Memory。
type ContextBuilder struct {
	Store store.Store
}

// NewContextBuilder 创建上下文构建器。
func NewContextBuilder(st store.Store) *ContextBuilder {
	return &ContextBuilder{Store: st}
}

// Build 构建上下文文本。
func (b *ContextBuilder) Build(ctx context.Context, task *domain.Task, spec domain.AgentSpec) string {
	var sb strings.Builder
	sb.WriteString("### Agent\n")
	sb.WriteString(spec.Name + "（" + spec.Role + "）\n\n")

	sb.WriteString("### 任务目标\n")
	goal := task.Goal
	if goal == "" {
		goal = spec.Description
	}
	sb.WriteString(goal + "\n\n")

	if len(task.Input) > 0 {
		if raw, err := json.Marshal(task.Input); err == nil {
			sb.WriteString("### 任务输入\n" + string(raw) + "\n\n")
		}
	}

	if b.Store != nil {
		if hits, err := b.Store.SearchDocuments(ctx, goal, 5); err == nil && len(hits) > 0 {
			sb.WriteString("### 相关知识\n")
			for _, h := range hits {
				sb.WriteString("- " + h.Title + "：" + truncate(h.Snippet, 120) + "\n")
			}
			sb.WriteString("\n")
		}
		if mems, err := b.Store.SearchMemories(ctx, "", "", 5); err == nil && len(mems) > 0 {
			sb.WriteString("### 记忆\n")
			for _, m := range mems {
				sb.WriteString("- " + m.Content + "\n")
			}
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// Planner 只产生 Plan，不执行。
type Planner interface {
	Plan(ctx context.Context, rc *RunContext) (*domain.Plan, error)
}

// RulePlanner 依据 AgentSpec.Steps 产出确定性计划（离线/降级）。
type RulePlanner struct{}

// Plan 实现 Planner。
func (RulePlanner) Plan(_ context.Context, rc *RunContext) (*domain.Plan, error) {
	goal := rc.Task.Goal
	if goal == "" {
		goal = rc.Spec.Name
	}
	steps := rc.Spec.Steps
	if len(steps) == 0 {
		steps = []domain.PlanStep{{ID: "noop", Action: "output", Args: map[string]any{"message": "该 Agent 未配置步骤"}}}
	}
	return &domain.Plan{Goal: goal, Steps: steps}, nil
}

// LLMPlanner 用模型产出计划；解析失败则回退 RulePlanner。
type LLMPlanner struct {
	Provider llm.Provider
}

// Plan 实现 Planner。
func (p LLMPlanner) Plan(ctx context.Context, rc *RunContext) (*domain.Plan, error) {
	if p.Provider == nil {
		return RulePlanner{}.Plan(ctx, rc)
	}
	resp, err := p.Provider.Chat(ctx, llm.ChatRequest{
		JSON: true,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: "你是任务规划器。只输出 JSON：{\"goal\":\"...\",\"steps\":[{\"id\":\"\",\"action\":\"tool|output\",\"tool\":\"\",\"args\":{},\"description\":\"\"}]}"},
			{Role: llm.RoleUser, Content: rc.Context},
		},
	})
	if err != nil {
		return RulePlanner{}.Plan(ctx, rc)
	}
	content := resp.Content
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start < 0 || end <= start {
		return RulePlanner{}.Plan(ctx, rc)
	}
	var plan domain.Plan
	if json.Unmarshal([]byte(content[start:end+1]), &plan) != nil || len(plan.Steps) == 0 {
		return RulePlanner{}.Plan(ctx, rc)
	}
	return &plan, nil
}
