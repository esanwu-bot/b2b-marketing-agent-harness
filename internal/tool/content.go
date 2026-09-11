package tool

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/llm"
)

// NewContentTools 返回内容生成相关 Tool（基于 LLM Provider）。
func NewContentTools(p llm.Provider) []domain.Tool {
	return []domain.Tool{
		&FuncTool{
			ToolName:        "content.generate",
			ToolDescription: "根据选题生成渠道内容（多语言、带品牌口吻）",
			ToolRisk:        domain.RiskRead,
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"topic":    map[string]any{"type": "string"},
					"channel":  map[string]any{"type": "string"},
					"format":   map[string]any{"type": "string"},
					"language": map[string]any{"type": "string"},
				},
				"required": []string{"topic"},
			},
			Fn: func(ctx context.Context, input map[string]any) (*domain.ToolResult, error) {
				topic := str(input["topic"], "嵌入式技术趋势")
				language := str(input["language"], "zh-CN")
				channel := str(input["channel"], "linkedin")
				format := str(input["format"], "post")

				resp, err := p.Chat(ctx, llm.ChatRequest{
					Temperature: 0.6,
					JSON:        true,
					Messages: []llm.Message{
						{Role: llm.RoleSystem, Content: "你是 B2B 半导体技术市场内容专家，输出专业、可信、有洞见的内容。" + llm.MarkerLinkedInPost},
						{Role: llm.RoleUser, Content: "topic=" + topic + "\nchannel=" + channel + "\nformat=" + format + "\nlanguage=" + language},
					},
				})
				if err != nil {
					return nil, err
				}
				title, body, hashtags := parseGenerated(resp.Content, topic)
				return &domain.ToolResult{
					Output: map[string]any{
						"title": title, "body": body, "hashtags": hashtags,
						"channel": channel, "format": format, "language": language,
						"model": resp.Model, "tokens": resp.Usage.TotalTokens,
					},
					Summary: "已生成内容：" + title,
				}, nil
			},
		},
		&FuncTool{
			ToolName:        "content.translate",
			ToolDescription: "翻译内容到目标语言",
			ToolRisk:        domain.RiskRead,
			Fn: func(_ context.Context, input map[string]any) (*domain.ToolResult, error) {
				text := str(input["text"], "")
				target := str(input["target_language"], "en")
				return &domain.ToolResult{
					Output:  map[string]any{"text": "[" + target + "] " + text, "target_language": target},
					Summary: "已翻译为 " + target,
				}, nil
			},
		},
		&FuncTool{
			ToolName:        "content.rewrite",
			ToolDescription: "按品牌口吻/审核意见改写内容",
			ToolRisk:        domain.RiskRead,
			Fn: func(_ context.Context, input map[string]any) (*domain.ToolResult, error) {
				text := str(input["text"], "")
				instruction := str(input["instruction"], "更简洁、更专业")
				return &domain.ToolResult{
					Output:  map[string]any{"text": text + "\n\n（已按「" + instruction + "」改写）"},
					Summary: "已改写内容",
				}, nil
			},
		},
	}
}

func parseGenerated(content, topic string) (title, body string, hashtags []string) {
	content = strings.TrimSpace(content)
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		var m struct {
			Title    string   `json:"title"`
			Body     string   `json:"body"`
			Hashtags []string `json:"hashtags"`
		}
		if json.Unmarshal([]byte(content[start:end+1]), &m) == nil && m.Body != "" {
			return m.Title, m.Body, m.Hashtags
		}
	}
	return "技术观察：" + topic, content, []string{"嵌入式", "半导体"}
}
