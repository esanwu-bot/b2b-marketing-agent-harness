// Package builtin 提供半导体营销的参考实现：6 个内置 Agent 规格与示例 Workflow。
//
// 这些规格全部由 domain.AgentSpec 驱动，不新增任何 Runtime 代码；
// 复制此包即可替换为其它行业的 Agent。
package builtin

import "github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"

// Actor 用于演示 loop 占位，保持导出。
// Agents 返回 6 个内置 Agent 规格。
func Agents() []domain.AgentSpec {
	return []domain.AgentSpec{
		researchAgent(),
		trendAgent(),
		strategyAgent(),
		contentAgent(),
		reviewAgent(),
		analyticsAgent(),
	}
}

// ResearchAgent 行业/竞品情报采集。
func researchAgent() domain.AgentSpec {
	return domain.AgentSpec{
		ID:           "research",
		Name:         "研究 Agent",
		Role:         "Research",
		Description:  "采集半导体行业新闻与竞品动态，并沉淀进知识库",
		SystemPrompt: "你是半导体行业研究分析师，擅长从新闻与竞品动态中抽取有价值的情报。",
		Tools:        []string{"crawler.search", "knowledge.store_batch"},
		Planner:      "rule",
		Steps: []domain.PlanStep{
			{ID: "search", Action: "tool", Tool: "crawler.search", Args: map[string]any{"query": "{{goal}}", "limit": 6}, Description: "检索行业信息"},
			{ID: "store", Action: "tool", Tool: "knowledge.store_batch", Args: map[string]any{"items": "{{search.items}}"}, Description: "写入知识库"},
			{ID: "output", Action: "output", Args: map[string]any{"documents": "{{store.stored}}", "found": "{{search.count}}"}},
		},
	}
}

// TrendAgent 趋势发现与聚类。
func trendAgent() domain.AgentSpec {
	return domain.AgentSpec{
		ID:           "trend",
		Name:         "趋势 Agent",
		Role:         "Trend",
		Description:  "对知识库做主题聚类，产出可选题材",
		SystemPrompt: "你是趋势分析师，善于从大量信息中归纳主题与机会。",
		Tools:        []string{"trend.analyze"},
		Planner:      "rule",
		Steps: []domain.PlanStep{
			{ID: "cluster", Action: "tool", Tool: "trend.analyze", Description: "主题聚类"},
			{ID: "output", Action: "output", Args: map[string]any{"topics": "{{cluster.topics}}", "count": "{{cluster.count}}"}},
		},
	}
}

// StrategyAgent 选题与打分。
func strategyAgent() domain.AgentSpec {
	return domain.AgentSpec{
		ID:           "strategy",
		Name:         "策略 Agent",
		Role:         "Strategy",
		Description:  "对选题打分并选出优先内容方向",
		SystemPrompt: "你是内容策略负责人，依据受众匹配度、技术相关度与竞品热度做选题决策。",
		Tools:        []string{"strategy.select"},
		Planner:      "rule",
		Steps: []domain.PlanStep{
			{ID: "select", Action: "tool", Tool: "strategy.select", Args: map[string]any{"top": 3}, Description: "选题打分"},
			{ID: "output", Action: "output", Args: map[string]any{"selected": "{{select.selected}}", "top_title": "{{select.top_title}}", "top_topic_id": "{{select.top_topic_id}}"}},
		},
	}
}

// ContentAgent 多语言内容生成。
func contentAgent() domain.AgentSpec {
	return domain.AgentSpec{
		ID:           "content",
		Name:         "内容 Agent",
		Role:         "Content",
		Description:  "根据选题生成渠道内容草稿",
		SystemPrompt: "你是 B2B 半导体技术市场内容专家，产出专业、可信、有洞见的内容。",
		Tools:        []string{"strategy.top", "content.generate", "content.save"},
		Planner:      "rule",
		Steps: []domain.PlanStep{
			{ID: "pick", Action: "tool", Tool: "strategy.top", Description: "选择选题"},
			{ID: "generate", Action: "tool", Tool: "content.generate", Args: map[string]any{"topic": "{{pick.topic}}", "channel": "linkedin", "format": "post", "language": "zh-CN"}, Description: "生成内容"},
			{ID: "save", Action: "tool", Tool: "content.save", Args: map[string]any{
				"title": "{{generate.title}}", "body": "{{generate.body}}", "hashtags": "{{generate.hashtags}}",
				"topic_id": "{{pick.topic_id}}", "channel": "linkedin", "format": "post", "language": "zh-CN",
			}, Description: "保存草稿"},
			{ID: "output", Action: "output", Args: map[string]any{"content_id": "{{save.content_id}}", "title": "{{generate.title}}"}},
		},
	}
}

// ReviewAgent 技术事实 + 品牌审核。
func reviewAgent() domain.AgentSpec {
	return domain.AgentSpec{
		ID:           "review",
		Name:         "审核 Agent",
		Role:         "Review",
		Description:  "检查技术事实准确性与品牌口吻一致性",
		SystemPrompt: "你是 B2B 半导体技术内容审核员，严格检查技术事实与品牌口吻。",
		Tools:        []string{"content.latest", "review.check"},
		Planner:      "rule",
		Steps: []domain.PlanStep{
			{ID: "latest", Action: "tool", Tool: "content.latest", Description: "读取最新草稿"},
			{ID: "check", Action: "tool", Tool: "review.check", Args: map[string]any{"content_id": "{{latest.content_id}}"}, Description: "审核"},
			{ID: "output", Action: "output", Args: map[string]any{"verdict": "{{check.verdict}}", "score": "{{check.score}}", "content_id": "{{latest.content_id}}"}},
		},
	}
}

// AnalyticsAgent 数据回流与优化建议。
func analyticsAgent() domain.AgentSpec {
	return domain.AgentSpec{
		ID:           "analytics",
		Name:         "分析 Agent",
		Role:         "Analytics",
		Description:  "拉取内容表现并产出优化建议",
		SystemPrompt: "你是内容数据分析师，善于把表现数据转化为可执行的优化建议。",
		Tools:        []string{"linkedin.latest_publication", "linkedin.get_analytics", "analytics.insights"},
		Planner:      "rule",
		Steps: []domain.PlanStep{
			{ID: "latestPub", Action: "tool", Tool: "linkedin.latest_publication", Description: "读取最近发布"},
			{ID: "metrics", Action: "tool", Tool: "linkedin.get_analytics", Args: map[string]any{"publication_id": "{{latestPub.publication_id}}"}, Description: "拉取指标"},
			{ID: "insights", Action: "tool", Tool: "analytics.insights", Args: map[string]any{"publication_id": "{{latestPub.publication_id}}"}, Description: "生成建议"},
			{ID: "output", Action: "output", Args: map[string]any{"insights": "{{insights.insights}}", "recommendation": "{{insights.recommendation}}"}},
		},
	}
}
