package builtin

import "github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"

// DailySemiconductorIntelligence 是参考 Workflow：
// 每日情报 → 选题 → 内容 → 审核 → 审批 → 发布 → 分析。
func DailySemiconductorIntelligence() domain.WorkflowDef {
	return domain.WorkflowDef{
		Key:         "daily_semiconductor_intelligence",
		Name:        "每日半导体情报与内容流水线",
		Description: "Research → Intelligence → Content → Approval → Publishing → Analytics",
		Trigger:     map[string]any{"type": "cron", "schedule": "0 8 * * *", "timezone": "Asia/Shanghai"},
		Steps: []domain.WorkflowStepDef{
			{ID: "research", Kind: domain.WfAgent, Agent: "research", Task: "采集今日半导体行业与竞品情报"},
			{ID: "trend", Kind: domain.WfAgent, Agent: "trend", Task: "对情报做主题聚类，产出选题候选"},
			{ID: "strategy", Kind: domain.WfAgent, Agent: "strategy", Task: "对选题打分并选择优先方向"},
			{ID: "content", Kind: domain.WfAgent, Agent: "content", Task: "为优先选题生成 LinkedIn 内容草稿"},
			{ID: "review", Kind: domain.WfAgent, Agent: "review", Task: "对最新草稿做技术与品牌审核"},
			{ID: "approval", Kind: domain.WfApproval},
			{ID: "publish", Kind: domain.WfChannel, Channel: "linkedin", Action: "publish",
				Input: map[string]any{"content_id": "{{content.output.content_id}}"}},
			{ID: "analytics", Kind: domain.WfAgent, Agent: "analytics", Task: "分析内容表现并给出优化建议"},
		},
	}
}

// Workflows 返回内置 Workflow 列表。
func Workflows() []domain.WorkflowDef {
	return []domain.WorkflowDef{DailySemiconductorIntelligence()}
}
