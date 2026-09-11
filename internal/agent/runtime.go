package agent

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/llm"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/policy"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/store"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/tool"
)

// Runtime 是 Agent 执行引擎：Plan → Context → Execute → Tool → Observe → Reflect。
type Runtime struct {
	Store   store.Store
	Tools   *tool.Registry
	LLM     llm.Provider
	Policy  *policy.Engine
	Context *ContextBuilder
	Log     *slog.Logger
}

// NewRuntime 创建 Runtime。
func NewRuntime(st store.Store, tools *tool.Registry, p llm.Provider, pol *policy.Engine, log *slog.Logger) *Runtime {
	if log == nil {
		log = slog.Default()
	}
	return &Runtime{
		Store:   st,
		Tools:   tools,
		LLM:     p,
		Policy:  pol,
		Context: NewContextBuilder(st),
		Log:     log,
	}
}

// RunContext 承载一次 Run 的可变状态（对应 AgentState）。
type RunContext struct {
	Task      *domain.Task
	Spec      domain.AgentSpec
	Policy    domain.RunPolicy
	Run       *domain.Run
	State     *domain.AgentState
	Context   string
	Variables map[string]any

	rt        *Runtime
	seq       int
	toolCalls int
	tokens    int
}

// Run 执行一次 Agent。
func (rt *Runtime) Run(ctx context.Context, task *domain.Task, spec domain.AgentSpec) (*domain.Result, error) {
	if task == nil {
		return nil, fmt.Errorf("task 不能为空")
	}
	pol := mergePolicy(spec.Policy)

	run := &domain.Run{
		TaskID:       task.ID,
		AgentID:      spec.ID,
		AgentVersion: 1,
		Status:       domain.RunRunning,
		StartedAt:    time.Now(),
	}
	if err := rt.Store.CreateRun(ctx, run); err != nil {
		return nil, err
	}
	state := &domain.AgentState{
		RunID:     run.ID,
		TaskID:    task.ID,
		AgentID:   spec.ID,
		Status:    domain.RunRunning,
		Variables: map[string]any{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	rc := &RunContext{Task: task, Spec: spec, Policy: pol, Run: run, State: state, Variables: state.Variables, rt: rt}

	upCtx, cancel := context.WithTimeout(ctx, pol.Timeout)
	defer cancel()

	rc.Context = rt.Context.Build(upCtx, task, spec)
	rt.recordMessage(upCtx, rc, llm.Message{Role: llm.RoleSystem, Content: spec.SystemPrompt})
	rt.recordMessage(upCtx, rc, llm.Message{Role: llm.RoleUser, Content: rc.Context})

	// 1) Plan
	planner := rt.plannerFor(spec)
	plan, err := planner.Plan(upCtx, rc)
	if err != nil {
		return rt.fail(upCtx, rc, "规划失败: "+err.Error())
	}
	rt.recordStep(upCtx, rc, domain.Step{
		Kind:   domain.StepPlan,
		Name:   "plan",
		Input:  map[string]any{"steps": len(plan.Steps)},
		Output: map[string]any{"goal": plan.Goal},
		Status: "success",
	})
	_ = rt.Store.AppendDecision(upCtx, run.ID, domain.Decision{
		Kind: "plan", Summary: fmt.Sprintf("规划 %d 步", len(plan.Steps)), Confidence: 0.9,
		Payload: map[string]any{"goal": plan.Goal},
	})

	// 2) Execute
	for i, ps := range plan.Steps {
		if i >= rc.Policy.MaxSteps {
			return rt.fail(upCtx, rc, fmt.Sprintf("超过最大步数 %d", rc.Policy.MaxSteps))
		}
		rc.State.CurrentStep = i
		switch ps.Action {
		case "output":
			out := rt.resolveArgs(rc, ps.Args)
			rc.Variables[ps.ID] = out
			rc.Variables["output"] = out
			rt.recordStep(upCtx, rc, domain.Step{Kind: domain.StepOutput, Name: ps.ID, Output: out, Status: "success"})

		case "tool":
			res, awaiting, err := rt.execTool(upCtx, rc, ps)
			if err != nil {
				return rt.fail(upCtx, rc, err.Error())
			}
			if awaiting != nil {
				return awaiting, nil
			}
			_ = res

		default:
			return rt.fail(upCtx, rc, "未知的步骤动作: "+ps.Action)
		}
	}

	// 3) 完成
	finished := time.Now()
	run.Status = domain.RunSucceeded
	run.Result = rc.Variables
	run.PromptTokens = rc.tokens
	run.FinishedAt = &finished
	state.Status = domain.RunSucceeded
	state.UpdatedAt = finished
	state.Variables = rc.Variables
	run.State = state
	if err := rt.Store.UpdateRun(upCtx, run); err != nil {
		return nil, err
	}
	_ = rt.Store.AppendEvent(upCtx, domain.Event{Type: "agent.run.succeeded", AggregateType: "run", AggregateID: run.ID, Payload: map[string]any{"agent": spec.ID}})
	return &domain.Result{RunID: run.ID, Status: string(domain.RunSucceeded), Output: rc.Variables}, nil
}

// execTool 执行一个工具步骤（含策略与审批）。
func (rt *Runtime) execTool(ctx context.Context, rc *RunContext, ps domain.PlanStep) (*domain.ToolResult, *domain.Result, error) {
	tl, ok := rt.Tools.Get(ps.Tool)
	if !ok {
		return nil, nil, fmt.Errorf("未注册的工具: %s", ps.Tool)
	}
	if rc.toolCalls >= rc.Policy.MaxToolCalls {
		return nil, nil, fmt.Errorf("超过最大工具调用次数 %d", rc.Policy.MaxToolCalls)
	}

	action := domain.Action{Tool: ps.Tool, Risk: tl.Risk(), Arguments: ps.Args, RunID: rc.Run.ID, AgentID: rc.Spec.ID}
	decision := rt.Policy.Evaluate(action)
	_ = rt.Store.AppendDecision(ctx, rc.Run.ID, domain.Decision{
		Kind: "policy", Summary: decision.Rule, Rationale: decision.Reason, Confidence: 0.8,
		Payload: map[string]any{"tool": ps.Tool, "outcome": string(decision.Outcome)},
	})

	if decision.Outcome == domain.OutcomeDeny {
		return nil, nil, fmt.Errorf("策略拒绝执行 %s：%s", ps.Tool, decision.Reason)
	}

	args := rt.resolveArgs(rc, ps.Args)

	if decision.Outcome == domain.OutcomeRequireApproval {
		apr := &domain.Approval{
			RunID:       rc.Run.ID,
			Kind:        ps.Tool,
			Summary:     fmt.Sprintf("Agent %s 请求执行 %s", rc.Spec.ID, ps.Tool),
			Payload:     map[string]any{"arguments": args},
			Risk:        tl.Risk(),
			Status:      "pending",
			RequestedBy: rc.Spec.ID,
		}
		if err := rt.Store.CreateApproval(ctx, apr); err != nil {
			return nil, nil, err
		}
		if rt.Policy.AutoApprove() {
			now := time.Now()
			apr.Status = "approved"
			apr.DecidedBy = "policy:auto_approve"
			apr.DecidedAt = &now
			_ = rt.Store.UpdateApproval(ctx, apr)
		} else {
			rt.recordStep(ctx, rc, domain.Step{
				Kind: domain.StepApproval, Name: ps.Tool, Input: args, Status: "awaiting_approval",
				Output: map[string]any{"approval_id": apr.ID},
			})
			finished := time.Now()
			rc.Run.Status = domain.RunAwaitingApproval
			rc.Run.FinishedAt = &finished
			rc.State.Status = domain.RunAwaitingApproval
			rc.State.Variables = rc.Variables
			rc.Run.State = rc.State
			_ = rt.Store.UpdateRun(ctx, rc.Run)
			_ = rt.Store.AppendEvent(ctx, domain.Event{Type: "agent.run.awaiting_approval", AggregateID: rc.Run.ID, Payload: map[string]any{"approval_id": apr.ID, "tool": ps.Tool}})
			return nil, &domain.Result{RunID: rc.Run.ID, Status: string(domain.RunAwaitingApproval), Output: map[string]any{"approval_id": apr.ID}}, nil
		}
	}

	start := time.Now()
	result, err := tl.Execute(ctx, args)
	call := domain.ToolCall{
		ID: store.NewID("call"), RunID: rc.Run.ID, Step: rc.seq, Tool: ps.Tool, Arguments: args,
		Status: "success", StartedAt: start, FinishedAt: time.Now(),
	}
	if err != nil {
		call.Status = "failed"
		call.Error = err.Error()
		_ = rt.Store.AppendToolCall(ctx, rc.Run.ID, call)
		return nil, nil, fmt.Errorf("工具 %s 执行失败: %w", ps.Tool, err)
	}
	rc.toolCalls++
	call.Output = result.Output
	call.LatencyMS = int(time.Since(start).Milliseconds())
	_ = rt.Store.AppendToolCall(ctx, rc.Run.ID, call)

	if result.Output != nil {
		rc.Variables[ps.ID] = result.Output
		if t, ok := result.Output["tokens"].(int); ok {
			rc.tokens += t
		}
	}
	rt.recordStep(ctx, rc, domain.Step{
		Kind: domain.StepTool, Name: ps.Tool, Input: args, Output: result.Output, Status: "success",
		LatencyMS: call.LatencyMS,
	})
	return result, nil, nil
}

func (rt *Runtime) plannerFor(spec domain.AgentSpec) Planner {
	if spec.Planner == "llm" {
		return LLMPlanner{Provider: rt.LLM}
	}
	return RulePlanner{}
}

func (rt *Runtime) recordStep(ctx context.Context, rc *RunContext, step domain.Step) {
	step.Seq = rc.seq
	rc.seq++
	if step.Status == "" {
		step.Status = "success"
	}
	step.At = time.Now()
	_ = rt.Store.AppendStep(ctx, rc.Run.ID, step)
}

func (rt *Runtime) recordMessage(ctx context.Context, rc *RunContext, msg llm.Message) {
	dm := domain.Message{Role: string(msg.Role), Content: msg.Content, Name: msg.Name}
	dm.Seq = len(rc.State.Messages)
	rc.State.Messages = append(rc.State.Messages, dm)
	_ = rt.Store.AppendMessage(ctx, rc.Run.ID, dm)
}

func (rt *Runtime) fail(ctx context.Context, rc *RunContext, msg string) (*domain.Result, error) {
	rt.recordStep(ctx, rc, domain.Step{Kind: domain.StepError, Name: "error", Output: map[string]any{"error": msg}, Status: "failed"})
	finished := time.Now()
	rc.Run.Status = domain.RunFailed
	rc.Run.Error = msg
	rc.Run.FinishedAt = &finished
	rc.State.Status = domain.RunFailed
	rc.State.Variables = rc.Variables
	rc.Run.State = rc.State
	_ = rt.Store.UpdateRun(ctx, rc.Run)
	_ = rt.Store.AppendEvent(ctx, domain.Event{Type: "agent.run.failed", AggregateID: rc.Run.ID, Payload: map[string]any{"error": msg}})
	return &domain.Result{RunID: rc.Run.ID, Status: string(domain.RunFailed), Error: msg}, fmt.Errorf("%s", msg)
}

// resolveArgs 解析步骤参数中的 {{path}} 占位符（支持 task.* 与 stepID 输出）。
func (rt *Runtime) resolveArgs(rc *RunContext, args map[string]any) map[string]any {
	if args == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(args))
	for k, v := range args {
		out[k] = rt.resolveValue(rc, v)
	}
	return out
}

func (rt *Runtime) resolveValue(rc *RunContext, v any) any {
	switch x := v.(type) {
	case string:
		s := strings.TrimSpace(x)
		if strings.HasPrefix(s, "{{") && strings.HasSuffix(s, "}}") {
			path := strings.TrimSpace(s[2 : len(s)-2])
			if val, ok := rt.lookup(rc, path); ok {
				return val
			}
		}
		return x
	case map[string]any:
		m := make(map[string]any, len(x))
		for k, vv := range x {
			m[k] = rt.resolveValue(rc, vv)
		}
		return m
	case []any:
		arr := make([]any, len(x))
		for i, vv := range x {
			arr[i] = rt.resolveValue(rc, vv)
		}
		return arr
	default:
		return v
	}
}

func (rt *Runtime) lookup(rc *RunContext, path string) (any, bool) {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return nil, false
	}
	var cur any
	switch parts[0] {
	case "goal":
		return rc.Task.Goal, true
	case "task":
		cur = rc.Task.Input
		parts = parts[1:]
	default:
		v, ok := rc.Variables[parts[0]]
		if !ok {
			return nil, false
		}
		cur = v
		parts = parts[1:]
	}
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

func mergePolicy(p domain.RunPolicy) domain.RunPolicy {
	d := domain.DefaultRunPolicy()
	if p.MaxSteps > 0 {
		d.MaxSteps = p.MaxSteps
	}
	if p.MaxTokens > 0 {
		d.MaxTokens = p.MaxTokens
	}
	if p.Timeout > 0 {
		d.Timeout = p.Timeout
	}
	if p.MaxToolCalls > 0 {
		d.MaxToolCalls = p.MaxToolCalls
	}
	if p.MaxCost > 0 {
		d.MaxCost = p.MaxCost
	}
	d.RequireApproval = p.RequireApproval
	return d
}
