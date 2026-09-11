package agent_test

import (
	"context"
	"testing"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/agent"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/llm"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/policy"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/store"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/tool"
)

func TestRuntimeRunsToolPlanWithPlaceholders(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	reg := tool.NewRegistry()
	reg.Register(&tool.FuncTool{
		ToolName:        "test.echo",
		ToolDescription: "echo 输入",
		ToolRisk:        domain.RiskRead,
		Fn: func(_ context.Context, in map[string]any) (*domain.ToolResult, error) {
			return &domain.ToolResult{Output: map[string]any{"echo": in["x"]}, Summary: "ok"}, nil
		},
	})
	rt := agent.NewRuntime(st, reg, llm.NewMock("mock-1"), policy.New(true), nil)

	spec := domain.AgentSpec{
		ID: "t", Name: "测试 Agent", Planner: "rule",
		Steps: []domain.PlanStep{
			{ID: "e", Action: "tool", Tool: "test.echo", Args: map[string]any{"x": "你好"}},
			{ID: "out", Action: "output", Args: map[string]any{"v": "{{e.echo}}"}},
		},
	}
	task := &domain.Task{ID: "task-1", Goal: "验证占位符与工具执行"}
	res, err := rt.Run(ctx, task, spec)
	if err != nil {
		t.Fatalf("Run 失败: %v", err)
	}
	if res.Status != string(domain.RunSucceeded) {
		t.Fatalf("期望 succeeded，得到 %s", res.Status)
	}
	run, _ := st.GetRun(ctx, res.RunID)
	if run == nil || len(run.Steps) == 0 {
		t.Fatalf("未记录 run/steps")
	}
}

func TestRuntimeApprovalGate(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	reg := tool.NewRegistry()
	reg.Register(&tool.FuncTool{
		ToolName: "external.publish", ToolDescription: "对外发布", ToolRisk: domain.RiskExternal, Approval: true,
		Fn: func(_ context.Context, _ map[string]any) (*domain.ToolResult, error) {
			return &domain.ToolResult{Output: map[string]any{"published": true}}, nil
		},
	})
	// auto_approve=false => 外部动作应触发审批并挂起
	rt := agent.NewRuntime(st, reg, llm.NewMock("mock-1"), policy.New(false), nil)
	spec := domain.AgentSpec{ID: "t2", Name: "外部动作", Steps: []domain.PlanStep{
		{ID: "p", Action: "tool", Tool: "external.publish"},
	}}
	res, err := rt.Run(ctx, &domain.Task{ID: "task-2", Goal: "发布"}, spec)
	if err != nil {
		t.Fatalf("Run 返回错误: %v", err)
	}
	if res.Status != string(domain.RunAwaitingApproval) {
		t.Fatalf("期望 awaiting_approval，得到 %s", res.Status)
	}
	approvals, _ := st.ListApprovals(ctx, "pending", 10)
	if len(approvals) == 0 {
		t.Fatalf("未创建审批单")
	}
}
