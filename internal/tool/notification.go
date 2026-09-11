package tool

import (
	"context"
	"fmt"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
)

// NewNotificationTools 返回通知相关 Tool。
func NewNotificationTools() []domain.Tool {
	return []domain.Tool{
		&FuncTool{
			ToolName:        "notification.send",
			ToolDescription: "发送通知（站内/邮箱/Webhook）",
			ToolRisk:        domain.RiskExternal,
			Fn: func(_ context.Context, input map[string]any) (*domain.ToolResult, error) {
				channel := str(input["channel"], "email")
				subject := str(input["subject"], "通知")
				return &domain.ToolResult{
					Output:  map[string]any{"sent": true, "channel": channel, "subject": subject},
					Summary: fmt.Sprintf("已通过 %s 发送通知", channel),
				}, nil
			},
		},
	}
}
