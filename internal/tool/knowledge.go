package tool

import (
	"context"
	"crypto/sha1"
	"encoding/hex"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/store"
)

// NewKnowledgeTools 返回知识库相关 Tool。
func NewKnowledgeTools(st store.Store) []domain.Tool {
	return []domain.Tool{
		&FuncTool{
			ToolName:        "knowledge.store",
			ToolDescription: "把采集/调研结果写入知识库",
			ToolRisk:        domain.RiskWrite,
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"title":       map[string]any{"type": "string"},
					"content":     map[string]any{"type": "string"},
					"summary":     map[string]any{"type": "string"},
					"url":         map[string]any{"type": "string"},
					"source_type": map[string]any{"type": "string"},
				},
				"required": []string{"title", "content"},
			},
			Fn: func(ctx context.Context, input map[string]any) (*domain.ToolResult, error) {
				title := str(input["title"], "")
				content := str(input["content"], "")
				doc := &domain.Document{
					SourceType: str(input["source_type"], "manual"),
					Title:      title,
					URL:        str(input["url"], ""),
					Summary:    str(input["summary"], ""),
					Content:    content,
					Hash:       hash(title + content),
				}
				if err := st.SaveDocument(ctx, doc); err != nil {
					return nil, err
				}
				return &domain.ToolResult{
					Output:  map[string]any{"document_id": doc.ID, "hash": doc.Hash},
					Summary: "已入库：" + title,
				}, nil
			},
		},
		&FuncTool{
			ToolName:        "knowledge.search",
			ToolDescription: "从知识库检索相关信息",
			ToolRisk:        domain.RiskRead,
			Fn: func(ctx context.Context, input map[string]any) (*domain.ToolResult, error) {
				query := str(input["query"], "")
				limit := intOr(input["limit"], 8)
				hits, err := st.SearchDocuments(ctx, query, limit)
				if err != nil {
					return nil, err
				}
				return &domain.ToolResult{
					Output:  map[string]any{"query": query, "hits": hits, "count": len(hits)},
					Summary: "检索到 " + itoa(len(hits)) + " 条知识",
				}, nil
			},
		},
	}
}

func hash(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
