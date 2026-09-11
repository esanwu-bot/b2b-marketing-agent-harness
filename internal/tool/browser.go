package tool

import (
	"context"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
)

// NewBrowserTools 返回浏览器/网页相关 Tool（V1 为 Mock）。
func NewBrowserTools() []domain.Tool {
	return []domain.Tool{
		&FuncTool{
			ToolName:        "browser.search",
			ToolDescription: "通用网页搜索（Mock）",
			ToolRisk:        domain.RiskRead,
			Fn: func(_ context.Context, input map[string]any) (*domain.ToolResult, error) {
				query := str(input["query"], "")
				items := mockItems(3)
				return &domain.ToolResult{
					Output:  map[string]any{"query": query, "results": items, "count": len(items)},
					Summary: "搜索到 " + itoa(len(items)) + " 条结果",
				}, nil
			},
		},
		&FuncTool{
			ToolName:        "browser.fetch",
			ToolDescription: "抓取指定 URL 的正文（Mock）",
			ToolRisk:        domain.RiskRead,
			Fn: func(_ context.Context, input map[string]any) (*domain.ToolResult, error) {
				url := str(input["url"], "")
				return &domain.ToolResult{
					Output:  map[string]any{"url": url, "text": "（mock）页面正文内容。"},
					Summary: "已抓取页面 " + url,
				}, nil
			},
		},
	}
}
