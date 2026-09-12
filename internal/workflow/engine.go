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

// workflowApprovedKey 是 run.Context 中标记「本次运行已通过人工审批」的键。
// 审批通过恢复后，后续外部动作（如渠道发布）不再重复触发策略审批。
const workflowApprovedKey = "__workflow_approved"

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

	return e.runFrom(ctx, run, def, 0)
}

// Resume 从 awaiting_approval 状态恢复一个 WorkflowRun，继续执行剩余步骤。
// 调用方需先把对应审批单置为 approved。
func (e *Engine) Resume(ctx context.Context, runID string) (*domain.WorkflowRun, error) {
	run, err := e.Store.GetWorkflowRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	if run == nil {
		return nil, fmt.Errorf("workflow run 不存在: %s", runID)
	}
	if run.Status != domain.WfAwaitingApproval {
		return run, nil
	}

	def, err := e.Store.GetWorkflow(ctx, run.WorkflowKey)
	if err != nil {
		return nil, err
	}

	// 将处于 awaiting_approval 的最后一步标记为 success，并从下一步继续。
	last := len(run.Steps) - 1
	if last >= 0 && run.Steps[last].Status == "awaiting_approval" {
		run.Steps[last].Status = "success"
	}
	// 审批已通过：后续外部动作不再重复触发策略审批。
	if run.Context == nil {
		run.Context = map[string]any{}
	}
	run.Context[workflowApprovedKey] = true
	run.Status = domain.WfRunning
	run.FinishedAt = nil
	_ = e.Store.UpdateWorkflowRun(ctx, run)
	_ = e.Store.AppendEvent(ctx, domain.Event{Type: "workflow.run.resumed", AggregateType: "workflow_run", AggregateID: run.ID})

	return e.runFrom(ctx, run, def, len(run.Steps))
}

// Cancel 终止一个 awaiting_approval / running 状态的 WorkflowRun。
func (e *Engine) Cancel(ctx context.Context, runID, reason string) (*domain.WorkflowRun, error) {
	run, err := e.Store.GetWorkflowRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	if run == nil {
		return nil, fmt.Errorf("workflow run 不存在: %s", runID)
	}
	if run.Status != domain.WfAwaitingApproval && run.Status != domain.WfRunning {
		return run, nil
	}
	last := len(run.Steps) - 1
	if last >= 0 && (run.Steps[last].Status == "awaiting_approval" || run.Steps[last].Status == "running") {
		run.Steps[last].Status = "cancelled"
		if run.Steps[last].Output == nil {
			run.Steps[last].Output = map[string]any{}
		}
		run.Steps[last].Output["reason"] = reason
	}
	return e.finish(ctx, run, domain.WfCancelled, reason)
}

// runFrom 从 startIdx 开始顺序执行 def 的步骤，直到全部完成、失败或再次进入审批等待。
func (e *Engine) runFrom(ctx context.Context, run *domain.WorkflowRun, def *domain.WorkflowDef, startIdx int) (*domain.WorkflowRun, error) {
	for i := startIdx; i < len(def.Steps); i++ {
		s := def.Steps[i]
		stepRun := domain.WorkflowStepRun{Seq: i, Kind: s.Kind, Ref: refOf(s), Status: "running", At: time.Now()}

		res, err := e.executeStep(ctx, run, s)
		if err != nil {
			stepRun.Status = "failed"
			stepRun.Output = map[string]any{"error": err.Error()}
			run.Steps = append(run.Steps, stepRun)
			return e.finish(ctx, run, domain.WfFailed, err.Error())
		}
		stepRun.Status = res.status
		stepRun.RunID = res.runID
		stepRun.TaskID = res.taskID
		stepRun.Output = res.output
		run.Steps = append(run.Steps, stepRun)
		if res.status == "awaiting_approval" {
			return e.finish(ctx, run, domain.WfAwaitingApproval, "")
		}
		_ = e.Store.UpdateWorkflowRun(ctx, run)
	}

	return e.finish(ctx, run, domain.WfSucceeded, "")
}

// stepResult 是单步执行结果。
type stepResult struct {
	status string // success | failed | awaiting_approval
	runID  string
	taskID string
	output map[string]any
}

