// Package app 负责 Harness 的全部组件装配（Store / LLM / Tool / Policy / Runtime / Workflow）。
package app

import (
	"context"
	"log/slog"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/agent"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/builtin"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/config"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/llm"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/observability"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/policy"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/store"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/tool"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/workflow"
)

// App 汇总一次运行所需的全部组件。
type App struct {
	Cfg     *config.Config
	Log     *slog.Logger
	Store   store.Store
	LLM     llm.Provider
	Tools   *tool.Registry
	Policy  *policy.Engine
	Runtime *agent.Runtime
	Agents  *agent.Registry
	Flow    *workflow.Engine
}

// New 装配 Harness（默认使用内存存储 + mock LLM，可离线运行）。
func New(ctx context.Context, cfg *config.Config) (*App, error) {
	log := observability.New()
	st := store.NewMemory()

	provider := buildLLM(cfg)
	tools := tool.BuildRegistry(st, provider)
	pol := policy.New(cfg.Policy.AutoApprove)
	rt := agent.NewRuntime(st, tools, provider, pol, log)

	agents := agent.NewRegistry()
	for _, spec := range builtin.Agents() {
		agents.Register(agent.NewSpecAgent(spec, rt))
		if err := st.SaveAgent(ctx, spec); err != nil {
			return nil, err
		}
	}
	for _, info := range tools.Infos() {
		_ = st.SaveTool(ctx, info)
	}
	for _, wf := range builtin.Workflows() {
		if err := st.SaveWorkflow(ctx, wf); err != nil {
			return nil, err
		}
	}
	seedMemories(ctx, st)

	flow := workflow.NewEngine(st, agents, rt, tools, pol, log)
	return &App{Cfg: cfg, Log: log, Store: st, LLM: provider, Tools: tools, Policy: pol, Runtime: rt, Agents: agents, Flow: flow}, nil
}

func buildLLM(cfg *config.Config) llm.Provider {
	switch cfg.LLM.Provider {
	case "openai", "openai-compatible":
		return llm.NewOpenAI(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.Model)
	default:
		return llm.NewMock(cfg.LLM.Model)
	}
}

func seedMemories(ctx context.Context, st store.Store) {
	items := []domain.Memory{
		{Scope: domain.MemoryWorkspace, ScopeID: "default", Kind: "preference", Key: "brand_tone",
			Content: "品牌口吻：专业、可信、以工程价值为导向；避免夸张营销词，强调可落地的技术收益。", Importance: 0.95},
		{Scope: domain.MemoryWorkspace, ScopeID: "default", Kind: "fact", Key: "target_market",
			Content: "目标市场：中国及大中华区嵌入式/工业/汽车电子工程师与采购决策者。", Importance: 0.8},
	}
	for i := range items {
		_ = st.SaveMemory(ctx, &items[i])
	}
}

// RunWorkflow 执行一个 Workflow（供 API / CLI / Scheduler 调用）。
func (a *App) RunWorkflow(ctx context.Context, key, trigger string) (*domain.WorkflowRun, error) {
	return a.Flow.Run(ctx, key, trigger)
}

// ResumeWorkflow 恢复一个 awaiting_approval 状态的 WorkflowRun，继续执行剩余步骤。
func (a *App) ResumeWorkflow(ctx context.Context, runID string) (*domain.WorkflowRun, error) {
	return a.Flow.Resume(ctx, runID)
}

// CancelWorkflow 终止一个 awaiting_approval / running 状态的 WorkflowRun。
func (a *App) CancelWorkflow(ctx context.Context, runID, reason string) (*domain.WorkflowRun, error) {
	return a.Flow.Cancel(ctx, runID, reason)
}