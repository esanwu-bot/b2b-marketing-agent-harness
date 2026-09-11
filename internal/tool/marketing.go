package tool

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/llm"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/store"
)

// trendRule 是主题聚类的关键词规则。
type trendRule struct {
	Title    string
	Category string
	Keywords []string
}

var trendRules = []trendRule{
	{Title: "边缘 AI 与 TinyML 的工程化落地", Category: "边缘AI", Keywords: []string{"边缘", "AI", "TinyML", "推理", "NPU"}},
	{Title: "车载以太网与软件定义汽车", Category: "车载网络", Keywords: []string{"以太网", "10BASE-T1S", "车载", "TSN", "汽车"}},
	{Title: "低功耗电源与能效管理", Category: "电源管理", Keywords: []string{"电源", "功耗", "能效", "风扇", "PID", "微安"}},
	{Title: "嵌入式安全与功能安全", Category: "安全", Keywords: []string{"安全", "功能安全", "启动", "加密"}},
	{Title: "连接与时钟解决方案", Category: "连接", Keywords: []string{"时钟", "连接", "PHY", "同步"}},
}

// NewMarketingTools 返回内容/选题流水线相关的领域 Tool。
func NewMarketingTools(st store.Store, p llm.Provider) []domain.Tool {
	return []domain.Tool{
		&FuncTool{
			ToolName:        "knowledge.store_batch",
			ToolDescription: "批量把采集结果写入知识库",
			ToolRisk:        domain.RiskWrite,
			Fn: func(ctx context.Context, input map[string]any) (*domain.ToolResult, error) {
				items := toMapSlice(input["items"])
				stored := 0
				for _, it := range items {
					title := str(it["title"], "")
					content := str(it["content"], "")
					if title == "" || content == "" {
						continue
					}
					doc := &domain.Document{
						SourceType: "crawl",
						Title:      title,
						URL:        str(it["url"], ""),
						Summary:    str(it["summary"], ""),
						Content:    content,
						Hash:       hash(title + content),
					}
					if err := st.SaveDocument(ctx, doc); err == nil {
						stored++
					}
				}
				return &domain.ToolResult{Output: map[string]any{"stored": stored}, Summary: "入库 " + itoa(stored) + " 条知识"}, nil
			},
		},
		&FuncTool{
			ToolName:        "trend.analyze",
			ToolDescription: "对知识库做主题聚类，产出选题候选",
			ToolRisk:        domain.RiskWrite,
			Fn: func(ctx context.Context, input map[string]any) (*domain.ToolResult, error) {
				docs, _ := st.ListDocuments(ctx, 500)
				topics := clusterTopics(ctx, st, docs)
				return &domain.ToolResult{
					Output:  map[string]any{"topics": topics, "count": len(topics)},
					Summary: "聚类出 " + itoa(len(topics)) + " 个主题",
				}, nil
			},
		},
		&FuncTool{
			ToolName:        "strategy.select",
			ToolDescription: "对选题打分并选出优先内容方向",
			ToolRisk:        domain.RiskWrite,
			Fn: func(ctx context.Context, input map[string]any) (*domain.ToolResult, error) {
				n := intOr(input["top"], 3)
				topics, _ := st.ListTopics(ctx, 50)
				sort.Slice(topics, func(i, j int) bool { return topics[i].Score > topics[j].Score })
				selected := make([]map[string]any, 0)
				for _, t := range topics {
					if len(selected) >= n {
						break
					}
					if t.Status == "rejected" || t.Status == "used" {
						continue
					}
					t.Status = "approved"
					_ = st.SaveTopic(ctx, t)
					selected = append(selected, map[string]any{"id": t.ID, "title": t.Title, "score": t.Score, "category": t.Category})
				}
				out := map[string]any{"selected": selected, "count": len(selected)}
				if len(selected) > 0 {
					out["top_title"] = selected[0]["title"]
					out["top_topic_id"] = selected[0]["id"]
				}
				return &domain.ToolResult{Output: out, Summary: "选定 " + itoa(len(selected)) + " 个选题"}, nil
			},
		},
		&FuncTool{
			ToolName:        "content.save",
			ToolDescription: "保存内容草稿",
			ToolRisk:        domain.RiskWrite,
			Fn: func(ctx context.Context, input map[string]any) (*domain.ToolResult, error) {
				c := &domain.Content{
					TopicID:     str(input["topic_id"], ""),
					ChannelKind: domain.ChannelKind(str(input["channel"], "linkedin")),
					Format:      str(input["format"], "post"),
					Language:    str(input["language"], "zh-CN"),
					Title:       str(input["title"], ""),
					Body:        str(input["body"], ""),
					Hashtags:    toStringSlice(input["hashtags"]),
					Status:      domain.ContentDraft,
					CreatedBy:   "content_agent",
				}
				if err := st.SaveContent(ctx, c); err != nil {
					return nil, err
				}
				return &domain.ToolResult{Output: map[string]any{"content_id": c.ID, "status": string(c.Status)}, Summary: "已保存草稿：" + c.Title}, nil
			},
		},
		&FuncTool{
			ToolName:        "review.check",
			ToolDescription: "技术事实与品牌口吻审核",
			ToolRisk:        domain.RiskWrite,
			Fn: func(ctx context.Context, input map[string]any) (*domain.ToolResult, error) {
				contentID := str(input["content_id"], "")
				text := str(input["text"], "")
				if text == "" && contentID != "" {
					if c, _ := st.GetContent(ctx, contentID); c != nil {
						text = c.Title + "\n" + c.Body
					}
				}
				resp, err := p.Chat(ctx, llm.ChatRequest{
					JSON: true,
					Messages: []llm.Message{
						{Role: llm.RoleSystem, Content: "你是 B2B 半导体技术内容审核员，检查技术事实与品牌口吻。" + llm.MarkerReview},
						{Role: llm.RoleUser, Content: text},
					},
				})
				verdict, score, issues := parseReview(resp)
				status := domain.ContentApproved
				if verdict != "pass" {
					status = domain.ContentInReview
				}
				if contentID != "" {
					if c, _ := st.GetContent(ctx, contentID); c != nil {
						c.Status = status
						if issues != "" {
							if c.Metadata == nil {
								c.Metadata = map[string]any{}
							}
							c.Metadata["review_issues"] = issues
						}
						_ = st.UpdateContent(ctx, c)
					}
				}
				out := map[string]any{"verdict": verdict, "score": score, "issues": issues, "content_id": contentID}
				if err != nil {
					out["error"] = err.Error()
				}
				return &domain.ToolResult{Output: out, Summary: "审核结论：" + verdict}, nil
			},
		},
		&FuncTool{
			ToolName:        "analytics.insights",
			ToolDescription: "根据内容表现产出优化建议",
			ToolRisk:        domain.RiskRead,
			Fn: func(ctx context.Context, input map[string]any) (*domain.ToolResult, error) {
				pubID := str(input["publication_id"], "")
				resp, err := p.Chat(ctx, llm.ChatRequest{
					JSON: true,
					Messages: []llm.Message{
						{Role: llm.RoleSystem, Content: "你是 B2B 内容分析专家，输出可执行的优化建议。" + llm.MarkerAnalytics},
						{Role: llm.RoleUser, Content: "publication_id=" + pubID},
					},
				})
				insights, recommendation := parseInsights(resp)
				out := map[string]any{"insights": insights, "recommendation": recommendation, "publication_id": pubID}
				if err != nil {
					out["error"] = err.Error()
				}
				return &domain.ToolResult{Output: out, Summary: "已生成优化建议"}, nil
			},
		},
	}
}

