package domain

import "time"

// WorkflowStepKind 编排步骤类型。
type WorkflowStepKind string

const (
	WfAgent     WorkflowStepKind = "agent"     // 调用某个 Agent
	WfApproval  WorkflowStepKind = "approval"  // 人工审批
	WfChannel   WorkflowStepKind = "channel"   // 渠道发布
	WfDelay     WorkflowStepKind = "delay"     // 等待
	WfCondition WorkflowStepKind = "condition" // 条件分支（V1 简化为表达式=true 继续）
)

// WorkflowStepDef 是 Workflow 定义中的一步。
type WorkflowStepDef struct {
	ID         string           `json:"id"`
	Kind       WorkflowStepKind `json:"kind"`
	Agent      string           `json:"agent,omitempty"`
	Task       string           `json:"task,omitempty"`
	Action     string           `json:"action,omitempty"` // channel action: publish/schedule
	Channel    string           `json:"channel,omitempty"`
	ApprovalOf string           `json:"approval_of,omitempty"`
	Input      map[string]any   `json:"input,omitempty"`
}

// WorkflowDef 是 Workflow 定义（对应 workflows 表 definition 列）。
type WorkflowDef struct {
	Key         string            `json:"key"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Trigger     map[string]any    `json:"trigger,omitempty"`
	Steps       []WorkflowStepDef `json:"steps"`
}

// WorkflowRunStatus 编排运行状态。
type WorkflowRunStatus string

const (
	WfRunning          WorkflowRunStatus = "running"
	WfSucceeded        WorkflowRunStatus = "succeeded"
	WfFailed           WorkflowRunStatus = "failed"
	WfAwaitingApproval WorkflowRunStatus = "awaiting_approval"
	WfCancelled        WorkflowRunStatus = "cancelled"
)

// WorkflowRun 是一次编排执行（对应 workflow_runs 表）。
type WorkflowRun struct {
	ID          string            `json:"id"`
	WorkflowKey string            `json:"workflow_key"`
	Status      WorkflowRunStatus `json:"status"`
	TriggerType string            `json:"trigger_type"`
	Context     map[string]any    `json:"context,omitempty"`
	Steps       []WorkflowStepRun `json:"steps,omitempty"`
	Error       string            `json:"error,omitempty"`
	StartedAt   time.Time         `json:"started_at"`
	FinishedAt  *time.Time        `json:"finished_at,omitempty"`
}

// WorkflowStepRun 是编排运行中的一步。
type WorkflowStepRun struct {
	Seq    int              `json:"seq"`
	Kind   WorkflowStepKind `json:"kind"`
	Ref    string           `json:"ref"`
	Status string           `json:"status"`
	TaskID string           `json:"task_id,omitempty"`
	RunID  string           `json:"run_id,omitempty"`
	Output map[string]any   `json:"output,omitempty"`
	At     time.Time        `json:"at"`
}
