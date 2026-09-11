// Package llm 定义 LLM Provider 抽象与实现（mock / openai-compatible）。
//
// 契约与具体模型解耦：Agent 只依赖 Provider，可替换 OpenAI / Anthropic / Gemini /
// 本地模型 / OpenAI-compatible 服务。
package llm

import "context"

// Role 对话角色。
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message 一条对话消息。
type Message struct {
	Role       Role   `json:"role"`
	Content    string `json:"content"`
	Name       string `json:"name,omitempty"`
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// ChatRequest 一次补全请求。
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
	// JSON 为 true 时要求模型输出 JSON（结构化输出）。
	JSON bool `json:"json"`
}

// Usage token 用量。
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatResponse 补全结果。
type ChatResponse struct {
	Content string `json:"content"`
	Model   string `json:"model"`
	Usage   Usage  `json:"usage"`
	// Provider 便于观测与成本统计。
	Provider string `json:"provider"`
}

// Provider 是模型提供方契约。
type Provider interface {
	Name() string
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

// 供 tools / agents 使用的提示标记：mock provider 通过标记返回确定性结果。
const (
	MarkerLinkedInPost = "LINKEDIN_POST"
	MarkerReview       = "REVIEW"
	MarkerTrend        = "TREND"
	MarkerStrategy     = "STRATEGY"
	MarkerAnalytics    = "ANALYTICS"
)