func clusterTopics(ctx context.Context, st store.Store, docs []*domain.Document) []map[string]any {
	type agg struct {
		rule  trendRule
		score float64
		refs  []string
	}
	buckets := map[string]*agg{}
	for _, d := range docs {
		hay := strings.ToLower(d.Title + " " + d.Summary + " " + d.Content)
		best := trendRule{}
		bestScore := 0.0
		for _, r := range trendRules {
			s := 0.0
			for _, kw := range r.Keywords {
				if strings.Contains(hay, strings.ToLower(kw)) {
					s++
				}
			}
			if s > bestScore {
				bestScore = s
				best = r
			}
		}
		if bestScore == 0 {
			continue
		}
		b := buckets[best.Category]
		if b == nil {
			b = &agg{rule: best}
			buckets[best.Category] = b
		}
		b.score += bestScore * 10
		b.refs = append(b.refs, d.ID)
	}

	out := make([]map[string]any, 0, len(buckets))
	for _, b := range buckets {
		topic := &domain.Topic{
			Title:      b.rule.Title,
			Category:   b.rule.Category,
			Keywords:   b.rule.Keywords,
			Summary:    "围绕「" + b.rule.Category + "」的行业动态与技术讨论，可作为内容选题。",
			Score:      b.score,
			ClusterKey: b.rule.Category,
			SourceRefs: b.refs,
			Status:     "candidate",
		}
		_ = st.SaveTopic(ctx, topic)
		out = append(out, map[string]any{"id": topic.ID, "title": topic.Title, "score": topic.Score, "category": topic.Category})
	}
	sort.Slice(out, func(i, j int) bool { return toFloat(out[i]["score"]) > toFloat(out[j]["score"]) })
	return out
}

func parseReview(resp *llm.ChatResponse) (verdict string, score float64, issues string) {
	verdict, score, issues = "pass", 0.9, ""
	if resp == nil {
		return
	}
	var m struct {
		Verdict     string   `json:"verdict"`
		Score       float64  `json:"score"`
		Issues      []string `json:"issues"`
		Suggestions []string `json:"suggestions"`
	}
	if json.Unmarshal([]byte(resp.Content), &m) == nil {
		if m.Verdict != "" {
			verdict = m.Verdict
		}
		if m.Score > 0 {
			score = m.Score
		}
		issues = strings.Join(append(m.Issues, m.Suggestions...), "；")
	}
	return
}

func parseInsights(resp *llm.ChatResponse) (insights []string, recommendation string) {
	if resp == nil {
		return nil, ""
	}
	var m struct {
		Insights       []string `json:"insights"`
		Recommendation string   `json:"recommendation"`
	}
	if json.Unmarshal([]byte(resp.Content), &m) == nil {
		return m.Insights, m.Recommendation
	}
	return []string{resp.Content}, ""
}

func toMapSlice(v any) []map[string]any {
	arr, ok := v.([]any)
	if !ok {
		if ms, ok2 := v.([]map[string]any); ok2 {
			return ms
		}
		return nil
	}
	out := make([]map[string]any, 0, len(arr))
	for _, it := range arr {
		if m, ok := it.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func toStringSlice(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		if ss, ok2 := v.([]string); ok2 {
			return ss
		}
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, it := range arr {
		if s, ok := it.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func toFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		return 0
	}
}
