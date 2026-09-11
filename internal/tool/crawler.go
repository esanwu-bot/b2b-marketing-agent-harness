package tool

import (
	"context"
	"fmt"
	"time"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/store"
)

// mockArticles 是采集 Tool 的示例数据（半导体行业，简体中文）。
var mockArticles = []map[string]any{
	{
		"title":   "边缘 AI 在工业视觉检测中的落地加速",
		"url":     "https://example.com/news/edge-ai-industrial-vision",
		"source":  "EE Times",
		"summary": "越来越多工业客户开始把推理下沉到设备端，对低功耗与实时性提出更高要求。",
		"content": "边缘 AI 正在从概念走向产线。工业视觉检测对延迟、功耗与可靠性要求苛刻，" +
			"推动 MCU + NPU 的异构方案与 TinyML 工具链快速成熟。厂商开始提供端到端的参考设计。",
	},
	{
		"title":   "10BASE-T1S 车载以太网进入规模量产",
		"url":     "https://example.com/news/10base-t1s-automotive",
		"source":  "TechNews Taiwan",
		"summary": "单对以太网在车身网络中的成本优势显现，软件定义汽车推动其加速采用。",
		"content": "10BASE-T1S 以单对双绞线实现多点总线，替代部分 CAN 场景，降低线束成本与重量。" +
			"随着软件定义汽车架构演进，车载区域控制器对确定性以太网的需求显著上升。",
	},
	{
		"title":   "低功耗 MCU 电源轨监测成为设计刚需",
		"url":     "https://example.com/news/low-power-rail-monitoring",
		"source":  "EDN",
		"summary": "每一微安的掌控能力，正在成为电池供电与能源受限系统的核心竞争力。",
		"content": "现代系统需要持续掌握低电压电源轨的用电情况。集成式电源监测与保护可以简化设计，" +
			"并在异常时及时响应，减少系统级风险。",
	},
	{
		"title":   "智能 PID 风扇控制提升环境能效",
		"url":     "https://example.com/news/smart-pid-fan",
		"source":  "Microchip Blog",
		"summary": "基于模型的风扇控制可显著降低能耗与噪声，同时提升系统稳定性。",
		"content": "智能 PID 风扇控制通过闭环调节，在负载变化时保持温度与噪声的平衡，" +
			"适合数据中心、工业与消费类散热场景。",
	},
	{
		"title":   "汽车以太网安全：从 PHY 到 TSN 的系统视角",
		"url":     "https://example.com/news/automotive-ethernet-security",
		"source":  "Semiconductor Engineering",
		"summary": "功能安全与信息安全需要在物理层到协议栈协同设计。",
		"content": "车载以太网的安全不仅是软件问题，PHY 层的诊断能力、TSN 的时间同步，" +
			"以及安全启动共同构成纵深防御的基础。",
	},
	{
		"title":   "TinyML 工具链降低边缘部署门槛",
		"url":     "https://example.com/news/tinyml-toolchain",
		"source":  "IoT For All",
		"summary": "从模型压缩到自动代码生成，工具链成熟度决定边缘 AI 的落地速度。",
		"content": "TinyML 工具链正在把量化、剪枝与部署自动化，帮助嵌入式团队更快验证想法，" +
			"缩短从原型到量产的距离。",
	},
}

// NewCrawlerTools 返回采集相关 Tool（V1 为 Mock，可替换为 crawl-worker-redis HTTP 适配器）。
func NewCrawlerTools() []domain.Tool {
	return []domain.Tool{
		&FuncTool{
			ToolName:        "crawler.create_job",
			ToolDescription: "在采集系统中创建采集作业（对接 crawl-worker-redis）",
			ToolRisk:        domain.RiskRead,
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"source": map[string]any{"type": "string"},
					"limit":  map[string]any{"type": "integer"},
				},
			},
			Fn: func(_ context.Context, input map[string]any) (*domain.ToolResult, error) {
				source := str(input["source"], "microchip_linkedin")
				return &domain.ToolResult{
					Output:  map[string]any{"job_id": store.NewID("crawl"), "source": source, "queued": intOr(input["limit"], 100)},
					Summary: "已创建采集作业 source=" + source,
				}, nil
			},
		},
		&FuncTool{
			ToolName:        "crawler.get_job",
			ToolDescription: "查询采集作业进度",
			ToolRisk:        domain.RiskRead,
			Fn: func(_ context.Context, input map[string]any) (*domain.ToolResult, error) {
				return &domain.ToolResult{
					Output:  map[string]any{"job_id": str(input["job_id"], ""), "status": "succeeded", "progress": 100},
					Summary: "采集作业已完成",
				}, nil
			},
		},
		&FuncTool{
			ToolName:        "crawler.get_results",
			ToolDescription: "获取采集结果（新闻/竞品内容）",
			ToolRisk:        domain.RiskRead,
			Fn: func(_ context.Context, input map[string]any) (*domain.ToolResult, error) {
				limit := intOr(input["limit"], 5)
				items := mockItems(limit)
				return &domain.ToolResult{
					Output:  map[string]any{"items": items, "count": len(items)},
					Summary: fmt.Sprintf("获取到 %d 条采集结果", len(items)),
				}, nil
			},
		},
		&FuncTool{
			ToolName:        "crawler.search",
			ToolDescription: "按关键词检索行业信息",
			ToolRisk:        domain.RiskRead,
			Fn: func(_ context.Context, input map[string]any) (*domain.ToolResult, error) {
				limit := intOr(input["limit"], 4)
				items := mockItems(limit)
				return &domain.ToolResult{
					Output:  map[string]any{"query": str(input["query"], ""), "items": items, "count": len(items)},
					Summary: fmt.Sprintf("检索到 %d 条信息", len(items)),
				}, nil
			},
		},
	}
}

func mockItems(limit int) []map[string]any {
	if limit <= 0 || limit > len(mockArticles) {
		limit = len(mockArticles)
	}
	out := make([]map[string]any, 0, limit)
	for i := 0; i < limit; i++ {
		item := map[string]any{}
		for k, v := range mockArticles[i] {
			item[k] = v
		}
		item["fetched_at"] = time.Now()
		out = append(out, item)
	}
	return out
}

func str(v any, def string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return def
}

func intOr(v any, def int) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return def
	}
}
