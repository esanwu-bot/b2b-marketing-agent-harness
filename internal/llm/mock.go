package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// MockProvider 是离线确定性 Provider：无需任何外部服务即可跑通全链路。
//
// 它通过提示中的标记（Marker*）分派到固定模板，便于测试与演示；
// 真实环境请使用 OpenAI-compatible Provider。
type MockProvider struct {
	Model string
}

// NewMock 创建 mock provider。
func NewMock(model string) *MockProvider {
	if model == "" {
		model = "mock-1"
	}
	return &MockProvider{Model: model}
}

// Name 实现 Provider。
func (m *MockProvider) Name() string { return "mock" }

// Chat 实现 Provider，返回确定性内容。
func (m *MockProvider) Chat(_ context.Context, req ChatRequest) (*ChatResponse, error) {
	prompt := concat(req.Messages)
	marker := detectMarker(prompt)

	var out string
	switch marker {
	case MarkerLinkedInPost:
		out = mockLinkedInPost(prompt)
	case MarkerReview:
		out = `{"verdict":"pass","score":0.92,"issues":[],"suggestions":["可补充一个客户案例以增强说服力"]}`
	case MarkerTrend:
		out = `{"topics":["边缘 AI 在工业检测中的落地","10BASE-T1S 车载以太网","低功耗 MCU 电源轨监测","智能 PID 风扇控制","汽车以太网安全"]}`
	case MarkerStrategy:
		out = `{"selected":["边缘 AI 在工业检测中的落地","10BASE-T1S 车载以太网"],"reason":"技术相关度高、目标受众匹配、竞品讨论度上升"}`
	case MarkerAnalytics:
		out = `{"insights":["技术长文互动率高于产品通告","含数据图表的帖子 CTR 提升明显","周二/周四上午发布效果最佳"],"recommendation":"加大技术深度内容配比，并在周二/周四上午排期"}`
	default:
		out = "（mock）已根据上下文生成结果。"
	}

	usage := Usage{
		PromptTokens:     approxTokens(prompt),
		CompletionTokens: approxTokens(out),
	}
	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	return &ChatResponse{Content: out, Model: m.Model, Usage: usage, Provider: m.Name()}, nil
}

func concat(msgs []Message) string {
	var b strings.Builder
	for _, msg := range msgs {
		b.WriteString(string(msg.Role))
		b.WriteString(": ")
		b.WriteString(msg.Content)
		b.WriteString("\n")
	}
	return b.String()
}

func detectMarker(prompt string) string {
	for _, marker := range []string{MarkerLinkedInPost, MarkerReview, MarkerTrend, MarkerStrategy, MarkerAnalytics} {
		if strings.Contains(prompt, marker) {
			return marker
		}
	}
	return ""
}

func mockLinkedInPost(prompt string) string {
	topic := extractAfter(prompt, "topic=")
	if topic == "" {
		topic = "嵌入式技术趋势"
	}
	body := fmt.Sprintf(`【技术观察】%s

在最近的方案评估中，我们发现越来越多客户开始关注该方向在真实产线中的落地效果，而不仅是纸面参数。

三点值得关注：
1. 系统级视角：单点性能提升需要与电源、时钟、连接方案协同设计。
2. 开发效率：从参考设计到量产的距离，取决于工具链与生态成熟度。
3. 长期可持续：可扩展性与供货稳定性，往往比首发指标更重要。

欢迎在评论区分享你的实践经验。

#嵌入式 #半导体 #边缘AI #工业控制`, topic)

	if strings.Contains(prompt, "json") || strings.Contains(prompt, `"body"`) {
		b, _ := json.Marshal(map[string]any{"title": "技术观察：" + topic, "body": body, "hashtags": []string{"嵌入式", "半导体", "边缘AI"}})
		return string(b)
	}
	return body
}

func extractAfter(s, key string) string {
	i := strings.Index(s, key)
	if i < 0 {
		return ""
	}
	rest := s[i+len(key):]
	if j := strings.IndexAny(rest, "\n\r"); j >= 0 {
		rest = rest[:j]
	}
	return strings.TrimSpace(rest)
}

// approxTokens 粗略估算 token（中英混合按 ~2 字符/token）。
func approxTokens(s string) int {
	n := len([]rune(s))
	if n == 0 {
		return 0
	}
	return n/2 + 1
}
