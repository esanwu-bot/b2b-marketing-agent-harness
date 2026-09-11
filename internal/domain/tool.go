package domain

import "context"

// Tool 是 Agent 可调用的原子能力。
type Tool interface {
	Name() string
	Description() string
	InputSchema() map[string]any
	// Risk：read（只读）/ write（写库/本地）/ external（对外发布、发送）
	Risk() string
	RequiresApproval() bool
	Execute(ctx context.Context, input map[string]any) (*ToolResult, error)
}

// ToolResult 是工具执行结果。
type ToolResult struct {
	Output map[string]any `json:"output"`
	// Summary 供 Trace / UI 展示的简短说明
	Summary string `json:"summary,omitempty"`
}

// ToolInfo 是 Tool 的可序列化元数据（用于注册表 / 落库 / API）。
type ToolInfo struct {
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	InputSchema      map[string]any `json:"input_schema,omitempty"`
	Risk             string         `json:"risk"`
	RequiresApproval bool           `json:"requires_approval"`
}

// Info 从 Tool 提取元数据。
func Info(t Tool) ToolInfo {
	return ToolInfo{
		Name:             t.Name(),
		Description:      t.Description(),
		InputSchema:      t.InputSchema(),
		Risk:             t.Risk(),
		RequiresApproval: t.RequiresApproval(),
	}
}

// 工具风险等级。
const (
	RiskRead     = "read"
	RiskWrite    = "write"
	RiskExternal = "external"
)
