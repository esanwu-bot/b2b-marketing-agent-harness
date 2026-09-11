// Package workflow 是 Workflow 引擎：按定义顺序编排 Agent / 审批 / 渠道发布。
package workflow

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/agent"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/policy"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/store"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/tool"
)

// Engine 执行 Workflow。
type Engine struct {
	Store   store.Store
	Agents  *agent.Registry
	Runtime *agent.Runtime
	Tools   *tool.Registry
	Policy  *policy.Engine
	Log     *slog.Logger
}

// NewEngine 创建 Workflow 引擎。
func NewEngine(st store.Store, agents *agent.Registry, rt *agent.Runtime, tools *tool.Registry, pol *policy.Engine, log *slog.Logger) *Engine {
	if log == nil {
		log = slog.Default()
	}
	return &Engine{Store: st, Agents: agents, Runtime: rt, Tools: tools, Policy: pol, Log: log}
}

// Run 执行指定 Workflow。
func (e *Engine) Run(ctx context.Context, key, triggerType string) (*domain.WorkflowRun, error) {
	def, err := e.Store.GetWorkflow(ctx, key)
	if err != nil {
		return nil, err
	}
	if def == nil {
		return nil, fmt.Errorf("workflow 不存在: %s", key)
	}

	run := &domain.WorkflowRun{
		WorkflowKey: key,
		Status:      domain.WfRunning,
		TriggerType: triggerType,
		Context:     map[string]any{},
		StartedAt:   time.Now(),
	}
	if err := e.Store.CreateWorkflowRun(ctx, run); err != nil {
		return nil, err
	}
	_ = e.Store.AppendEvent(ctx, domain.Event{Type: "workflow.run.started", AggregateType: "workflow_run", AggregateID: run.ID, Payload: map[string]any{"workflow": key}})

	for i, s := range def.Steps {
		stepRun := domain.WorkflowStepRun{Seq: i, Kind: s.Kind, Ref: refOf(s), Status: "running", At: time.Now()}

		switch s.Kind {
		case domain.WfAgent:
			step, err := e.runAgentStep(ctx, run, s)
			if err != nil {
				stepRun.Status = "failed"
				stepRun.Output = map[string]any{"error": err.Error()}
				run.Steps = append(run.Steps, stepRun)
				return e.finish(ctx, run, domain.WfFailed, err.Error())
			}
			stepRun.Status = step.status
			stepRun.RunID = step.runID
			stepRun.TaskID = step.taskID
			stepRun.Output = step.output
			if step.status == "awaiting_approval" {
				run.Steps = append(run.Steps, stepRun)
				return e.finish(ctx, run, domain.WfAwaitingApproval, "")
			}

		case domain.WfApproval:
			apr := &domain.Approval{
				WorkflowRunID: run.ID, Kind: "workflow_approval", Risk: domain.RiskExternal,
				Summary: "Workflow 请求人工审批：" + def.Name, Status: "pending", RequestedBy: "workflow",
			}
			if err := e.Store.CreateApproval(ctx, apr); err != nil {
				return nil, err
			}
			if e.Policy.AutoApprove() {
				now := time.Now()
				apr.Status, apr.DecidedBy, apr.DecidedAt = "approved", "policy:auto_approve", &now
				_ = e.Store.UpdateApproval(ctx, apr)
				stepRun.Status = "success"
				stepRun.Output = map[string]any{"approval_id": apr.ID, "status": "auto_approved"}
			} else {
				stepRun.Status = "awaiting_approval"
				stepRun.Output = map[string]any{"approval_id": apr.ID}
				run.Steps = append(run.Steps, stepRun)
				return e.finish(ctx, run, domain.WfAwaitingApproval, "")
			}

		case domain.WfChannel:
			out, err := e.runChannelStep(ctx, run, s)
			if err != nil {
				stepRun.Status = "failed"
				stepRun.Output = map[string]any{"error": err.Error()}
				run.Steps = append(run.Steps, stepRun)
				return e.finish(ctx, run, domain.WfFailed, err.Error())
			}
			stepRun.Status = "success"
			stepRun.Output = out

		default: // delay / condition 等：V1 直接跳过
			stepRun.Status = "success"
			stepRun.Output = map[string]any{"skipped": true}
		}

		run.Steps = append(run.Steps, stepRun)
		_ = e.Store.UpdateWorkflowRun(ctx, run)
	}

	return e.finish(ctx, run, domain.WfSucceeded, "")
}

type agentStepResult struct {
	status string
	runID  string
	taskID string
	output map[string]any
}