// executeStep 调度执行单步，返回 stepResult（err 仅用于真正的执行失败）。
func (e *Engine) executeStep(ctx context.Context, run *domain.WorkflowRun, s domain.WorkflowStepDef) (*stepResult, error) {
	switch s.Kind {
	case domain.WfAgent:
		return e.executeAgentStep(ctx, run, s)
	case domain.WfApproval:
		return e.executeApprovalStep(ctx, run, s)
	case domain.WfChannel:
		return e.executeChannelStep(ctx, run, s)
	default: // delay / condition 等：V1 直接跳过
		return &stepResult{status: "success", output: map[string]any{"skipped": true}}, nil
	}
}

func (e *Engine) executeAgentStep(ctx context.Context, run *domain.WorkflowRun, s domain.WorkflowStepDef) (*stepResult, error) {
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

	res := &stepResult{runID: result.RunID, taskID: task.ID, output: result.Output, status: "success"}
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

func (e *Engine) executeApprovalStep(ctx context.Context, run *domain.WorkflowRun, _ domain.WorkflowStepDef) (*stepResult, error) {
	def, _ := e.Store.GetWorkflow(ctx, run.WorkflowKey)
	name := run.WorkflowKey
	if def != nil {
		name = def.Name
	}
	apr := &domain.Approval{
		WorkflowRunID: run.ID, Kind: "workflow_approval", Risk: domain.RiskExternal,
		Summary: "Workflow 请求人工审批：" + name, Status: "pending", RequestedBy: "workflow",
	}
	if err := e.Store.CreateApproval(ctx, apr); err != nil {
		return nil, err
	}
	if e.Policy.AutoApprove() {
		now := time.Now()
		apr.Status, apr.DecidedBy, apr.DecidedAt = "approved", "policy:auto_approve", &now
		_ = e.Store.UpdateApproval(ctx, apr)
		return &stepResult{status: "success", output: map[string]any{"approval_id": apr.ID, "status": "auto_approved"}}, nil
	}
	return &stepResult{status: "awaiting_approval", output: map[string]any{"approval_id": apr.ID}}, nil
}

func (e *Engine) executeChannelStep(ctx context.Context, run *domain.WorkflowRun, s domain.WorkflowStepDef) (*stepResult, error) {
	toolName := channelToolName(s.Channel, s.Action)
	tl, ok := e.Tools.Get(toolName)
	if !ok {
		return nil, fmt.Errorf("未注册的渠道工具: %s", toolName)
	}
	args := map[string]any{}
	for k, v := range s.Input {
		args[k] = resolveValue(run.Context, v)
	}
	// 若本次运行已通过人工审批（run.Context[__workflow_approved]=true），外部动作直接放行。
	alreadyApproved, _ := run.Context[workflowApprovedKey].(bool)

	decision := e.Policy.Evaluate(domain.Action{Tool: toolName, Risk: tl.Risk(), Arguments: args})
	if decision.Outcome == domain.OutcomeDeny {
		return nil, fmt.Errorf("策略拒绝渠道动作 %s：%s", toolName, decision.Reason)
	}
	if decision.Outcome == domain.OutcomeRequireApproval && !e.Policy.AutoApprove() && !alreadyApproved {
		apr := &domain.Approval{WorkflowRunID: run.ID, Kind: toolName, Summary: "Workflow 请求发布：" + toolName, Payload: args, Risk: tl.Risk(), Status: "pending"}
		if err := e.Store.CreateApproval(ctx, apr); err != nil {
			return nil, err
		}
		return &stepResult{status: "awaiting_approval", output: map[string]any{"approval_id": apr.ID, "status": "awaiting_approval"}}, nil
	}
	result, err := tl.Execute(ctx, args)
	if err != nil {
		return nil, err
	}
	return &stepResult{status: "success", output: result.Output}, nil
}

func (e *Engine) finish(ctx context.Context, run *domain.WorkflowRun, status domain.WorkflowRunStatus, errMsg string) (*domain.WorkflowRun, error) {
	now := time.Now()
	run.Status = status
	run.Error = errMsg
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