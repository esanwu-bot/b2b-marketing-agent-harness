package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIProvider 是 OpenAI-compatible Provider（/chat/completions）。
//
// 兼容 OpenAI / DeepSeek / Qwen / 本地 vLLM / Ollama(OpenAI 模式) 等。
type OpenAIProvider struct {
	BaseURL string
	APIKey  string
	Model   string
	Client  *http.Client
}

// NewOpenAI 创建 OpenAI-compatible provider。
func NewOpenAI(baseURL, apiKey, model string) *OpenAIProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAIProvider{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
		Client:  &http.Client{Timeout: 60 * time.Second},
	}
}

// Name 实现 Provider。
func (p *OpenAIProvider) Name() string { return "openai" }

type oaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

type oaiRequest struct {
	Model          string         `json:"model"`
	Messages       []oaiMessage   `json:"messages"`
	Temperature    float64        `json:"temperature"`
	MaxTokens      int            `json:"max_tokens,omitempty"`
	ResponseFormat map[string]any `json:"response_format,omitempty"`
}

type oaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Model string `json:"model"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Chat 实现 Provider。
func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	model := req.Model
	if model == "" {
		model = p.Model
	}
	msgs := make([]oaiMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, oaiMessage{Role: string(m.Role), Content: m.Content, Name: m.Name})
	}
	body := oaiRequest{Model: model, Messages: msgs, Temperature: req.Temperature, MaxTokens: req.MaxTokens}
	if req.JSON {
		body.ResponseFormat = map[string]any{"type": "json_object"}
	}
	raw, _ := json.Marshal(body)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)
	}
	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("LLM HTTP %d: %s", resp.StatusCode, string(data))
	}
	var out oaiResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("解析 LLM 响应失败: %w", err)
	}
	if out.Error != nil {
		return nil, fmt.Errorf("LLM 错误: %s", out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return nil, fmt.Errorf("LLM 返回空 choices")
	}
	return &ChatResponse{
		Content:  out.Choices[0].Message.Content,
		Model:    out.Model,
		Provider: p.Name(),
		Usage: Usage{
			PromptTokens:     out.Usage.PromptTokens,
			CompletionTokens: out.Usage.CompletionTokens,
			TotalTokens:      out.Usage.TotalTokens,
		},
	}, nil
}
