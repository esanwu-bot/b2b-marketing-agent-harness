package tool

import (
	"context"
	"time"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/store"
)

// NewLinkedInTools 返回 LinkedIn 渠道 Tool。
//
// V1 为 Mock 发布（不调用真实 API）；发布类动作需要人工审批。
// 真实 Company Page API 将在 V1.5 接入，接口保持不变。
func NewLinkedInTools(st store.Store) []domain.Tool {
	return []domain.Tool{
		&FuncTool{
			ToolName:        "linkedin.create_post",
			ToolDescription: "在 LinkedIn 公司主页发布一条内容",
			ToolRisk:        domain.RiskExternal,
			Approval:        true,
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"content_id": map[string]any{"type": "string"},
					"text":       map[string]any{"type": "string"},
				},
				"required": []string{"text"},
			},
			Fn: func(ctx context.Context, input map[string]any) (*domain.ToolResult, error) {
				now := time.Now()
				pub := &domain.Publication{
					ContentID:   str(input["content_id"], ""),
					ChannelKind: domain.ChannelLinkedIn,
					Status:      "published",
					ExternalID:  store.NewID("li"),
					ExternalURL: "https://www.linkedin.com/feed/update/mock",
					PublishedAt: &now,
				}
				if err := st.SavePublication(ctx, pub); err != nil {
					return nil, err
				}
				if id := pub.ContentID; id != "" {
					if c, err := st.GetContent(ctx, id); err == nil && c != nil {
						c.Status = domain.ContentPublished
						c.PublishedAt = &now
						c.ExternalURL = pub.ExternalURL
						_ = st.UpdateContent(ctx, c)
					}
				}
				return &domain.ToolResult{
					Output:  map[string]any{"publication_id": pub.ID, "external_id": pub.ExternalID, "url": pub.ExternalURL},
					Summary: "已发布到 LinkedIn（mock）",
				}, nil
			},
		},
		&FuncTool{
			ToolName:        "linkedin.schedule_post",
			ToolDescription: "排期一条 LinkedIn 内容",
			ToolRisk:        domain.RiskExternal,
			Approval:        true,
			Fn: func(ctx context.Context, input map[string]any) (*domain.ToolResult, error) {
				when := time.Now().Add(24 * time.Hour)
				if s := str(input["scheduled_at"], ""); s != "" {
					if t, err := time.Parse(time.RFC3339, s); err == nil {
						when = t
					}
				}
				pub := &domain.Publication{
					ContentID:   str(input["content_id"], ""),
					ChannelKind: domain.ChannelLinkedIn,
					Status:      "scheduled",
					ScheduledAt: &when,
				}
				if err := st.SavePublication(ctx, pub); err != nil {
					return nil, err
				}
				return &domain.ToolResult{
					Output:  map[string]any{"publication_id": pub.ID, "scheduled_at": when.Format(time.RFC3339)},
					Summary: "已排期发布",
				}, nil
			},
		},
		&FuncTool{
			ToolName:        "linkedin.get_analytics",
			ToolDescription: "获取 LinkedIn 内容表现指标",
			ToolRisk:        domain.RiskRead,
			Fn: func(ctx context.Context, input map[string]any) (*domain.ToolResult, error) {
				pubID := str(input["publication_id"], "")
				metric := &domain.Metric{
					PublicationID:  pubID,
					MeasuredAt:     time.Now(),
					Impressions:    1280,
					Reach:          940,
					Likes:          56,
					Comments:       12,
					Shares:         7,
					Clicks:         43,
					EngagementRate: 0.0586,
					CTR:            0.0336,
				}
				if pubID != "" {
					_ = st.SaveMetric(ctx, metric)
				}
				return &domain.ToolResult{
					Output:  map[string]any{"publication_id": pubID, "impressions": metric.Impressions, "engagement_rate": metric.EngagementRate, "ctr": metric.CTR},
					Summary: "已获取 LinkedIn 指标",
				}, nil
			},
		},
	}
}
