package domain

import "time"

// Action 是一个待策略裁决的动作。
type Action struct {
	Tool       string         `json:"tool"` // linkedin.publish / crawler.fetch / content.generate ...
	Risk       string         `json:"risk"`
	Arguments  map[string]any `json:"arguments,omitempty"`
	RunID      string         `json:"run_id,omitempty"`
	AgentID    string         `json:"agent_id,omitempty"`
	Confidence float64        `json:"confidence,omitempty"`
}

// DecisionOutcome 策略裁决结果。
type DecisionOutcome string

const (
	OutcomeAllow           DecisionOutcome = "allow"            // 直接执行
	OutcomeRequireApproval DecisionOutcome = "require_approval" // 人工审批
	OutcomeDeny            DecisionOutcome = "deny"             // 拒绝
)

// PolicyDecision 是策略引擎的判定。
type PolicyDecision struct {
	Outcome DecisionOutcome `json:"outcome"`
	Rule    string          `json:"rule"`
	Reason  string          `json:"reason"`
	Risk    string          `json:"risk"`
}

// Approval 是 Human-in-the-loop 审批单（对应 approvals 表）。
type Approval struct {
	ID            string         `json:"id"`
	WorkspaceID   string         `json:"workspace_id,omitempty"`
	RunID         string         `json:"run_id,omitempty"`
	WorkflowRunID string         `json:"workflow_run_id,omitempty"`
	Kind          string         `json:"kind"` // publish_linkedin / send_message / delete_content
	SubjectType   string         `json:"subject_type,omitempty"`
	SubjectID     string         `json:"subject_id,omitempty"`
	Summary       string         `json:"summary"`
	Payload       map[string]any `json:"payload,omitempty"`
	Risk          string         `json:"risk"`
	Confidence    float64        `json:"confidence"`
	Status        string         `json:"status"` // pending | approved | rejected | expired | cancelled
	RequestedBy   string         `json:"requested_by,omitempty"`
	DecidedBy     string         `json:"decided_by,omitempty"`
	DecisionNote  string         `json:"decision_note,omitempty"`
	RequestedAt   time.Time      `json:"requested_at"`
	DecidedAt     *time.Time     `json:"decided_at,omitempty"`
}

// Event 是领域事件（对应 events 表）。
type Event struct {
	Type          string         `json:"type"`
	AggregateType string         `json:"aggregate_type,omitempty"`
	AggregateID   string         `json:"aggregate_id,omitempty"`
	Payload       map[string]any `json:"payload,omitempty"`
	At            time.Time      `json:"at"`
}
