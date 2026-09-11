package tool

import (
	"context"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/store"
)

// NewPipelineTools 返回流水线衔接用的辅助 Tool（读取上一步产物）。
func NewPipelineTools(st store.Store) []domain.Tool {
	return []domain.Tool{
		&FuncTool{
			ToolName:        "strategy.top",
			ToolDescription: "读取优先级最高的已通过选题",
			ToolRisk:        domain.RiskRead,
			Fn: func(ctx context.Context, _ map[string]any) (*domain.ToolResult, error) {
				topics, _ := st.ListTopics(ctx, 50)
				for _, t := range topics {
					if t.Status == "approved" {
						return &domain.ToolResult{
							Output:  map[string]any{"topic": t.Title, "title": t.Title, "topic_id": t.ID, "score": t.Score},
							Summary: "选中选题：" + t.Title,
						}, nil
					}
				}
				// 无已通过选题时退化为最高分候选
				if len(topics) > 0 {
					t := topics[0]
					return &domain.ToolResult{
						Output:  map[string]any{"topic": t.Title, "title": t.Title, "topic_id": t.ID, "score": t.Score},
						Summary: "使用候选选题：" + t.Title,
					}, nil
				}
				return &domain.ToolResult{
					Output:  map[string]any{"topic": "嵌入式技术趋势", "title": "嵌入式技术趋势", "topic_id": ""},
					Summary: "无选题，使用默认主题",
				}, nil
			},
		},
		&FuncTool{
			ToolName:        "content.latest",
			ToolDescription: "读取最新一条内容草稿",
			ToolRisk:        domain.RiskRead,
			Fn: func(ctx context.Context, _ map[string]any) (*domain.ToolResult, error) {
				cs, _ := st.ListContent(ctx, 1)
				if len(cs) > 0 {
					return &domain.ToolResult{
						Output:  map[string]any{"content_id": cs[0].ID, "title": cs[0].Title, "body": cs[0].Body, "status": string(cs[0].Status)},
						Summary: "最新内容：" + cs[0].Title,
					}, nil
				}
				return &domain.ToolResult{Output: map[string]any{"content_id": "", "title": "", "body": ""}, Summary: "暂无内容"}, nil
			},
		},
		&FuncTool{
			ToolName:        "linkedin.latest_publication",
			ToolDescription: "读取最近一次 LinkedIn 发布记录",
			ToolRisk:        domain.RiskRead,
			Fn: func(ctx context.Context, _ map[string]any) (*domain.ToolResult, error) {
				pubs, _ := st.ListPublications(ctx, 1)
				if len(pubs) > 0 {
					return &domain.ToolResult{
						Output:  map[string]any{"publication_id": pubs[0].ID, "external_id": pubs[0].ExternalID, "url": pubs[0].ExternalURL},
						Summary: "最近发布：" + pubs[0].ID,
					}, nil
				}
				return &domain.ToolResult{Output: map[string]any{"publication_id": ""}, Summary: "暂无发布记录"}, nil
			},
		},
	}
}
