package domain

import "time"

// MemoryScope 记忆作用域。
type MemoryScope string

const (
	MemoryAgent     MemoryScope = "agent"
	MemoryUser      MemoryScope = "user"
	MemoryWorkspace MemoryScope = "workspace"
	MemoryTask      MemoryScope = "task"
)

// Memory 记录“过去发生了什么”（Agent / 用户 / 工作区 / 任务的记忆）。
type Memory struct {
	ID          string         `json:"id"`
	WorkspaceID string         `json:"workspace_id,omitempty"`
	Scope       MemoryScope    `json:"scope"`
	ScopeID     string         `json:"scope_id"`
	Kind        string         `json:"kind"` // preference | fact | summary | episode
	Key         string         `json:"key"`
	Content     string         `json:"content"`
	Importance  float64        `json:"importance"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