func (e *Engine) runAgentStep(ctx context.Context, run *domain.WorkflowRun, s domain.WorkflowStepDef) (*agentStepResult, error) {
	ag, ok := e.Agents.Get(s.Agent)
	if !ok {
		return nil, fmt.Errorf("未注册的 Agent: %s", s.Agent)
	}
	goal := s.Task
	if goal == "" {
		goal = ag.Name()
	}
	input := cloneMap(run.Context)
	if len(s.Input) > 0 {
		for k, v := range s.Input {
			input[k] = resolveValue(run.Context, v)
		}
	}
	task := &domain.Task{
		WorkflowRunID: run.ID,
		AgentID:       s.Agent,
		Type:          s.Task,
		Goal:          goal,
		Input:         input,
		Status:        domain.TaskRunning,
	}
	if err := e.Store.CreateTask(ctx, task); err != nil {
		return nil, err
	}

	result, err := ag.Run(ctx, task)
	if err != nil {
		task.Status = domain.TaskFailed
		_ = e.Store.UpdateTask(ctx, task)
		return nil, err
	}

	res := &agentStepResult{runID: result.RunID, taskID: task.ID, output: result.Output, status: "success"}
	switch result.Status {
	case string(domain.RunAwaitingApproval):
		res.status = "awaiting_approval"
		task.Status = domain.TaskAwaitingApproval
	case string(domain.RunFailed):
		res.status = "failed"
		task.Status = domain.TaskFailed
	default:
		task.Status = domain.TaskSucceeded
	}
	_ = e.Store.UpdateTask(ctx, task)

	if result.Output != nil {
		// 以 Agent ID 为键写回编排上下文，便于后续步骤/渠道引用。
		run.Context[s.Agent] = result.Output
	}
	return res, nil
}

func (e *Engine) runChannelStep(ctx context.Context, run *domain.WorkflowRun, s domain.WorkflowStepDef) (map[string]any, error) {
	toolName := channelToolName(s.Channel, s.Action)
	tl, ok := e.Tools.Get(toolName)
	if !ok {
		return nil, fmt.Errorf("未注册的渠道工具: %s", toolName)
	}
	args := map[string]any{}
	for k, v := range s.Input {
		args[k] = resolveValue(run.Context, v)
	}
	// 渠道动作通常为 external，需策略放行（demo/低风险环境可自动审批）。
	decision := e.Policy.Evaluate(domain.Action{Tool: toolName, Risk: tl.Risk(), Arguments: args})
	if decision.Outcome == domain.OutcomeDeny {
		return nil, fmt.Errorf("策略拒绝渠道动作 %s：%s", toolName, decision.Reason)
	}
	if decision.Outcome == domain.OutcomeRequireApproval && !e.Policy.AutoApprove() {
		apr := &domain.Approval{WorkflowRunID: run.ID, Kind: toolName, Summary: "Workflow 请求发布：" + toolName, Payload: args, Risk: tl.Risk(), Status: "pending"}
		_ = e.Store.CreateApproval(ctx, apr)
		return map[string]any{"approval_id": apr.ID, "status": "awaiting_approval"}, fmt.Errorf("等待审批")
	}
	result, err := tl.Execute(ctx, args)
	if err != nil {
		return nil, err
	}
	return result.Output, nil
}

func (e *Engine) finish(ctx context.Context, run *domain.WorkflowRun, status domain.WorkflowRunStatus, errMsg string) (*domain.WorkflowRun, error) {
	now := time.Now()
	run.Status = status
	run.Error = errMsg
	run.FinishedAt = &now
	if status != domain.WfRunning && status != domain.WfAwaitingApproval {
		run.FinishedAt = &now
	}
	_ = e.Store.UpdateWorkflowRun(ctx, run)
	_ = e.Store.AppendEvent(ctx, domain.Event{Type: "workflow.run." + string(status), AggregateType: "workflow_run", AggregateID: run.ID, Payload: map[string]any{"error": errMsg}})
	return run, nil
}

func refOf(s domain.WorkflowStepDef) string {
	switch s.Kind {
	case domain.WfAgent:
		return "agent:" + s.Agent
	case domain.WfChannel:
		return "channel:" + s.Channel
	case domain.WfApproval:
		return "approval"
	default:
		return string(s.Kind)
	}
}

func channelToolName(channel, action string) string {
	switch action {
	case "schedule":
		return channel + ".schedule_post"
	default:
		return channel + ".create_post"
	}
}

func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// resolveValue 解析 {{path}} 占位符（在编排上下文中查找）。
func resolveValue(root map[string]any, v any) any {
	switch x := v.(type) {
	case string:
		s := strings.TrimSpace(x)
		if strings.HasPrefix(s, "{{") && strings.HasSuffix(s, "}}") {
			path := strings.TrimSpace(s[2 : len(s)-2])
			if val, ok := lookupPath(root, path); ok {
				return val
			}
		}
		return x
	case map[string]any:
		m := make(map[string]any, len(x))
		for k, vv := range x {
			m[k] = resolveValue(root, vv)
		}
		return m
	default:
		return v
	}
}

func lookupPath(root map[string]any, path string) (any, bool) {
	parts := strings.Split(path, ".")
	var cur any = root
	for _, p := range parts {
		switch c := cur.(type) {
		case map[string]any:
			v, ok := c[p]
			if !ok {
				return nil, false
			}
			cur = v
		case []any:
			idx, err := strconv.Atoi(p)
			if err != nil || idx < 0 || idx >= len(c) {
				return nil, false
			}
			cur = c[idx]
		default:
			return nil, false
		}
	}
	return cur, true
}
