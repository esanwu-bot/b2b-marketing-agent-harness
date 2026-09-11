package domain

import "time"

// TaskStatus 任务生命周期。
type TaskStatus string

const (
	TaskPending          TaskStatus = "pending"
	TaskQueued           TaskStatus = "queued"
	TaskRunning          TaskStatus = "running"
	TaskSucceeded        TaskStatus = "succeeded"
	TaskFailed           TaskStatus = "failed"
	TaskCancelled        TaskStatus = "cancelled"
	TaskAwaitingApproval TaskStatus = "awaiting_approval"
)

// Task 是一个可被 Agent / Workflow 消费的调度单元。
type Task struct {
	ID            string         `json:"id"`
	WorkspaceID   string         `json:"workspace_id,omitempty"`
	WorkflowRunID string         `json:"workflow_run_id,omitempty"`
	AgentID       string         `json:"agent_id,omitempty"`
	Type          string         `json:"type"`
	Goal          string         `json:"goal"`
	Input         map[string]any `json:"input,omitempty"`
	Priority      int            `json:"priority"`
	Status        TaskStatus     `json:"status"`
	ScheduledAt   *time.Time     `json:"scheduled_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}
