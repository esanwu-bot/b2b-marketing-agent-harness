package agent

import (
	"context"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
)

// SpecAgent 是「由规格驱动」的通用 Agent：行为完全由 AgentSpec + Runtime 决定。
//
// 6 个内置 Agent（research/trend/strategy/content/review/analytics）都是它的实例，
// 因此新增 Agent 只需新增一份 spec，无需改 Runtime。
type SpecAgent struct {
	spec domain.AgentSpec
	rt   *Runtime
}

// NewSpecAgent 创建规格驱动 Agent。
func NewSpecAgent(spec domain.AgentSpec, rt *Runtime) *SpecAgent {
	return &SpecAgent{spec: spec, rt: rt}
}

// ID 实现 domain.Agent。
func (a *SpecAgent) ID() string { return a.spec.ID }

// Name 实现 domain.Agent。
func (a *SpecAgent) Name() string { return a.spec.Name }

// Spec 实现 domain.Agent。
func (a *SpecAgent) Spec() domain.AgentSpec { return a.spec }

// Run 实现 domain.Agent。
func (a *SpecAgent) Run(ctx context.Context, task *domain.Task) (*domain.Result, error) {
	return a.rt.Run(ctx, task, a.spec)
}
