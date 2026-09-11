package domain

import "time"

// RunStatus 一次 Agent Run 的状态。
type RunStatus string

const (
	RunRunning          RunStatus = "running"
	RunSucceeded        RunStatus = "succeeded"
	RunFailed           RunStatus = "failed"
	RunCancelled        RunStatus = "cancelled"
	RunAwaitingApproval RunStatus = "awaiting_approval"
)

// StepKind Run 中一步的类型。
type StepKind string

const (
	StepPlan     StepKind = "plan"
	StepLLM      StepKind = "llm"
	StepTool     StepKind = "tool"
	StepObserve  StepKind = "observe"
	StepReflect  StepKind = "reflect"
	StepApproval StepKind = "approval"
	StepOutput   StepKind = "output"
	StepError    StepKind = "error"
)

// Step 是 Run 中的一步（对应 task_steps 表）。
type Step struct {
	Seq       int            `json:"seq"`
	Kind      StepKind       `json:"kind"`
	Name      string         `json:"name"`
	Input     map[string]any `json:"input,omitempty"`
	Output    map[string]any `json:"output,omitempty"`
	Status    string         `json:"status"`
	LatencyMS int            `json:"latency_ms"`
	At        time.Time      `json:"at"`
}

// Message 是 LLM 对话消息（对应 agent_messages 表）。
type Message struct {
	Seq        int            `json:"seq"`
	Role       string         `json:"role"` // system | user | assistant | tool
	Name       string         `json:"name,omitempty"`
	Content    string         `json:"content"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall     `json:"tool_calls,omitempty"`
	Tokens     int            `json:"tokens"`
	Extra      map[string]any `json:"extra,omitempty"`
}

// ToolCall 记录一次工具调用（对应 tool_executions 表）。
type ToolCall struct {
	ID         string         `json:"id"`
	RunID      string         `json:"run_id"`
	Step       int            `json:"step"`
	Tool       string         `json:"tool"`
	Arguments  map[string]any `json:"arguments"`
	Output     map[string]any `json:"output,omitempty"`
	Status     string         `json:"status"`
	Error      string         `json:"error,omitempty"`
	ApprovalID string         `json:"approval_id,omitempty"`
	LatencyMS  int            `json:"latency_ms"`
	StartedAt  time.Time      `json:"started_at"`
	FinishedAt time.Time      `json:"finished_at"`
}

// Decision 记录一次判断（对应 decisions 表）。
type Decision struct {
	Kind       string         `json:"kind"`
	Summary    string         `json:"summary"`
	Rationale  string         `json:"rationale"`
	Confidence float64        `json:"confidence"`
	Payload    map[string]any `json:"payload,omitempty"`
	At         time.Time      `json:"at"`
}

// AgentState 可持久化的运行状态，支持 Resume。
type AgentState struct {
	RunID       string         `json:"run_id"`
	TaskID      string         `json:"task_id"`
	AgentID     string         `json:"agent_id"`
	Status      RunStatus      `json:"status"`
	CurrentStep int            `json:"current_step"`
	Variables   map[string]any `json:"variables,omitempty"`
	Messages    []Message      `json:"messages,omitempty"`
	ToolCalls   []ToolCall     `json:"tool_calls,omitempty"`
	Decisions   []Decision     `json:"decisions,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// Run 是一次 Agent 执行的聚合（对应 task_runs 表）。
type Run struct {
	ID           string         `json:"id"`
	TaskID       string         `json:"task_id"`
	AgentID      string         `json:"agent_id"`
	AgentVersion int            `json:"agent_version"`
	Status       RunStatus      `json:"status"`
	State        *AgentState    `json:"state,omitempty"`
	Result       map[string]any `json:"result,omitempty"`
	Error        string         `json:"error,omitempty"`
	Steps        []Step         `json:"steps,omitempty"`
	PromptTokens int            `json:"prompt_tokens"`
	CompTokens   int            `json:"comp_tokens"`
	Cost         float64        `json:"cost"`
	StartedAt    time.Time      `json:"started_at"`
	FinishedAt   *time.Time     `json:"finished_at,omitempty"`
}

// LLMCall 记录一次模型调用（对应 llm_calls 表），用于成本与观测。
type LLMCall struct {
	Provider         string         `json:"provider"`
	Model            string         `json:"model"`
	Request          map[string]any `json:"request,omitempty"`
	Response         map[string]any `json:"response,omitempty"`
	PromptTokens     int            `json:"prompt_tokens"`
	CompletionTokens int            `json:"completion_tokens"`
	TotalTokens      int            `json:"total_tokens"`
	Cost             float64        `json:"cost"`
	LatencyMS        int            `json:"latency_ms"`
	Status           string         `json:"status"`
	Error            string         `json:"error,omitempty"`
	At               time.Time      `json:"at"`
}
